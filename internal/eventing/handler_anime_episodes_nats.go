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
	"github.com/weeb-vip/anime-sync/internal/services/episode_processor"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// EventingAnimeEpisodeNats is the NATS counterpart of EventingAnimeEpisodeKafka.
//
// It is a separate entry point rather than a flag on the existing one so that
// production keeps running the Kafka command untouched while staging moves
// over. The two share the processor, the database layer and the retry policy;
// only the transport differs.
func EventingAnimeEpisodeNats() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	natsConfig := &epNats.Config{
		URL:               cfg.NatsConfig.URL,
		ConsumerGroupName: cfg.NatsConfig.ConsumerGroupName,
		// Bind to Debezium's stream instead of deriving one from the subject.
		// JetStream rejects overlapping subjects, so creating a stream for
		// anime-db-staging.public.anime_episodes would collide with the stream Debezium
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

	database := db.NewDB(cfg.DBConfig)

	processorOptions := episode_processor.Options{
		NoErrorOnDelete: true,
	}

	episodeProcessorInstance := episode_processor.NewAnimeProcessor[*epNats.Message](processorOptions, database)

	processorInstance := processor.NewProcessor[*epNats.Message, episode_processor.Payload](driver, cfg.NatsConfig.Subject, episodeProcessorInstance.Process)

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
		AddMiddleware(NewNatsLoggerMiddleware[episode_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[episode_processor.Payload]().Process).
		AddMiddleware(backoffretry.NewBackoffRetry[episode_processor.Payload](retryDriver, backoffretry.Config{
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
	retryProcessorInstance := processor.NewProcessor[*epNats.Message, episode_processor.Payload](retryDriver, retrySubject, episodeProcessorInstance.Process).
		AddMiddleware(NewNatsLoggerMiddleware[episode_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[episode_processor.Payload]().Process).
		AddMiddleware(backoffretry.NewBackoffRetry[episode_processor.Payload](retryDriver, backoffretry.Config{
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
