package episode_processor

import (
	"context"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/weeb-vip/anime-sync/internal/db"
	anime_episode "github.com/weeb-vip/anime-sync/internal/db/repositories/anime_episode"
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
type EpisodeProcessor[DM any] interface {
	Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error)
}

type EpisodeProcessorImpl[DM any] struct {
	Repository anime_episode.AnimeEpisodeRepositoryImpl
	Options    Options
}

func NewAnimeProcessor[DM any](opt Options, db *db.DB) EpisodeProcessor[DM] {
	return &EpisodeProcessorImpl[DM]{
		Repository: anime_episode.NewAnimeRepository(db),
		Options:    opt,
	}
}

func (p *EpisodeProcessorImpl[DM]) Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error) {
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
			// An episode whose anime has not arrived yet. Debezium gives no
			// ordering guarantee across tables, so this is expected occasionally
			// rather than exceptional. Retrying cannot fix it -- the anime arrives
			// on its own topic, not by re-attempting the episode -- so the event is
			// dropped and the consumer moves on. The row returns with the next
			// update to it, or with the next snapshot.
			if db.IsForeignKeyViolation(err) {
				log.Warn("skipping episode whose anime is not present",
					zap.Stringp("anime_id", newAnime.AnimeID),
					zap.String("episode_id", newAnime.ID))
				return data, nil
			}
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

func (p *EpisodeProcessorImpl[DM]) parseToEntity(ctx context.Context, data Schema) (*anime_episode.AnimeEpisode, error) {
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
