package pulsar_postgres_processor

import (
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"log"
	"time"
)

type PulsarPostgresProcessorImpl interface {
	Process(data Payload) error
	parseToEntity(data Schema) (*anime.Anime, error)
}

type PulsarPostgresProcessor struct {
	Repository anime.AnimeRepositoryImpl
}

func NewPulsarPostgresProcessor(db *db.DB) PulsarPostgresProcessorImpl {
	return &PulsarPostgresProcessor{
		Repository: anime.NewAnimeRepository(db),
	}
}

func (p *PulsarPostgresProcessor) Process(data Payload) error {

	if data.Before == nil && data.After != nil {
		// add to db
		newAnime, err := p.parseToEntity(*data.After)
		if err != nil {
			return err
		}
		p.Repository.Upsert(newAnime)
	}

	if data.After == nil && data.Before != nil {
		// delete from db
		oldAnime, err := p.parseToEntity(*data.Before)
		if err != nil {
			return err
		}

		p.Repository.Delete(oldAnime)
	}

	if data.Before != nil && data.After != nil {
		// update db
		newAnime, err := p.parseToEntity(*data.After)
		if err != nil {
			return err
		}
		p.Repository.Upsert(newAnime)
	}

	if data.Before != nil && data.After == nil {
		log.Println("WARN: data.After is nil, skipping update")
	}

	return nil

}

func (p *PulsarPostgresProcessor) parseToEntity(data Schema) (*anime.Anime, error) {
	var newAnime anime.Anime

	var animeStartDate time.Time
	if data.StartDate != nil {
		startDate, err := time.Parse(time.RFC3339, *data.StartDate)
		if err != nil {
			return nil, err
		}
		animeStartDate = startDate
	}

	var animeEndDate time.Time
	if data.EndDate != nil {
		endDate, err := time.Parse(time.RFC3339, *data.EndDate)
		if err != nil {
			return nil, err
		}
		animeEndDate = endDate
	}
	newAnime.ID = data.Id
	newAnime.Type = anime.RECORD_TYPE(*data.Type)
	newAnime.TitleEn = *data.TitleEn
	newAnime.TitleJp = *data.TitleJp
	newAnime.TitleRomaji = *data.TitleRomaji
	newAnime.TitleKanji = *data.TitleKanji
	newAnime.TitleSynonyms = *data.TitleSynonyms
	newAnime.ImageURL = *data.ImageUrl
	newAnime.Synopsis = *data.Synopsis
	newAnime.Episodes = *data.Episodes
	newAnime.Status = *data.Status
	newAnime.StartDate = &animeStartDate
	newAnime.EndDate = &animeEndDate
	newAnime.Genres = *data.Genres
	newAnime.Duration = *data.Duration
	newAnime.Broadcast = *data.Broadcast
	newAnime.Source = *data.Source
	newAnime.Licensors = *data.Licensors
	newAnime.Studios = *data.Studios
	newAnime.Rating = *data.Rating
	newAnime.CreatedAt = time.Now()
	newAnime.UpdatedAt = time.Now()

	return &newAnime, nil
}
