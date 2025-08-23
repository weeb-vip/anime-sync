package eventing

import (
	"context"
	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/ThatCatDev/ep/v2/middlewares/kafka/backoffretry"
	"github.com/ThatCatDev/ep/v2/processor"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/producer"
	"github.com/weeb-vip/anime-sync/internal/services/anime_processor"
	"github.com/weeb-vip/anime-sync/internal/services/pulsar_anime_postgres_processor"
	"go.uber.org/zap"
)

func EventingAnimeKafka() error {
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

	posgresProcessorOptions := anime_processor.Options{
		NoErrorOnDelete: true,
	}

	algoliaProducer := producer.NewProducer[pulsar_anime_postgres_processor.ProducerPayload](ctx, cfg.PulsarConfig, cfg.PulsarConfig.ProducerAlgoliaTopic)

	postgresProcessor := anime_processor.NewAnimeProcessor(posgresProcessorOptions, database, algoliaProducer, kafkaProducer(ctx, driver, cfg.KafkaConfig.ProducerTopic))

	processorInstance := processor.NewProcessor[*kafka.Message, anime_processor.Payload](driver, cfg.KafkaConfig.Topic, postgresProcessor.Process)

	log.Info("initializing backoff retry middleware", zap.String("topic", cfg.KafkaConfig.Topic))
	backoffRetryInstance := backoffretry.NewBackoffRetry[anime_processor.Payload](driver, backoffretry.Config{
		MaxRetries: 3,
		HeaderKey:  "retry",
		RetryQueue: cfg.KafkaConfig.Topic + "-retry",
	})

	log.Info("Starting Kafka processor", zap.String("topic", cfg.KafkaConfig.Topic))
	err := processorInstance.
		AddMiddleware(backoffRetryInstance.Process).
		Run(ctx)

	if err != nil && ctx.Err() == nil { // Ignore error if caused by context cancellation
		log.Error("Error consuming messages", zap.String("error", err.Error()))
		return err
	}

	return nil
}

func addFFToCtx(ctx context.Context, cfg config.Config) context.Context {
	ffClient := flagsmith.NewClient(cfg.FFConfig.APIKey,
		flagsmith.WithBaseURL(cfg.FFConfig.BaseURL),
		flagsmith.WithContext(ctx))

	// create value in context to store flagsmith client
	return context.WithValue(ctx, internal.FFClient{}, ffClient)
}

func kafkaProducer(ctx context.Context, driver drivers.Driver[*kafka.Message], topic string) func(ctx context.Context, message *kafka.Message) error {
	return func(ctx context.Context, message *kafka.Message) error {
		log := logger.FromCtx(ctx)
		log.Info("Producing message to Kafka", zap.String("topic", topic), zap.String("key", string(message.Key)), zap.String("value", string(message.Value)))
		if err := driver.Produce(ctx, topic, message); err != nil {
			log.Error("Failed to produce message", zap.String("topic", topic), zap.Error(err))
			return err
		}
		return nil
	}
}
