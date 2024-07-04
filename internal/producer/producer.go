package producer

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
)

type Producer[T any] interface {
	Send(ctx context.Context, data string) error
}

type ProducerImpl[T any] struct {
	client pulsar.Client
	config config.PulsarConfig
}

func NewProducer[T any](ctx context.Context, cfg config.PulsarConfig) Producer[T] {
	log := logger.FromCtx(ctx)
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.URL,
	})

	if err != nil {
		log.Fatal("Error creating pulsar client: ", zap.String("error", err.Error()))
		return nil
	}

	return &ProducerImpl[T]{
		config: cfg,
		client: client,
	}
}

func (p *ProducerImpl[T]) Send(ctx context.Context, data string) error {
	log := logger.FromCtx(ctx)
	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic: p.config.ProducerTopic,
	})

	if err != nil {
		log.Fatal("Error creating pulsar producer: ", zap.String("error", err.Error()))
		return err
	}

	defer producer.Close()

	msg := pulsar.ProducerMessage{
		Payload: []byte(data),
	}

	_, err = producer.Send(ctx, &msg)
	if err != nil {
		log.Fatal("Error sending message: ", zap.String("error", err.Error()))
		return err
	}

	return nil
}
