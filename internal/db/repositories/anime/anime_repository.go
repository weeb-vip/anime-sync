package anime

import (
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
