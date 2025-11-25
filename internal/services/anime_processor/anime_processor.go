package anime_processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Flagsmith/flagsmith-go-client/v2"
	"github.com/ThatCatDev/ep/v2/event"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/anime-sync/internal"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/anime_tag"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/tag"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
)

type Options struct {
	NoErrorOnDelete bool
}

type AnimeProcessor interface {
	Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error)
}

type AnimeProcessorImpl struct {
	Repository         anime.AnimeRepositoryImpl
	TagRepository      tag.TagRepositoryImpl
	AnimeTagRepository anime_tag.AnimeTagRepositoryImpl
	Options            Options
	AlgoliaProducer    func(ctx context.Context, message *kafka.Message) error
	Producer           func(ctx context.Context, message *kafka.Message) error
}

func NewAnimeProcessor(opt Options, db *db.DB, algoliaProducer func(ctx context.Context, message *kafka.Message) error, producer func(ctx context.Context, message *kafka.Message) error) AnimeProcessor {
	return &AnimeProcessorImpl{
		Repository:         anime.NewAnimeRepository(db),
		TagRepository:      tag.NewTagRepository(db),
		AnimeTagRepository: anime_tag.NewAnimeTagRepository(db),
		Options:            opt,
		AlgoliaProducer:    algoliaProducer,
		Producer:           producer,
	}
}

