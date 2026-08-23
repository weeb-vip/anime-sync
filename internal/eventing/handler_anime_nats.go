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
	"github.com/weeb-vip/anime-sync/internal/services/anime_processor"
	"go.uber.org/zap"
)

// EventingAnimeNats is the NATS counterpart of EventingAnimeKafka.
//
// It is a separate entry point rather than a flag on the existing one so that
// production keeps running the Kafka command untouched while staging moves
// over. The two share the processor, the database layer and the retry policy;
// only the transport differs.
func EventingAnimeNats() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	natsConfig := &epNats.Config{
		URL:               cfg.NatsConfig.URL,
		ConsumerGroupName: cfg.NatsConfig.ConsumerGroupName,
		// Bind to Debezium's stream instead of deriving one from the subject.
		// JetStream rejects overlapping subjects, so creating a stream for
		// anime-db-staging.public.anime would collide with the stream Debezium
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

	processorOptions := anime_processor.Options{
		NoErrorOnDelete: true,
	}

	postgresProcessor := anime_processor.NewAnimeProcessor[*epNats.Message](processorOptions, database, natsProducer(ctx, driver, cfg.NatsConfig.AlgoliaSubject), natsProducer(ctx, driver, cfg.NatsConfig.ProducerSubject))

	processorInstance := processor.NewProcessor[*epNats.Message, anime_processor.Payload](driver, cfg.NatsConfig.Subject, postgresProcessor.Process)

	log.Info("initializing backoff retry middleware", zap.String("subject", cfg.NatsConfig.Subject))
	backoffRetryInstance := backoffretry.NewBackoffRetry[anime_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		// A subject rather than a topic, but the same idea: exhausted messages
		// land somewhere inspectable instead of being dropped or redelivered
		// forever.
		RetryQueue: cfg.NatsConfig.Subject + "-retry",
	})

	log.Info("Starting NATS processor", zap.String("subject", cfg.NatsConfig.Subject))

	err := processorInstance.
		AddMiddleware(NewNatsLoggerMiddleware[anime_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[anime_processor.Payload]().Process).
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))

		return err
	}

	return nil
}
