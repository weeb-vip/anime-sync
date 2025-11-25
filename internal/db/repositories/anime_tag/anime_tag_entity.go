package anime_tag

import (
	"time"
)

type AnimeTag struct {
	AnimeID   string    `gorm:"column:anime_id;type:varchar(36);primaryKey" json:"anime_id"`
	TagID     int64     `gorm:"column:tag_id;primaryKey" json:"tag_id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName sets table name
func (AnimeTag) TableName() string {
	return "anime_tags"
}
