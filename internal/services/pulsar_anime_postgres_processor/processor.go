package pulsar_anime_postgres_processor

import (
	"context"
	"encoding/json"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"github.com/weeb-vip/anime-sync/internal/producer"
	"log"
	"strings"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type PulsarAnimePostgresProcessorImpl interface {
	Process(ctx context.Context, data Payload) error
	parseToEntity(ctx context.Context, data Schema) (*anime.Anime, error)
}

type PulsarAnimePostgresProcessor struct {
	Repository    anime.AnimeRepositoryImpl
	Options       Options
	Producer      producer.Producer[Schema]
	ProducerImage producer.Producer[ImageSchema]
}

func NewPulsarAnimePostgresProcessor(opt Options, db *db.DB, producer producer.Producer[Schema], producerImage producer.Producer[ImageSchema]) PulsarAnimePostgresProcessorImpl {
	return &PulsarAnimePostgresProcessor{
		Repository:    anime.NewAnimeRepository(db),
		Options:       opt,
		Producer:      producer,
		ProducerImage: producerImage,
	}
}

func (p *PulsarAnimePostgresProcessor) Process(ctx context.Context, data Payload) error {

	if data.Before == nil && data.After != nil {
		// add to db
		newAnime, err := p.parseToEntity(ctx, *data.After)
		if err != nil {
			return err
		}
		err = p.Repository.Upsert(newAnime, nil)
		if err != nil {
			return err
		}
		jsonAnime, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   data.After,
		})
		if err != nil {
			return err
		}
		var title string
		if data.After.TitleEn != nil {
			title = strings.ToLower(*data.After.TitleEn)
			title = strings.ReplaceAll(title, " ", "_")
		} else if data.After.TitleJp != nil {
			title = strings.ToLower(*data.After.TitleJp)
			title = strings.ReplaceAll(title, " ", "_")
		} else {
			return nil
		}
		payload := &ImagePayload{
			Data: ImageSchema{
				Name: title,
				URL:  *data.After.ImageUrl,
				Type: DataTypeAnime,
			},
		}
		jsonImage, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		err = p.Producer.Send(ctx, jsonAnime)
		if err != nil {
			return err
		}

		err = p.ProducerImage.Send(ctx, jsonImage)
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
		var oldTitle *string
		if data.Before.TitleEn != nil && *data.Before.TitleEn != *data.After.TitleEn {
			oldTitle = data.Before.TitleEn
		}
		if data.Before.TitleEn == nil && data.Before.TitleJp != nil && *data.Before.TitleJp != *data.After.TitleJp {
			oldTitle = data.Before.TitleJp
		}
		err = p.Repository.Upsert(newAnime, oldTitle)
		if err != nil {
			return err
		}
		// convert new anime to json
		jsonAnime, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   data.After,
		})
		if err != nil {
			return err
		}
		var title string
		if data.After.TitleEn != nil {
			title = strings.ToLower(*data.After.TitleEn)
			title = strings.ReplaceAll(title, " ", "_")
		} else if data.After.TitleJp != nil {
			title = strings.ToLower(*data.After.TitleJp)
			title = strings.ReplaceAll(title, " ", "_")
		} else {
			return nil
		}
		payload := &ImagePayload{
			Data: ImageSchema{
				Name: title,
				URL:  *data.After.ImageUrl,
				Type: DataTypeAnime,
			},
		}
		jsonImage, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		err = p.Producer.Send(ctx, jsonAnime)
		if err != nil {
			return err
		}

		err = p.ProducerImage.Send(ctx, jsonImage)
		if err != nil {
			return err
		}
	}

	if data.Before != nil && data.After == nil {
		log.Println("WARN: data.After is nil, skipping update")
	}

	return nil

}

func (p *PulsarAnimePostgresProcessor) parseToEntity(ctx context.Context, data Schema) (*anime.Anime, error) {
	var newAnime anime.Anime

	var animeStartDate *string
	if data.StartDate != nil {
		startDate, err := time.Parse(time.RFC3339, *data.StartDate)
		if err != nil {
			return nil, err
		}
		animeStartDateFormatted := startDate.Format("2006-01-02 15:04:05")
		animeStartDate = &animeStartDateFormatted

	}

	var animeEndDate *string
	if data.EndDate != nil {
		endDate, err := time.Parse(time.RFC3339, *data.EndDate)
		if err != nil {
			return nil, err
		}
		animeEndDateFormatted := endDate.Format("2006-01-02 15:04:05")
		animeEndDate = &animeEndDateFormatted
	}
	var record_type *anime.RECORD_TYPE
	if data.Type != nil {
		record := anime.RECORD_TYPE(*data.Type)
		record_type = &record

	}

	newAnime.ID = data.Id
	newAnime.Ranking = data.Ranking
	newAnime.AnidbID = data.AnidbID
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
