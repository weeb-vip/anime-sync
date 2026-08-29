package work

import "time"

// Work is a manga, light novel, novel, manhwa or manhua -- the thing an anime
// is adapted from. One table for the whole family because MyAnimeList serves
// them from one namespace and their fields are identical apart from
// demographic; the type column is what tells them apart.
type Work struct {
	ID    string `gorm:"column:id;primaryKey" json:"id"`
	MalID *int   `gorm:"column:mal_id;null" json:"mal_id"`
	Type  string `gorm:"column:type" json:"type"`

	// Assigned by a trigger in the scraper and never recomputed, so it is
	// carried across rather than generated here -- a second generator would
	// mint a different slug for the same work.
	UrlSlug *string `gorm:"column:url_slug;null" json:"url_slug"`

	TitleEn       *string `gorm:"column:title_en;null" json:"title_en"`
	TitleJp       *string `gorm:"column:title_jp;null" json:"title_jp"`
	TitleSynonyms *string `gorm:"column:title_synonyms;type:text;null" json:"title_synonyms"`
	Synopsis      *string `gorm:"column:synopsis;type:text;null" json:"synopsis"`
	ImageURL      *string `gorm:"column:image_url;null" json:"image_url"`
	Status        *string `gorm:"column:status;null" json:"status"`
	Volumes       *int    `gorm:"column:volumes;null" json:"volumes"`
	Chapters      *int    `gorm:"column:chapters;null" json:"chapters"`
	PublishedFrom *string `gorm:"column:published_from;null" json:"published_from"`
	PublishedTo   *string `gorm:"column:published_to;null" json:"published_to"`
	Demographic   *string `gorm:"column:demographic;null" json:"demographic"`
	Serialization *string `gorm:"column:serialization;null" json:"serialization"`
	Authors       *string `gorm:"column:authors;type:text;null" json:"authors"`

	// double precision on both sides. Debezium encodes a numeric column as
	// base64 bytes under its default decimal.handling.mode, which does not
	// unmarshal into a float at all.
	Score *float64 `gorm:"column:score;null" json:"score"`

	Ranking   *int      `gorm:"column:ranking;null" json:"ranking"`
	Members   *int      `gorm:"column:members;null" json:"members"`
	Favorites *int      `gorm:"column:favorites;null" json:"favorites"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Work) TableName() string {
	return "work"
}
