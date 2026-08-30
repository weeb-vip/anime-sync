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
	"github.com/weeb-vip/anime-sync/internal/services/work_processor"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// EventingWorkNats copies the scraper's `work` table into the read store.
//
// NATS only; there is no Kafka twin. The Kafka handlers exist because this
// service predates the move and production ran on Kafka while staging changed
// over. Nothing has ever carried works over Kafka, so a second entry point
// would be dead code written for a transport already retired.
//
// It publishes on two subjects: covers to image-sync, because the scraper
// stores MyAnimeList's own URL and a page whose art is the point should not
// depend on an external host to render, and the record itself to the works
// search index, which is separate from the anime one.
func EventingWorkNats() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	natsConfig := &epNats.Config{
		URL:               cfg.NatsConfig.URL,
		ConsumerGroupName: cfg.NatsConfig.ConsumerGroupName,
		// Bind to Debezium's stream instead of deriving one from the subject.
		// JetStream rejects overlapping subjects, so creating a stream for
		// anime-db.public.work would collide with the stream Debezium declares
		// over anime-db.>.
		StreamName:              cfg.NatsConfig.StreamName,
		ConsumerAutoOffsetReset: &cfg.NatsConfig.Offset,
	}

	driver := epNats.NewNatsDriver(natsConfig)
	defer func(d drivers.Driver[*epNats.Message]) {
		if err := d.Close(); err != nil {
			log.Error("Error closing NATS driver", zap.String("error", err.Error()))
		}
	}(driver)

	// Same stream again, for the same reason: the retry and dead-letter
	// subjects are derived from the CDC subject, so anime-db.public.work-retry
	// is still inside anime-db.> and cannot have a stream of its own.
	retryDriver := newNatsRetryDriver(cfg)
	defer func(d drivers.Driver[*epNats.Message]) {
		if err := d.Close(); err != nil {
			log.Error("Error closing NATS retry driver", zap.String("error", err.Error()))
		}
	}(retryDriver)

	// Publishing goes through its own driver with no StreamName. image-sync is
	// outside anime-db.> and owns its own stream, so producing through the CDC
	// driver would ask JetStream to file the message in Debezium's stream --
	// which it refuses, leaving the message unacked and redelivering forever.
	producerDriver := newNatsProducerDriver(cfg)
	defer func(d drivers.Driver[*epNats.Message]) {
		if err := d.Close(); err != nil {
			log.Error("Error closing NATS producer driver", zap.String("error", err.Error()))
		}
	}(producerDriver)

	database := db.NewDB(cfg.DBConfig)

	workProcessor := work_processor.NewWorkProcessor[*epNats.Message](
		work_processor.Options{NoErrorOnDelete: true},
		database,
		natsProducer(ctx, producerDriver, cfg.NatsConfig.ProducerSubject),
		natsProducer(ctx, producerDriver, cfg.NatsConfig.AlgoliaWorkSubject),
	)

	retrySubject := cfg.NatsConfig.Subject + "-retry"
	dlqSubject := cfg.NatsConfig.Subject + "-dlq"

	processorInstance := processor.NewProcessor[*epNats.Message, work_processor.Payload](driver, cfg.NatsConfig.Subject, workProcessor.Process).
		AddMiddleware(NewNatsLoggerMiddleware[work_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[work_processor.Payload]().Process).
		AddMiddleware(backoffretry.NewBackoffRetry[work_processor.Payload](retryDriver, backoffretry.Config{
			MaxRetries: maxRetries,
			HeaderKey:  retryHeaderKey,
			RetryQueue: retrySubject,
		}).Process)

	// The retry consumer, in this same process rather than a second deployment.
	// It repeats the transform middleware because backoffretry republishes the
	// raw driver payload -- still wrapped in its Debezium envelope.
	retryProcessorInstance := processor.NewProcessor[*epNats.Message, work_processor.Payload](retryDriver, retrySubject, workProcessor.Process).
		AddMiddleware(NewNatsLoggerMiddleware[work_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[work_processor.Payload]().Process).
		AddMiddleware(backoffretry.NewBackoffRetry[work_processor.Payload](retryDriver, backoffretry.Config{
			MaxRetries: maxRetries,
			HeaderKey:  retryHeaderKey,
			RetryQueue: dlqSubject,
		}).Process)

	log.Info("Starting NATS processors",
		zap.String("subject", cfg.NatsConfig.Subject),
		zap.String("retry_subject", retrySubject),
		zap.String("dlq_subject", dlqSubject))

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return processorInstance.Run(groupCtx) })
	group.Go(func() error { return retryProcessorInstance.Run(groupCtx) })

	if err := group.Wait(); err != nil && ctx.Err() == nil {
		log.Error("Error consuming messages", zap.String("error", err.Error()))

		return err
	}

	return nil
}
