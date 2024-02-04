package pulsar_anime_postgres_processor

import (
	"database/sql"
	"github.com/weeb-vip/anime-sync/internal/db"
	anime_episode "github.com/weeb-vip/anime-sync/internal/db/repositories/anime_episode"

	"log"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimeEpisodePostgresProcessorImpl interface {
	Process(data Payload) error
	parseToEntity(data Schema) (*anime_episode.AnimeEpisode, error)
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

func (p *PulsarAnimeEpisodePostgresProcessor) Process(data Payload) error {

	if data.Before == nil && data.After != nil {
		// add to db
		newAnime, err := p.parseToEntity(*data.After)
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
		oldAnime, err := p.parseToEntity(*data.Before)
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
		newAnime, err := p.parseToEntity(*data.After)
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

func (p *PulsarAnimeEpisodePostgresProcessor) parseToEntity(data Schema) (*anime_episode.AnimeEpisode, error) {
	var newEpisode anime_episode.AnimeEpisode

	episodeAird := sql.NullTime{
		Time:  time.Time{},
		Valid: false,
	}

	if data.Aired != nil {
		aired, err := time.Parse(time.RFC3339, *data.Aired)
		if err != nil {
			return nil, err
		}
		episodeAird = sql.NullTime{
			Time:  aired,
			Valid: true,
		}
	}

	newEpisode.ID = data.Id
	newEpisode.AnimeID = data.AnimeId
	newEpisode.TitleEn = data.TitleEn
	newEpisode.TitleJp = data.TitleJp
	newEpisode.Aired = episodeAird
	newEpisode.Episode = data.Episode
	newEpisode.Synopsis = data.Synopsis

	newEpisode.CreatedAt = time.Now()
	newEpisode.UpdatedAt = time.Now()

	return &newEpisode, nil
}
