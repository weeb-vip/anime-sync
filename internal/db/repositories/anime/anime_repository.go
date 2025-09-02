package anime

import (
	"context"
	"github.com/weeb-vip/anime-sync/internal/db"
)

type RECORD_TYPE string

type AnimeRepositoryImpl interface {
	Upsert(anime *Anime, oldTitle *string) error
	Delete(anime *Anime) error
}

type AnimeRepository struct {
	db *db.DB
}

func NewAnimeRepository(db *db.DB) AnimeRepositoryImpl {
	return &AnimeRepository{db: db}
}

func (a *AnimeRepository) Upsert(anime *Anime, oldTitle *string) error {
	// Log the anime struct before saving for debugging
	if anime.TheTVDBID != nil {
		a.db.DB.Logger.Info(context.Background(), "Upserting anime with TheTVDBID: ID=%s, TheTVDBID=%s", anime.ID, *anime.TheTVDBID)
	} else {
		a.db.DB.Logger.Info(context.Background(), "Upserting anime without TheTVDBID: ID=%s", anime.ID)
	}

	err := a.db.DB.Save(anime).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *AnimeRepository) Delete(anime *Anime) error {
	err := a.db.DB.Delete(anime).Error
	if err != nil {
		return err
	}
	return nil
}
