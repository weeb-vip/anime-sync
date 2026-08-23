package anime_season_processor

import (
	"context"
	"encoding/json"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime_season"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

// The driver message type is a parameter because the processor never looks at
// it. Nothing here reads DriverMessage, RawData or Headers -- only Payload,
// which the transform middleware has already filled in. Hard-coding
// *kafka.Message meant this could not be reused over NATS despite none of the
// logic being Kafka-specific.
//
// Producers take the encoded value rather than a driver message for the same
// reason: every call site only ever set Value, so building the transport's
// message belongs in the handler that knows which transport it is.
type AnimeSeasonProcessor[DM any] interface {
	Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error)
}

type AnimeSeasonProcessorImpl[DM any] struct {
	Repository      anime_season.AnimeSeasonRepositoryImpl
	Options         Options
	AlgoliaProducer func(ctx context.Context, value []byte) error
}

func NewAnimeSeasonProcessor[DM any](opt Options, db *db.DB, algoliaProducer func(ctx context.Context, value []byte) error) AnimeSeasonProcessor[DM] {
	return &AnimeSeasonProcessorImpl[DM]{
		Repository:      anime_season.NewAnimeSeasonRepository(db),
		Options:         opt,
		AlgoliaProducer: algoliaProducer,
	}
}

func (p *AnimeSeasonProcessorImpl[DM]) Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

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

		err = p.AlgoliaProducer(ctx, jsonAnimeSeason)
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

		err = p.AlgoliaProducer(ctx, jsonAnimeSeason)
		if err != nil {
			return data, err
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: payload.After is nil, skipping update")
	}

	return data, nil
}

func (p *AnimeSeasonProcessorImpl[DM]) parseToEntity(ctx context.Context, data Schema) (*anime_season.AnimeSeason, error) {
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
