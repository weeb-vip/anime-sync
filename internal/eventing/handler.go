package eventing

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/anime-sync/config"
	"log"
)

func Eventing() error {
	cfg := config.LoadConfigOrPanic()
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.PulsarConfig.URL,
	})

	defer client.Close()

	reader, err := client.CreateReader(pulsar.ReaderOptions{
		Topic:          "public/default/myanimelist.public.anime",
		StartMessageID: pulsar.EarliestMessageID(),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	for reader.HasNext() {
		msg, err := reader.Next(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Received message msgId: %#v -- content: '%s'\n",
			msg.ID(), string(msg.Payload()))
	}

	return nil
}
