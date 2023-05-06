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

		consumer.Ack(msg)
		return nil
	}
}

//
//type Payload struct {
//	Before interface{} `json:"before"`
//	After  struct {
//		Id            string      `json:"id"`
//		TitleEn       string      `json:"title_en"`
//		TitleJp       string      `json:"title_jp"`
//		TitleRomaji   interface{} `json:"title_romaji"`
//		TitleKanji    interface{} `json:"title_kanji"`
//		Type          string      `json:"type"`
//		ImageUrl      interface{} `json:"image_url"`
//		Synopsis      string      `json:"synopsis"`
//		Episodes      int         `json:"episodes"`
//		Status        string      `json:"status"`
//		Duration      string      `json:"duration"`
//		Broadcast     interface{} `json:"broadcast"`
//		Source        string      `json:"source"`
//		CreatedAt     int64       `json:"created_at"`
//		UpdatedAt     int64       `json:"updated_at"`
//		Rating        string      `json:"rating"`
//		StartDate     interface{} `json:"start_date"`
//		EndDate       interface{} `json:"end_date"`
//		TitleSynonyms string      `json:"title_synonyms"`
//		Genres        interface{} `json:"genres"`
//		Licensors     string      `json:"licensors"`
//		Studios       string      `json:"studios"`
//	} `json:"after"`
//	Source struct {
//		Version   string      `json:"version"`
//		Connector string      `json:"connector"`
//		Name      string      `json:"name"`
//		TsMs      int64       `json:"ts_ms"`
//		Snapshot  string      `json:"snapshot"`
//		Db        string      `json:"db"`
//		Sequence  string      `json:"sequence"`
//		Schema    string      `json:"schema"`
//		Table     string      `json:"table"`
//		TxId      int         `json:"txId"`
//		Lsn       int         `json:"lsn"`
//		Xmin      interface{} `json:"xmin"`
//	} `json:"source"`
//	Op          string      `json:"op"`
//	TsMs        int64       `json:"ts_ms"`
//	Transaction interface{} `json:"transaction"`
//}
