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

	log.Info("initializing backoff retry middleware", zap.String("subject", cfg.NatsConfig.Subject))
	backoffRetryInstance := backoffretry.NewBackoffRetry[episode_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		// A subject rather than a topic, but the same idea: exhausted messages
		// land somewhere inspectable instead of being dropped or redelivered
		// forever.
		RetryQueue: cfg.NatsConfig.Subject + "-retry",
	})

	log.Info("Starting NATS processor", zap.String("subject", cfg.NatsConfig.Subject))

	err := processorInstance.
		AddMiddleware(NewNatsLoggerMiddleware[episode_processor.Payload]().Process).
		AddMiddleware(NewNatsTransformMiddleware[episode_processor.Payload]().Process).
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))

		return err
	}

	return nil
}
