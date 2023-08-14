package eventing

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/services/processor"
	"github.com/weeb-vip/anime-sync/internal/services/pulsar_anime_postgres_processor"

	"log"
	"time"
)

func EventingAnime() error {
	cfg := config.LoadConfigOrPanic()

	database := db.NewDB(cfg.DBConfig)

	posgresProcessorOptions := pulsar_anime_postgres_processor.Options{
		NoErrorOnDelete: true,
	}
	postgresProcessor := pulsar_anime_postgres_processor.NewPulsarAnimePostgresProcessor(posgresProcessorOptions, database)

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
		Topic:            "public/default/myanimelist.public.anime",
		SubscriptionName: "my-sub",
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
