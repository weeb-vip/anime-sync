package work_processor

import (
	"context"
	"time"

	"github.com/ThatCatDev/ep/v2/event"
	"github.com/weeb-vip/anime-sync/internal/db"
	"github.com/weeb-vip/anime-sync/internal/db/repositories/work"
	"github.com/weeb-vip/anime-sync/internal/logger"
	"go.uber.org/zap"
)

type Options struct {
	NoErrorOnDelete bool
}

// Generic over the driver message for the same reason the other processors are:
// nothing here reads DriverMessage, RawData or Headers, only Payload, which the
// transform middleware has already filled in.
//
// No producer. The other CDC processors republish to the image and algolia
// subjects; a work has no cover to fetch through image-sync and is not in the
// search index, so there is nothing downstream to notify. Adding an empty
// producer would only be somewhere for a future mistake to hide.
type WorkProcessor[DM any] interface {
	Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error)
}

type WorkProcessorImpl[DM any] struct {
	Repository work.WorkRepositoryImpl
	Options    Options
}

func NewWorkProcessor[DM any](opt Options, database *db.DB) WorkProcessor[DM] {
	return &WorkProcessorImpl[DM]{
		Repository: work.NewWorkRepository(database),
		Options:    opt,
	}
}

func (p *WorkProcessorImpl[DM]) Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error) {
	log := logger.FromCtx(ctx)
	payload := data.Payload

	switch {
	case payload.Before == nil && payload.After != nil,
		payload.Before != nil && payload.After != nil:
		// Create and update are the same write: Save upserts on the primary key,
		// and the scraper is the only writer, so there is no field this side
		// knows better than the event does.
		newWork := p.parseToEntity(*payload.After)
		if err := p.Repository.Upsert(newWork); err != nil {
			log.Error("Error upserting work", zap.String("id", payload.After.ID), zap.Error(err))
			return data, err
		}

	case payload.After == nil && payload.Before != nil:
		oldWork := p.parseToEntity(*payload.Before)
		if err := p.Repository.Delete(oldWork); err != nil {
			if p.Options.NoErrorOnDelete {
				log.Warn("WARN: error deleting work from db", zap.Error(err))
				return data, nil
			}
			return data, err
		}

	default:
		log.Warn("WARN: work event with neither before nor after, skipping")
	}

	return data, nil
}

func (p *WorkProcessorImpl[DM]) parseToEntity(data Schema) *work.Work {
	return &work.Work{
		ID:            data.ID,
		MalID:         data.MalID,
		Type:          data.Type,
		UrlSlug:       data.UrlSlug,
		TitleEn:       data.TitleEn,
		TitleJp:       data.TitleJp,
		TitleSynonyms: data.TitleSynonyms,
		Synopsis:      data.Synopsis,
		ImageURL:      data.ImageUrl,
		Status:        data.Status,
		Volumes:       data.Volumes,
		Chapters:      data.Chapters,
		PublishedFrom: data.PublishedFrom,
		PublishedTo:   data.PublishedTo,
		Demographic:   data.Demographic,
		Serialization: data.Serialization,
		Authors:       data.Authors,
		Score:         data.Score,
		Ranking:       data.Ranking,
		Members:       data.Members,
		Favorites:     data.Favorites,
		// Matching the other processors: the row's own timestamps record when
		// this store learned of the change, not when the scraper made it. The
		// event carries the source times if they are ever wanted.
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
