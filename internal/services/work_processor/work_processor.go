package work_processor

import (
	"context"
	"encoding/json"
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
// Two producers. Covers go to image-sync so they reach the CDN like every other
// image in the product -- the scraper stores MyAnimeList's own URL, and serving
// that directly would put an external host in the hot path of a page whose art
// is the point. The record itself goes to the search index.
//
// Search was deliberately left out when works only ever existed as the thing an
// anime was adapted from, reachable from that anime's page and never searched
// for directly. `collect manga` changed that: it discovers works nothing
// adapts, and those have no anime linking to them, no listing query, and a slug
// a visitor cannot guess. Unindexed, they are unreachable.
type WorkProcessor[DM any] interface {
	Process(ctx context.Context, data event.Event[DM, Payload]) (event.Event[DM, Payload], error)
}

type WorkProcessorImpl[DM any] struct {
	Repository      work.WorkRepositoryImpl
	Options         Options
	ImageProducer   func(ctx context.Context, value []byte) error
	AlgoliaProducer func(ctx context.Context, value []byte) error
}

func NewWorkProcessor[DM any](opt Options, database *db.DB, imageProducer func(ctx context.Context, value []byte) error, algoliaProducer func(ctx context.Context, value []byte) error) WorkProcessor[DM] {
	return &WorkProcessorImpl[DM]{
		Repository:      work.NewWorkRepository(database),
		Options:         opt,
		ImageProducer:   imageProducer,
		AlgoliaProducer: algoliaProducer,
	}
}

// publishSearch puts the work into the search index, or takes it out.
//
// Its own subject and its own index, not the anime one. The two records share
// almost no fields, so a single index would be half-empty whichever record it
// held, and every anime search would have to filter manga back out.
//
// This matters more than it did when works only existed as the thing an anime
// was adapted from: `collect manga` discovers works nothing adapts, and those
// are reachable only by search. Nothing links to them, there is no listing
// query, and workBySlug needs a slug the visitor has no way to learn. An
// unindexed work of that kind cannot be found at all.
func (p *WorkProcessorImpl[DM]) publishSearch(ctx context.Context, action Action, data Schema) error {
	if p.AlgoliaProducer == nil {
		return nil
	}

	encoded, err := json.Marshal(&ProducerPayload{
		Action: action,
		Data:   &data,
	})
	if err != nil {
		return err
	}

	return p.AlgoliaProducer(ctx, encoded)
}

// publishCover asks image-sync to fetch this work's cover onto the CDN.
//
// A work with no image is the ordinary case for a sparse MyAnimeList entry, not
// a failure, so it is skipped rather than published with an empty URL -- which
// image-sync would accept and then fail to fetch.
func (p *WorkProcessorImpl[DM]) publishCover(ctx context.Context, data Schema) error {
	if p.ImageProducer == nil || data.ImageUrl == nil || *data.ImageUrl == "" {
		return nil
	}

	title := ""
	if data.TitleEn != nil {
		title = *data.TitleEn
	} else if data.TitleJp != nil {
		title = *data.TitleJp
	}

	encoded, err := json.Marshal(&ImagePayload{
		Data: ImageSchema{
			ID:   data.ID,
			Name: title,
			URL:  *data.ImageUrl,
			Type: DataTypeWork,
		},
	})
	if err != nil {
		return err
	}

	return p.ImageProducer(ctx, encoded)
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

		if err := p.publishCover(ctx, *payload.After); err != nil {
			log.Error("Error sending work cover to image-sync", zap.String("id", payload.After.ID), zap.Error(err))
			return data, err
		}

		if err := p.publishSearch(ctx, UpdateAction, *payload.After); err != nil {
			log.Error("Error sending work to algolia", zap.String("id", payload.After.ID), zap.Error(err))
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

		// The index has to be told as well. These events are the only thing
		// feeding it, and a deleted row produces no further ones, so a work
		// left in the index stays searchable forever and every hit on it is a
		// 404 -- which is exactly how ~2,860 merged-away anime lingered there.
		if err := p.publishSearch(ctx, DeleteAction, *payload.Before); err != nil {
			log.Error("Error sending work delete to algolia", zap.String("id", payload.Before.ID), zap.Error(err))
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
