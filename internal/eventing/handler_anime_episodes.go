package eventing

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/services/processor"
	pulsar_anime_postgres_processor "github.com/weeb-vip/anime-sync/internal/services/pulsar_anime_episode_postgres_processor"
	"go.uber.org/zap"

	"time"
)

func EventingAnimeEpisode() error {
	cfg := config.LoadConfigOrPanic()
	ctx := context.Background()
	log := logger.Get()

	database := db.NewDB(cfg.DBConfig)

	posgresProcessorOptions := pulsar_anime_postgres_processor.Options{
		NoErrorOnDelete: true,
	}
	postgresProcessor := pulsar_anime_postgres_processor.NewPulsarAnimeEpisodePostgresProcessor(posgresProcessorOptions, database)

	messageProcessor := processor.NewProcessor[pulsar_anime_postgres_processor.Payload]()

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.PulsarConfig.URL,
	})

	if err != nil {
		log.Fatal("Error creating pulsar client: ", zap.String("error", err.Error()))
		return err
	}

	defer client.Close()

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            cfg.PulsarConfig.Topic,
		SubscriptionName: cfg.PulsarConfig.SubscribtionName,
		Type:             pulsar.Shared,
	})

	defer consumer.Close()

	// create channel to receive messages

	for {
		msg, err := consumer.Receive(ctx)
		if err != nil {
			log.Fatal("Error receiving message: ", zap.String("error", err.Error()))
		}

		log.Info("Received message", zap.String("msgId", msg.ID().String()))

		err = messageProcessor.Process(ctx, string(msg.Payload()), postgresProcessor.Process)
		if err != nil {
			log.Warn("error processing message: ", zap.String("error", err.Error()))
			continue
		}
		consumer.Ack(msg)
		time.Sleep(50 * time.Millisecond)
	}

}
