package eventing

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/services/processor"
	pulsar_anime_postgres_processor "github.com/weeb-vip/anime-sync/internal/services/pulsar_anime_episode_postgres_processor"

	"log"
	"time"
)

func EventingAnimeEpisode() error {
	cfg := config.LoadConfigOrPanic()

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
		log.Fatal(err)
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
		msg, err := consumer.Receive(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Received message msgId: %#v -- content: '%s'\n",
			msg.ID(), string(msg.Payload()))

		err = messageProcessor.Process(string(msg.Payload()), postgresProcessor.Process)
		if err != nil {
			log.Println("error processing message: ", err)
			continue
		}
		consumer.Ack(msg)
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