func (p *AnimeProcessorImpl) Process(ctx context.Context, data event.Event[*kafka.Message, Payload]) (event.Event[*kafka.Message, Payload], error) {
	log := logger.FromCtx(ctx)

	payload := data.Payload

	log.Info("Getting flagsmith client from context")
	flagsmithClientInterface := ctx.Value(internal.FFClient{})
	if flagsmithClientInterface == nil {
		log.Error("Flagsmith client not found in context")
		return data, fmt.Errorf("flagsmith client not found in context")
	}

	// Try the new interface first
	flagsmithClient, ok := flagsmithClientInterface.(FlagSmithClient)
	if !ok {
		// Fall back to the raw Flagsmith client
		rawClient, ok := flagsmithClientInterface.(interface {
			GetEnvironmentFlags() (flagsmith.Flags, error)
		})
		if !ok {
			log.Error("Flagsmith client has wrong type")
			return data, fmt.Errorf("flagsmith client has wrong type")
		}

		log.Info("Getting environment flags from flagsmith client")
		flags, err := rawClient.GetEnvironmentFlags()
		if err != nil {
			log.Error("Failed to get environment flags", zap.Error(err))
			return data, fmt.Errorf("failed to get environment flags: %w", err)
		}

		log.Info("Checking if feature 'enable_kafka' is enabled")
		isEnabled, _ := flags.IsFeatureEnabled("enable_kafka")
		log.Info("Feature 'enable_kafka' is enabled", zap.Bool("isEnabled", isEnabled))

		return p.processPayload(ctx, data, payload, isEnabled, log)
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

	return p.processPayload(ctx, data, payload, isEnabled, log)
}

// processPayload handles the main processing logic after feature flag check
func (p *AnimeProcessorImpl) processPayload(ctx context.Context, data event.Event[*kafka.Message, Payload], payload Payload, isEnabled bool, log *zap.Logger) (event.Event[*kafka.Message, Payload], error) {
	// log the payload
	log.Debug("Payload", zap.Any("payload", payload))

	if payload.Before == nil && payload.After != nil {
		// add to db
		// Log the incoming payload for debugging
		if payload.After.TheTVDBID != nil {
			log.Info("Create operation with TheTVDBID", zap.String("id", payload.After.ID), zap.String("thetvdbid", *payload.After.TheTVDBID))
		} else {
			log.Info("Create operation without TheTVDBID", zap.String("id", payload.After.ID))
		}

		newAnime, err := p.ParseToEntity(ctx, *payload.After)
		if err != nil {
			return data, err
		}

		// Log the anime entity before saving
		if newAnime.TheTVDBID != nil {
			log.Info("Creating anime with TheTVDBID", zap.String("id", newAnime.ID), zap.String("thetvdbid", *newAnime.TheTVDBID))
		} else {
			log.Info("Creating anime without TheTVDBID", zap.String("id", newAnime.ID))
		}

		err = p.Repository.Upsert(newAnime, nil)
		if err != nil {
			return data, err
		}

		// Handle tag associations
		err = p.syncTags(ctx, newAnime.ID, payload.After.Genres)
		if err != nil {
			log.Warn("Failed to sync tags", zap.Error(err))
		}

		jsonAnime, err := json.Marshal(ProducerPayload{
			Action: CreateAction,
			Data:   payload.After,
		})
		if err != nil {
			log.Error("Error marshalling payload", zap.Error(err))
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
			log.Error("Error marshalling image payload", zap.Error(err))
			return data, err
		}

		err = p.AlgoliaProducer(ctx, &kafka.Message{
			Value: jsonAnime,
		})
		if err != nil {
			log.Error("Error sending message to algolia producer", zap.Error(err))
			return data, err
		}

		if payload.After.ImageUrl != nil {
			log.Info("Sending update to producer", zap.String("title", title), zap.String("imageURL", imageURL))
			log.Info("Sending image to Kafka", zap.String("imageURL", *payload.After.ImageUrl))
			err = p.Producer(ctx, &kafka.Message{
				Value: jsonImage,
			})

			if err != nil {
				log.Error("Error sending message to Kafka producer", zap.Error(err))
				return data, err
			}

		} else {
			log.Warn("ImageURL is nil, skipping image producer")
		}
	}

	if payload.After == nil && payload.Before != nil {
		// delete from db
		oldAnime, err := p.ParseToEntity(ctx, *payload.Before)
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
		// Log the incoming payload for debugging
		if payload.After.TheTVDBID != nil {
			log.Info("Update operation with TheTVDBID", zap.String("id", payload.After.ID), zap.String("thetvdbid", *payload.After.TheTVDBID))
		} else {
			log.Info("Update operation without TheTVDBID", zap.String("id", payload.After.ID))
		}

		newAnime, err := p.ParseToEntity(ctx, *payload.After)
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

		// Log the anime entity before saving
		if newAnime.TheTVDBID != nil {
			log.Info("Saving anime with TheTVDBID", zap.String("id", newAnime.ID), zap.String("thetvdbid", *newAnime.TheTVDBID))
		} else {
			log.Info("Saving anime without TheTVDBID", zap.String("id", newAnime.ID))
		}

		err = p.Repository.Upsert(newAnime, oldTitle)
		if err != nil {
			return data, err
		}

		// Handle tag associations
		err = p.syncTags(ctx, newAnime.ID, payload.After.Genres)
		if err != nil {
			log.Warn("Failed to sync tags", zap.Error(err))
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
		err = p.AlgoliaProducer(ctx, &kafka.Message{
			Value: jsonAnime,
		})
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
		} else {
			log.Warn("ImageURL is nil, skipping image producer")
		}
	}

	if payload.Before != nil && payload.After == nil {
		log.Warn("WARN: payload.After is nil, skipping update")
	}

	return data, nil
}

func (p *AnimeProcessorImpl) ParseToEntity(ctx context.Context, data Schema) (*anime.Anime, error) {
	log := logger.FromCtx(ctx)
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
	newAnime.TheTVDBID = data.TheTVDBID
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

	// Convert rating from string to float64
	var animeRating *float64
	if data.Rating != nil {
		if ratingFloat, err := strconv.ParseFloat(*data.Rating, 64); err == nil {
			animeRating = &ratingFloat
		} else {
			log.Warn("Failed to parse rating as float", zap.String("rating", *data.Rating), zap.Error(err))
		}
	}
	newAnime.Rating = animeRating

	newAnime.CreatedAt = time.Now()
	newAnime.UpdatedAt = time.Now()

	// Debug logging for TheTVDBID
	if data.TheTVDBID != nil {
		log.Info("TheTVDBID received in parseToEntity", zap.String("id", data.ID), zap.String("thetvdbid", *data.TheTVDBID))
	} else {
		log.Info("TheTVDBID is nil for anime", zap.String("id", data.ID))
	}

	return &newAnime, nil
}

func (p *AnimeProcessorImpl) syncTags(ctx context.Context, animeID string, genres *string) error {
	log := logger.FromCtx(ctx)

	if genres == nil || *genres == "" {
		// No genres, clear all tags
		return p.AnimeTagRepository.SetTagsForAnime(animeID, []int64{})
	}

	// Parse genres JSON array into tag names
	var genreList []string
	if err := json.Unmarshal([]byte(*genres), &genreList); err != nil {
		log.Warn("Failed to parse genres as JSON array", zap.String("genres", *genres), zap.Error(err))
		return p.AnimeTagRepository.SetTagsForAnime(animeID, []int64{})
	}

	var tagIDs []int64
	for _, genreName := range genreList {
		trimmedName := strings.TrimSpace(genreName)
		if trimmedName != "" {
			t, err := p.TagRepository.FindOrCreate(trimmedName)
			if err != nil {
				log.Warn("Failed to find or create tag", zap.String("tag", trimmedName), zap.Error(err))
				continue
			}
			tagIDs = append(tagIDs, t.ID)
		}
	}

	return p.AnimeTagRepository.SetTagsForAnime(animeID, tagIDs)
}
