package anime

import (
	"github.com/weeb-vip/anime-sync/internal/db"
	"log"
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
	var existingAnime Anime
	// find by id
	err := a.db.DB.Where("id = ?", anime.ID).First(&anime).Error
	if err != nil {
		if err.Error() != "record not found" {
			return err
		}
		// if not found, create new
		log.Println("Creating new anime record with ID:", anime.ID)
		err = a.db.DB.Create(anime).Error
		if err != nil {
			return err
		}
		return nil
	}
	anime.ID = existingAnime.ID
	err = a.db.DB.Save(anime).Error
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
