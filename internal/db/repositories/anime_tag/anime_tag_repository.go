package anime_tag

import (
	"github.com/weeb-vip/anime-sync/internal/db"
)

type AnimeTagRepositoryImpl interface {
	SetTagsForAnime(animeID string, tagIDs []int64) error
	GetTagIDsForAnime(animeID string) ([]int64, error)
	AddTagToAnime(animeID string, tagID int64) error
	RemoveTagFromAnime(animeID string, tagID int64) error
	DeleteAllTagsForAnime(animeID string) error
}

type AnimeTagRepository struct {
	db *db.DB
}

func NewAnimeTagRepository(db *db.DB) AnimeTagRepositoryImpl {
	return &AnimeTagRepository{db: db}
}

// SetTagsForAnime replaces all tags for an anime with the given tag IDs
func (r *AnimeTagRepository) SetTagsForAnime(animeID string, tagIDs []int64) error {
	// Delete existing tag associations
	err := r.db.DB.Where("anime_id = ?", animeID).Delete(&AnimeTag{}).Error
	if err != nil {
		return err
	}

	// Insert new tag associations
	if len(tagIDs) > 0 {
		animeTags := make([]AnimeTag, len(tagIDs))
		for i, tagID := range tagIDs {
			animeTags[i] = AnimeTag{
				AnimeID: animeID,
				TagID:   tagID,
			}
		}
		err = r.db.DB.Create(&animeTags).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTagIDsForAnime returns all tag IDs associated with an anime
func (r *AnimeTagRepository) GetTagIDsForAnime(animeID string) ([]int64, error) {
	var animeTags []AnimeTag
	err := r.db.DB.Where("anime_id = ?", animeID).Find(&animeTags).Error
	if err != nil {
		return nil, err
	}

	tagIDs := make([]int64, len(animeTags))
	for i, at := range animeTags {
		tagIDs[i] = at.TagID
	}
	return tagIDs, nil
}

// AddTagToAnime adds a single tag to an anime
func (r *AnimeTagRepository) AddTagToAnime(animeID string, tagID int64) error {
	animeTag := AnimeTag{
		AnimeID: animeID,
		TagID:   tagID,
	}
	return r.db.DB.Create(&animeTag).Error
}

// RemoveTagFromAnime removes a single tag from an anime
func (r *AnimeTagRepository) RemoveTagFromAnime(animeID string, tagID int64) error {
	return r.db.DB.Where("anime_id = ? AND tag_id = ?", animeID, tagID).Delete(&AnimeTag{}).Error
}

// DeleteAllTagsForAnime removes all tags for an anime
func (r *AnimeTagRepository) DeleteAllTagsForAnime(animeID string) error {
	return r.db.DB.Where("anime_id = ?", animeID).Delete(&AnimeTag{}).Error
}
