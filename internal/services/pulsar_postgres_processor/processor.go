package pulsar_postgres_processor

import (
	"database/sql"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"log"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarPostgresProcessorImpl interface {
	Process(data Payload) error
	parseToEntity(data Schema) (*anime.Anime, error)
}

type PulsarPostgresProcessor struct {
	Repository anime.AnimeRepositoryImpl
	Options    Options
}

func NewPulsarPostgresProcessor(opt Options, db *db.DB) PulsarPostgresProcessorImpl {
	return &PulsarPostgresProcessor{
		Repository: anime.NewAnimeRepository(db),
		Options:    opt,
	}
}

func (p *PulsarPostgresProcessor) Process(data Payload) error {

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

func (p *PulsarPostgresProcessor) parseToEntity(data Schema) (*anime.Anime, error) {
	var newAnime anime.Anime

	animeStartDate := anime.NullTime{
		Time:  time.Time{},
		Valid: false,
	}
	if data.StartDate != nil {
		startDate, err := time.Parse(time.RFC3339, *data.StartDate)
		if err != nil {
			return nil, err
		}
		animeStartDate = sql.NullTime{
			Time:  startDate,
			Valid: true,
		}

	}

	animeEndDate := anime.NullTime{
		Time:  time.Time{},
		Valid: false,
	}
	if data.EndDate != nil {
		endDate, err := time.Parse(time.RFC3339, *data.EndDate)
		if err != nil {
			return nil, err
		}
		animeEndDate = anime.NullTime{
			Time:  endDate,
			Valid: true,
		}
	}
	var record_type *anime.RECORD_TYPE
	if data.Type != nil {
		record := anime.RECORD_TYPE(*data.Type)
		record_type = &record

	}

	newAnime.ID = data.Id
	newAnime.Type = record_type
	newAnime.TitleEn = data.TitleEn
	newAnime.TitleJp = data.TitleJp
	newAnime.TitleRomaji = data.TitleRomaji
	newAnime.TitleKanji = data.TitleKanji
	newAnime.TitleSynonyms = data.TitleSynonyms
	newAnime.ImageURL = data.ImageUrl
	newAnime.Synopsis = data.Synopsis
	newAnime.Episodes = data.Episodes
	newAnime.Status = data.Status
	newAnime.StartDate = animeStartDate
	newAnime.EndDate = animeEndDate
	newAnime.Genres = data.Genres
	newAnime.Duration = data.Duration
	newAnime.Broadcast = data.Broadcast
	newAnime.Source = data.Source
	newAnime.Licensors = data.Licensors
	newAnime.Studios = data.Studios
	newAnime.Rating = data.Rating
	newAnime.CreatedAt = time.Now()
	newAnime.UpdatedAt = time.Now()

	return &newAnime, nil
}
