package eventing

import (
	"context"

	"github.com/ThatCatDev/ep/v2/drivers"
	epNats "github.com/ThatCatDev/ep/v2/drivers/nats"
	"github.com/ThatCatDev/ep/v2/middlewares/nats/backoffretry"
	"github.com/ThatCatDev/ep/v2/processor"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/services/anime_season_processor"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// EventingAnimeSeasonNats is the NATS counterpart of EventingAnimeSeasonKafka.
//
// It is a separate entry point rather than a flag on the existing one so that
// production keeps running the Kafka command untouched while staging moves
// over. The two share the processor, the database layer and the retry policy;
// only the transport differs.
func EventingAnimeSeasonNats() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	natsConfig := &epNats.Config{
		URL:               cfg.NatsConfig.URL,
		ConsumerGroupName: cfg.NatsConfig.ConsumerGroupName,
		// Bind to Debezium's stream instead of deriving one from the subject.
		// JetStream rejects overlapping subjects, so creating a stream for
		// anime-db-staging.public.anime_seasons would collide with the stream Debezium
		// declares over anime-db-staging.>.
		StreamName:              cfg.NatsConfig.StreamName,
		ConsumerAutoOffsetReset: &cfg.NatsConfig.Offset,
	}

	driver := epNats.NewNatsDriver(natsConfig)
	defer func(driver drivers.Driver[*epNats.Message]) {
		if err := driver.Close(); err != nil {
			log.Error("Error closing NATS driver", zap.String("error", err.Error()))
		} else {
			log.Info("NATS driver closed successfully")
		}
	}(driver)

	// Publishing goes through its own driver so that these subjects get their
	// own streams instead of being forced into Debezium's. See
	// newNatsProducerDriver.
	producerDriver := newNatsProducerDriver(cfg)
	defer func(d drivers.Driver[*epNats.Message]) {
		if err := d.Close(); err != nil {
			log.Error("Error closing NATS producer driver", zap.String("error", err.Error()))
		}
	}(producerDriver)

	database := db.NewDB(cfg.DBConfig)

	processorOptions := anime_season_processor.Options{
		NoErrorOnDelete: true,
	}

	postgresProcessor := anime_season_processor.NewAnimeSeasonProcessor[*epNats.Message](processorOptions, database, natsProducer(ctx, producerDriver, cfg.NatsConfig.AlgoliaSubject))

	processorInstance := processor.NewProcessor[*epNats.Message, anime_season_processor.Payload](driver, cfg.NatsConfig.Subject, postgresProcessor.Process)

	retrySubject := cfg.NatsConfig.Subject + "-retry"
	dlqSubject := cfg.NatsConfig.Subject + "-dlq"

	// Retries go through their own driver, not the CDC one. The CDC driver
	// names Debezium's stream, and ep resolves the target stream from the
	// driver, so producing a retry through it puts the message back inside the
	// change-data-capture stream -- where it matches the wildcard, nothing
	// filters on it, and it quietly expires with the change events.
	retryDriver := newNatsRetryDriver(cfg)
	defer func(d drivers.Driver[*epNats.Message]) {
		if err := d.Close(); err != nil {
			log.Error("Error closing NATS retry driver", zap.String("error", err.Error()))
		}
	}(retryDriver)

	processorInstance = processorInstance.
		AddMiddleware(NewNatsLoggerMiddleware[anime_season_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[anime_season_processor.Payload]().Process).
		AddMiddleware(backoffretry.NewBackoffRetry[anime_season_processor.Payload](retryDriver, backoffretry.Config{
			MaxRetries: maxRetries,
			HeaderKey:  retryHeaderKey,
			RetryQueue: retrySubject,
		}).Process)

	// The retry consumer, in this same process rather than a second deployment.
	// It repeats the transform middleware because backoffretry republishes the
	// raw driver payload -- still wrapped in its Debezium envelope -- not the
	// decoded one.
	//
	// Exhausted retries go to a dead-letter subject rather than back onto the
	// retry subject: ep acks and drops a message once the counter reaches
	// MaxRetries, so cycling here would make a permanently failing message
	// disappear with no record.
	retryProcessorInstance := processor.NewProcessor[*epNats.Message, anime_season_processor.Payload](retryDriver, retrySubject, postgresProcessor.Process).
		AddMiddleware(NewNatsLoggerMiddleware[anime_season_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[anime_season_processor.Payload]().Process).
		AddMiddleware(backoffretry.NewBackoffRetry[anime_season_processor.Payload](retryDriver, backoffretry.Config{
			MaxRetries: maxRetries,
			HeaderKey:  retryHeaderKey,
			RetryQueue: dlqSubject,
		}).Process)

	log.Info("Starting NATS processors",
		zap.String("subject", cfg.NatsConfig.Subject),
		zap.String("retry_subject", retrySubject),
		zap.String("dlq_subject", dlqSubject))

	// One consumer returning must stop the other: Consume blocks until its
	// iterator is stopped, so without cancelling here a dead main consumer
	// would leave the process alive and apparently healthy.
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return processorInstance.Run(groupCtx) })
	group.Go(func() error { return retryProcessorInstance.Run(groupCtx) })

	if err := group.Wait(); err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))

		return err
	}

	return nil
}
