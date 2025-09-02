package anime_season_processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/internal"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime_season"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type AnimeSeasonProcessor interface {
	Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error)
}

type AnimeSeasonProcessorImpl struct {
	Repository      anime_season.AnimeSeasonRepositoryImpl
	Options         Options
	AlgoliaProducer func(ctx context.Context, message *kafka.Message) error
}

func NewAnimeSeasonProcessor(opt Options, db *db.DB, algoliaProducer func(ctx context.Context, message *kafka.Message) error) AnimeSeasonProcessor {
	return &AnimeSeasonProcessorImpl{
		Repository:      anime_season.NewAnimeSeasonRepository(db),
		Options:         opt,
		AlgoliaProducer: algoliaProducer,
	}
}

func (p *AnimeSeasonProcessorImpl) Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	log.Info("Getting flagsmith client from context")
	flagsmithClientInterface := ctx.Value(internal.FFClient{})
	if flagsmithClientInterface == nil {
		log.Error("Flagsmith client not found in context")
		return data, fmt.Errorf("flagsmith client not found in context")
	}

	flagsmithClient, ok := flagsmithClientInterface.(interface {
		GetEnvironmentFlags() (flagsmith.Flags, error)
	})
	if !ok {
		log.Error("Flagsmith client has wrong type")
		return data, fmt.Errorf("flagsmith client has wrong type")
	}

	log.Info("Getting environment flags from flagsmith client")
	flags, err := flagsmithClient.GetEnvironmentFlags()
	if err != nil {
		log.Error("Failed to get environment flags", zap.Error(err))
		return data, fmt.Errorf("failed to get environment flags: %w", err)
	}

	log.Info("Checking if feature 'enable_kafka' is enabled")
	isEnabled, _ := flags.IsFeatureEnabled("enable_kafka")
	log.Info("Feature 'enable_kafka' is enabled", zap.Bool("isEnabled", isEnabled))

	log.Debug("Payload", zap.Any("payload", payload))

	if payload.Before == nil && payload.After != nil {
		// add to db
		newAnimeSeason, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		err = p.Repository.Upsert(newAnimeSeason)
		if err != nil {
			return data, err
		}

		jsonAnimeSeason, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   payload.After,
		})
		if err != nil {
			log.Error("Error marshalling payload", zap.Error(err))
			return data, err
		}

		err = p.AlgoliaProducer(ctx, &kafka.Message{
			Value: jsonAnimeSeason,
		})
		if err != nil {
			log.Error("Error sending message to algolia producer", zap.Error(err))
			return data, err
		}
	}

	if payload.After == nil && payload.Before != nil {
		// delete from db
		oldAnimeSeason, err := p.parseToEntity(ctx, *payload.Before)
		if err != nil {
			return data, err
		}

		err = p.Repository.Delete(oldAnimeSeason)
		if err != nil {
			if p.Options.NoErrorOnDelete {
				log.Warn("WARN: error deleting from db: ", zap.Error(err))
				return data, nil
			} else {
				return data, err
			}
		}
		return data, nil
	}

	if payload.Before != nil && payload.After != nil {
		// update db
		newAnimeSeason, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}

		err = p.Repository.Upsert(newAnimeSeason)
		if err != nil {
			return data, err
		}

		jsonAnimeSeason, err := json.Marshal(ProducerPayload{
			Action: UpdateAction,
			Data:   payload.After,
		})
		if err != nil {
			return data, err
		}

		err = p.AlgoliaProducer(ctx, &kafka.Message{
			Value: jsonAnimeSeason,
		})
		if err != nil {
			return data, err
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: payload.After is nil, skipping update")
	}

	return data, nil
}

func (p *AnimeSeasonProcessorImpl) parseToEntity(ctx context.Context, data Schema) (*anime_season.AnimeSeason, error) {
	var newAnimeSeason anime_season.AnimeSeason

	newAnimeSeason.ID = data.ID
	newAnimeSeason.Season = data.Season
	newAnimeSeason.Status = anime_season.AnimeSeasonStatus(data.Status)
	newAnimeSeason.EpisodeCount = data.EpisodeCount
	newAnimeSeason.Notes = data.Notes
	newAnimeSeason.AnimeID = data.AnimeID
	newAnimeSeason.CreatedAt = time.Now()
	newAnimeSeason.UpdatedAt = time.Now()

	return &newAnimeSeason, nil
}
