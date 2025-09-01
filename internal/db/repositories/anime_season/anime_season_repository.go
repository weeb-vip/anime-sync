package anime_season

import (
	"github.com/weeb-vip/anime-sync/internal/db"
)

type AnimeSeasonRepositoryImpl interface {
	Upsert(animeSeason *AnimeSeason) error
	Delete(animeSeason *AnimeSeason) error
}

type AnimeSeasonRepository struct {
	db *db.DB
}

func NewAnimeSeasonRepository(db *db.DB) AnimeSeasonRepositoryImpl {
	return &AnimeSeasonRepository{db: db}
}

func (r *AnimeSeasonRepository) Upsert(animeSeason *AnimeSeason) error {
	err := r.db.DB.Save(animeSeason).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AnimeSeasonRepository) Delete(animeSeason *AnimeSeason) error {
	err := r.db.DB.Delete(animeSeason).Error
	if err != nil {
		return err
	}
	return nil
}