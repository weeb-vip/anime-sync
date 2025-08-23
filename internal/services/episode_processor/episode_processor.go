package episode_processor

import (
	"context"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/internal/db"
	anime_episode "github.com/weeb-vip/anime-sync/internal/db/repositories/anime_episode"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type EpisodeProcessor interface {
	Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error)
}

type EpisodeProcessorImpl struct {
	Repository anime_episode.AnimeEpisodeRepositoryImpl
	Options    Options
}

func NewAnimeProcessor(opt Options, db *db.DB) EpisodeProcessor {
	return &EpisodeProcessorImpl{
		Repository: anime_episode.NewAnimeRepository(db),
		Options:    opt,
	}
}

func (p *EpisodeProcessorImpl) Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	if payload.Before == nil && payload.After != nil {
		// add to db
		newAnime, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		err = p.Repository.Upsert(newAnime)
		if err != nil {
			return data, err
		}
	}

	if payload.After == nil && payload.Before != nil {
		// delete from db
		oldAnime, err := p.parseToEntity(ctx, *payload.Before)
		if err != nil {
			return data, err
		}

		err = p.Repository.Delete(oldAnime)
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
		newAnime, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		err = p.Repository.Upsert(newAnime)
		if err != nil {
			return data, err
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: data.After is nil, skipping update")
	}

	return data, nil

}

func (p *EpisodeProcessorImpl) parseToEntity(ctx context.Context, data Schema) (*anime_episode.AnimeEpisode, error) {
	var newEpisode anime_episode.AnimeEpisode

	newEpisode.ID = data.Id
	newEpisode.AnimeID = data.AnimeId
	newEpisode.TitleEn = data.TitleEn
	newEpisode.TitleJp = data.TitleJp
	newEpisode.Aired = data.Aired
	newEpisode.Episode = data.Episode
	newEpisode.Synopsis = data.Synopsis

	newEpisode.CreatedAt = time.Now()
	newEpisode.UpdatedAt = time.Now()

	return &newEpisode, nil
}
