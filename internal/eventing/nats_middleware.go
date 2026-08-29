package eventing

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/ThatCatDev/ep/v2/drivers"
	epNats "github.com/ThatCatDev/ep/v2/drivers/nats"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/ThatCatDev/ep/v2/middleware"
	"github.com/weeb-vip/anime-sync/config"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
)

// The NATS equivalents of the middlewares in handler_anime_kafka.go.
//
// They exist separately rather than being shared because the Kafka ones are
// typed on *kafka.Message and, more importantly, read the wrong field. ep's
// ExtractEvent marshals the whole driver message into RawData, so the key is
// whatever that struct calls its payload: "Value" for Kafka, "Data" for NATS.
//
// Reusing the Kafka transform against NATS would compile, run, log
// "Value key not found in RawData" once per message and leave Payload empty --
// a consumer that processes nothing while looking perfectly healthy.

type NatsLoggerMiddleware[M any] struct{}

func NewNatsLoggerMiddleware[M any]() *NatsLoggerMiddleware[M] {
	return &NatsLoggerMiddleware[M]{}
}

func (f *NatsLoggerMiddleware[M]) Process(ctx context.Context, data event.Event[*epNats.Message, M], next middleware.Handler[*epNats.Message, M]) (*event.Event[*epNats.Message, M], error) {
	log := logger.FromCtx(ctx)

	result, err := next(ctx, data)
	if err != nil {
		log.Error("Error processing message", zap.Error(err))

		return result, err
	}

	jsonPayload, marshalErr := json.Marshal(result.Payload)
	if marshalErr != nil {
		log.Error("Error marshalling processed message", zap.Error(marshalErr))
	} else {
		log.Info("Successfully processed message", zap.String("value", string(jsonPayload)))
	}

	return result, nil
}

type NatsTransformMiddleware[M any] struct{}

func NewNatsTransformMiddleware[M any]() *NatsTransformMiddleware[M] {
	return &NatsTransformMiddleware[M]{}
}

// Process unwraps a Debezium change event into the handler's payload type.
//
// The shape is identical to Kafka's because Debezium Server is configured with
// the same converter settings the Connect connector used --
// debezium.format.value=json with schemas.enable=true -- so the body is still
// {"schema": ..., "payload": ...}. Only the field it arrives under differs.
func (f *NatsTransformMiddleware[M]) Process(ctx context.Context, data event.Event[*epNats.Message, M], next middleware.Handler[*epNats.Message, M]) (*event.Event[*epNats.Message, M], error) {
	log := logger.FromCtx(ctx)

	// "Data", not "Value": epNats.Message names its body Data, and RawData is
	// that struct marshalled to JSON.
	rawValue, exists := data.RawData["Data"]
	if !exists {
		log.Warn("Data key not found in RawData")

		return next(ctx, data)
	}

	valueStr, ok := rawValue.(string)
	if !ok {
		log.Warn("Data in RawData is not a string")

		return next(ctx, data)
	}

	// []byte marshals to base64 in JSON, so the body arrives encoded even
	// though nothing encoded it on purpose.
	decodedBytes, err := base64.StdEncoding.DecodeString(valueStr)
	if err != nil {
		log.Error("Failed to decode base64 value", zap.Error(err))

		return nil, err
	}

	var debeziumMessage struct {
		Schema  interface{} `json:"schema"`
		Payload M           `json:"payload"`
	}
	if err := json.Unmarshal(decodedBytes, &debeziumMessage); err != nil {
		log.Error("Failed to unmarshal decoded payload", zap.Error(err))

		return nil, err
	}

	data.Payload = debeziumMessage.Payload

	return next(ctx, data)
}

// newNatsProducerDriver builds a driver for publishing, deliberately with no
// StreamName set.
//
// ep uses whichever stream its driver is configured with for Produce as well as
// for consuming. The consumer driver here is bound to Debezium's stream, so
// publishing image-sync through it asks JetStream to add image-sync to the CDC
// stream -- which it refuses once image-sync has a stream of its own:
//
//	failed to bind subject "image-sync" to stream "ANIMEDB":
//	subjects overlap with an existing stream
//
// The processor returns that error, so the message is never acked and
// redelivers forever. Four of the six CDC consumers were wedged this way.
//
// With StreamName empty the driver derives a stream from the subject, which is
// what these subjects want: image-sync and algolia-sync are produced by this
// service, not by Debezium, and belong in their own streams under their own
// retention rather than inside a change-data-capture stream.
func newNatsProducerDriver(cfg config.Config) drivers.Driver[*epNats.Message] {
	return epNats.NewNatsDriver(&epNats.Config{
		URL: cfg.NatsConfig.URL,
	})
}

const (
	maxRetries     = 3
	retryHeaderKey = "retry"
)

// newNatsRetryDriver builds the driver for the retry and dead-letter subjects.
//
// It names the CDC stream, unlike the producer driver, which leaves it empty.
//
// The retry and dead-letter subjects are derived from the CDC subject, so they
// sit under the same wildcard: anime-db.public.anime-retry still matches
// anime-db.>. JetStream refuses two streams whose subjects overlap, so leaving
// StreamName empty made the driver try to create its own and fail outright --
// "subjects overlap with an existing stream" -- taking every CDC consumer down
// with it. These subjects have to live in Debezium's stream because they are
// already inside its namespace.
//
// It still carries its own durable name: ep takes that from driver-level config
// rather than per-subject, so reusing the CDC driver's name would have the
// retry consumer reconfigure the main consumer out from under it.
func newNatsRetryDriver(cfg config.Config) drivers.Driver[*epNats.Message] {
	offset := cfg.NatsConfig.Offset

	return epNats.NewNatsDriver(&epNats.Config{
		URL:                     cfg.NatsConfig.URL,
		ConsumerGroupName:       cfg.NatsConfig.ConsumerGroupName + "-retry",
		StreamName:              cfg.NatsConfig.StreamName,
		ConsumerAutoOffsetReset: &offset,
	})
}

// natsProducer mirrors kafkaProducer: a Produce closure bound to one subject.
func natsProducer(ctx context.Context, driver drivers.Driver[*epNats.Message], subject string) func(ctx context.Context, value []byte) error {
	return func(ctx context.Context, value []byte) error {
		log := logger.FromCtx(ctx)
		log.Info("Producing message to NATS", zap.String("subject", subject), zap.String("value", string(value)))
		// Data, not Value -- epNats.Message names its body differently.
		if err := driver.Produce(ctx, subject, &epNats.Message{Data: value}); err != nil {
			log.Error("Failed to produce message", zap.String("subject", subject), zap.Error(err))

			return err
		}

		return nil
	}
}
