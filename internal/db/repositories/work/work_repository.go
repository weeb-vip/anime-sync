package work

import (
	"github.com/weeb-vip/anime-sync/internal/db"
)

type WorkRepositoryImpl interface {
	Upsert(work *Work) error
	Delete(work *Work) error
}

type WorkRepository struct {
	db *db.DB
}

func NewWorkRepository(db *db.DB) WorkRepositoryImpl {
	return &WorkRepository{db: db}
}

func (r *WorkRepository) Upsert(work *Work) error {
	return r.db.DB.Save(work).Error
}

func (r *WorkRepository) Delete(work *Work) error {
	return r.db.DB.Delete(work).Error
}
