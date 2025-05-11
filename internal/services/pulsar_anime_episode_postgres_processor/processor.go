package pulsar_anime_postgres_processor

import (
	"context"
	"github.com/weeb-vip/anime-sync/internal/db"
	anime_episode "github.com/weeb-vip/anime-sync/internal/db/repositories/anime_episode"
	"log"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeEpisodePostgresProcessorImpl interface {
	Process(ctx context.Context, data Payload) error
	parseToEntity(ctx context.Context, data Schema) (*anime_episode.AnimeEpisode, error)
}

type PulsarAnimeEpisodePostgresProcessor struct {
	Repository anime_episode.AnimeEpisodeRepositoryImpl
	Options    Options
}

func NewPulsarAnimeEpisodePostgresProcessor(opt Options, db *db.DB) PulsarAnimeEpisodePostgresProcessorImpl {
	return &PulsarAnimeEpisodePostgresProcessor{
		Repository: anime_episode.NewAnimeRepository(db),
		Options:    opt,
	}
}

func (p *PulsarAnimeEpisodePostgresProcessor) Process(ctx context.Context, data Payload) error {

	if data.Before == nil && data.After != nil {
		// add to db
		newAnime, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		err = p.Repository.Upsert(newAnime)
		if err != nil {
			return err
		}
	}

	if data.After == nil && data.Before != nil {
		// delete from db
		oldAnime, err := p.parseToEntity(ctx, *data.Before)
		if err != nil {
			return err
		}

		err = p.Repository.Delete(oldAnime)
		if err != nil {
			if p.Options.NoErrorOnDelete {
				log.Println("WARN: error deleting from db: ", err)
				return nil
			} else {
				return err
			}
		}
		return nil

	}

	if data.Before != nil && data.After != nil {
		// update db
		newAnime, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		err = p.Repository.Upsert(newAnime)
		if err != nil {
			return err
		}
	}

	if data.Before != nil && data.After == nil {
		log.Println("WARN: data.After is nil, skipping update")
	}

	return nil

}

func (p *PulsarAnimeEpisodePostgresProcessor) parseToEntity(ctx context.Context, data Schema) (*anime_episode.AnimeEpisode, error) {
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
