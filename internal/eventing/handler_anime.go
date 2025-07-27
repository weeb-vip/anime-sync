package eventing

import (
	"context"
	"fmt"
	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/producer"
	"github.com/weeb-vip/anime-sync/internal/services/consumer"
	"github.com/weeb-vip/anime-sync/internal/services/processor"
	"github.com/weeb-vip/anime-sync/internal/services/pulsar_anime_postgres_processor"
	"go.uber.org/zap"
)

func EventingAnime() error {
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
		ConsumerAutoOffsetReset:  nil,
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

	posgresProcessorOptions := pulsar_anime_postgres_processor.Options{
		NoErrorOnDelete: true,
	}

	algoliaProducer := producer.NewProducer[pulsar_anime_postgres_processor.ProducerPayload](ctx, cfg.PulsarConfig, cfg.PulsarConfig.ProducerAlgoliaTopic)
	imageProducer := producer.NewProducer[pulsar_anime_postgres_processor.ImagePayload](ctx, cfg.PulsarConfig, cfg.PulsarConfig.ProducerImageTopic)

	postgresProcessor := pulsar_anime_postgres_processor.NewPulsarAnimePostgresProcessor(posgresProcessorOptions, database, algoliaProducer, imageProducer, kafkaProducer(ctx, driver, cfg.KafkaConfig.Topic))

	messageProcessor := processor.NewProcessor[pulsar_anime_postgres_processor.Payload]()

	animeConsumer := consumer.NewConsumer[pulsar_anime_postgres_processor.Payload](ctx, cfg.PulsarConfig)

	log.Info("Starting anime eventing")
	err := animeConsumer.Receive(ctx, func(ctx context.Context, msg pulsar.Message) error {
		return messageProcessor.Process(ctx, string(msg.Payload()), postgresProcessor.Process)
	})
	if err != nil {
		log.Error(fmt.Sprintf("Error receiving message: %v", err))
		return err
	}

	return err
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
		if err := driver.Produce(ctx, topic, message); err != nil {
			log.Error("Failed to produce message", zap.String("topic", topic), zap.Error(err))
			return err
		}
		return nil
	}
}
