package eventing

import (
	"context"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/ThatCatDev/ep/v2/middlewares/kafka/backoffretry"
	"github.com/ThatCatDev/ep/v2/processor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/services/anime_season_processor"
	"go.uber.org/zap"
)

func EventingAnimeSeasonKafka() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	ctx = addFFToCtx(ctx, cfg)

	kafkaConfig := &epKafka.KafkaConfig{
		ConsumerGroupName:        cfg.KafkaConfig.ConsumerGroupName,
		BootstrapServers:         cfg.KafkaConfig.BootstrapServers,
		SaslMechanism:            nil,
		SecurityProtocol:         nil,
		Username:                 nil,
		Password:                 nil,
		ConsumerSessionTimeoutMs: nil,
		ConsumerAutoOffsetReset:  &cfg.KafkaConfig.Offset,
		ClientID:                 nil,
		Debug:                    nil,
	}

	driver := epKafka.NewKafkaDriver(kafkaConfig)
	defer func(driver drivers.Driver[*kafka.Message]) {
		err := driver.Close()
		if err != nil {
			log.Error("Error closing Kafka driver", zap.String("error", err.Error()))
		} else {
			log.Info("Kafka driver closed successfully")
		}
	}(driver)

	database := db.NewDB(cfg.DBConfig)

	postgresProcessorOptions := anime_season_processor.Options{
		NoErrorOnDelete: true,
	}

	postgresProcessor := anime_season_processor.NewAnimeSeasonProcessor(postgresProcessorOptions, database, kafkaProducer(ctx, driver, cfg.KafkaConfig.AlgoliaTopic))

	processorInstance := processor.NewProcessor[*kafka.Message, anime_season_processor.Payload](driver, cfg.KafkaConfig.Topic, postgresProcessor.Process)

	log.Info("initializing backoff retry middleware", zap.String("topic", cfg.KafkaConfig.Topic))
	backoffRetryInstance := backoffretry.NewBackoffRetry[anime_season_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		RetryQueue: cfg.KafkaConfig.Topic + "-retry",
	})

	log.Info("Starting Kafka processor", zap.String("topic", cfg.KafkaConfig.Topic))

	err := processorInstance.
		AddMiddleware(NewLoggerMiddleware[*kafka.Message, anime_season_processor.Payload]().Process).
		AddMiddleware(NewTransformMiddleware[*kafka.Message, anime_season_processor.Payload]().Process).
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil {
		log.Error("Error consuming messages", zap.String("error", err.Error()))
		return err
	}

	return nil
}
