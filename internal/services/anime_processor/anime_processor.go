package anime_processor

import (
	"context"
	"encoding/json"
	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/internal"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"github.com/weeb-vip/anime-sync/internal/producer"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Options struct {
	NoErrorOnDelete bool
}

type AnimeProcessor interface {
	Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error)
}

type AnimeProcessorImpl struct {
	Repository      anime.AnimeRepositoryImpl
	Options         Options
	AlgoliaProducer producer.Producer[Schema]
	Producer        func(ctx context.Context, message *kafka.Message) error
}

func NewAnimeProcessor(opt Options, db *db.DB, algoliaProducer producer.Producer[Schema], producer func(ctx context.Context, message *kafka.Message) error) AnimeProcessor {
	return &AnimeProcessorImpl{
		Repository:      anime.NewAnimeRepository(db),
		Options:         opt,
		AlgoliaProducer: algoliaProducer,
		Producer:        producer,
	}
}

func (p *AnimeProcessorImpl) Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	log.Info("Gettting flagsmith client from context")
	flagsmithClient, _ := ctx.Value(internal.FFClient{}).(*flagsmith.Client)

	log.Info("Getting environment flags from flagsmith client")
	flags, _ := flagsmithClient.GetEnvironmentFlags()

	log.Info("Checking if feature 'enable_kafka' is enabled")
	isEnabled, _ := flags.IsFeatureEnabled("enable_kafka")
	log.Info("Feature 'enable_kafka' is enabled", zap.Bool("isEnabled", isEnabled))

	if payload.Before == nil && payload.After != nil {
		// add to db
		newAnime, err := p.parseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}
		err = p.Repository.Upsert(newAnime, nil)
		if err != nil {
			return data, err
		}
		jsonAnime, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   payload.After,
		})
		if err != nil {
			return data, err
		}
		var title string
		if payload.After.TitleEn != nil {
			title = strings.ToLower(*payload.After.TitleEn)
			title = strings.ReplaceAll(title, " ", "_")
		} else if payload.After.TitleJp != nil {
			title = strings.ToLower(*payload.After.TitleJp)
			title = strings.ReplaceAll(title, " ", "_")
		} else {
			return data, nil
		}
		imageURL := ""
		if payload.After.ImageUrl != nil {
			imageURL = *payload.After.ImageUrl
		}
		imagePayload := &ImagePayload{
			Data: ImageSchema{
				Name: title,
				URL:  imageURL,
				Type: DataTypeAnime,
			},
		}

		jsonImage, err := json.Marshal(imagePayload)
		if err != nil {
			return data, err
		}

		err = p.AlgoliaProducer.Send(ctx, jsonAnime)
		if err != nil {
			return data, err
		}

		if payload.After.ImageUrl != nil {
			log.Info("Sending update to producer", zap.String("title", title), zap.String("imageURL", imageURL))
			log.Info("Sending image to Kafka", zap.String("imageURL", *payload.After.ImageUrl))
			err = p.Producer(ctx, &kafka.Message{
				Value: jsonImage,
			})

			if err != nil {
				return data, err
			}

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
		var oldTitle *string
		if payload.Before.TitleEn != nil && *payload.Before.TitleEn != *payload.After.TitleEn {
			oldTitle = payload.Before.TitleEn
		}
		if payload.Before.TitleEn == nil && payload.Before.TitleJp != nil && *payload.Before.TitleJp != *payload.After.TitleJp {
			oldTitle = payload.Before.TitleJp
		}
		err = p.Repository.Upsert(newAnime, oldTitle)
		if err != nil {
			return data, err
		}
		// convert new anime to json
		jsonAnime, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   payload.After,
		})
		if err != nil {
			return data, err
		}
		var title string
		if payload.After.TitleEn != nil {
			title = strings.ToLower(*payload.After.TitleEn)
			title = strings.ReplaceAll(title, " ", "_")
		} else if payload.After.TitleJp != nil {
			title = strings.ToLower(*payload.After.TitleJp)
			title = strings.ReplaceAll(title, " ", "_")
		} else {
			return data, nil
		}
		imageURL := ""
		if payload.After.ImageUrl != nil {
			imageURL = *payload.After.ImageUrl
		}
		imagePayload := &ImagePayload{
			Data: ImageSchema{
				Name: title,
				URL:  imageURL,
				Type: DataTypeAnime,
			},
		}

		jsonImage, err := json.Marshal(imagePayload)
		if err != nil {
			return data, err
		}
		err = p.AlgoliaProducer.Send(ctx, jsonAnime)
		if err != nil {
			return data, err
		}
		if payload.After.ImageUrl != nil {
			log.Info("Sending update to producer", zap.String("title", title), zap.String("imageURL", imageURL))

			log.Info("Sending image to Kafka producer", zap.String("title", title), zap.String("imageURL", imageURL))
			err = p.Producer(ctx, &kafka.Message{
				Value: jsonImage,
			})

			if err != nil {
				return data, err
			}
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: payload.After is nil, skipping update")
	}

	return data, nil
}

func (p *AnimeProcessorImpl) parseToEntity(ctx context.Context, data Schema) (*anime.Anime, error) {
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

	newAnime.ID = data.ID
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
