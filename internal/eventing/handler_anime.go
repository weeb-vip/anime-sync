package eventing

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/producer"
	"github.com/weeb-vip/anime-sync/internal/services/consumer"
	"github.com/weeb-vip/anime-sync/internal/services/processor"
	"github.com/weeb-vip/anime-sync/internal/services/pulsar_anime_postgres_processor"
)

func EventingAnime() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()

	database := db.NewDB(cfg.DBConfig)

	posgresProcessorOptions := pulsar_anime_postgres_processor.Options{
		NoErrorOnDelete: true,
	}

	animeProducer := producer.NewProducer[pulsar_anime_postgres_processor.Schema](ctx, cfg.PulsarConfig)

	postgresProcessor := pulsar_anime_postgres_processor.NewPulsarAnimePostgresProcessor(posgresProcessorOptions, database, animeProducer)

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
