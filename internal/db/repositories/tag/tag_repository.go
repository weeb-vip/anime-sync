package tag

import (
	"github.com/weeb-vip/anime-sync/internal/db"
)

type TagRepositoryImpl interface {
	FindOrCreate(name string) (*Tag, error)
	FindByName(name string) (*Tag, error)
	FindByNames(names []string) ([]Tag, error)
	Create(tag *Tag) error
}

type TagRepository struct {
	db *db.DB
}

func NewTagRepository(db *db.DB) TagRepositoryImpl {
	return &TagRepository{db: db}
}

func (r *TagRepository) FindOrCreate(name string) (*Tag, error) {
	var tag Tag
	err := r.db.DB.Where("name = ?", name).FirstOrCreate(&tag, Tag{Name: name}).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) FindByName(name string) (*Tag, error) {
	var tag Tag
	err := r.db.DB.Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *TagRepository) FindByNames(names []string) ([]Tag, error) {
	var tags []Tag
	err := r.db.DB.Where("name IN ?", names).Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *TagRepository) Create(tag *Tag) error {
	return r.db.DB.Create(tag).Error
}
