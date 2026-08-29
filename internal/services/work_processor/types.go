package work_processor

type Action = string

const (
	CreateAction Action = "create"
	UpdateAction Action = "update"
	DeleteAction Action = "delete"
)

// Schema mirrors the scraper's `work` table as Debezium emits it.
//
// score is a float because the column is double precision on both sides.
// Debezium's decimal.handling.mode defaults to `precise`, which would send a
// numeric column as base64 bytes and a scale -- unreadable as a float.
type Schema struct {
	ID            string   `json:"id"`
	MalID         *int     `json:"mal_id"`
	Type          string   `json:"type"`
	UrlSlug       *string  `json:"url_slug"`
	TitleEn       *string  `json:"title_en"`
	TitleJp       *string  `json:"title_jp"`
	TitleSynonyms *string  `json:"title_synonyms"`
	Synopsis      *string  `json:"synopsis"`
	ImageUrl      *string  `json:"image_url"`
	Status        *string  `json:"status"`
	Volumes       *int     `json:"volumes"`
	Chapters      *int     `json:"chapters"`
	PublishedFrom *string  `json:"published_from"`
	PublishedTo   *string  `json:"published_to"`
	Demographic   *string  `json:"demographic"`
	Serialization *string  `json:"serialization"`
	Authors       *string  `json:"authors"`
	Score         *float64 `json:"score"`
	Ranking       *int     `json:"ranking"`
	Members       *int     `json:"members"`
	Favorites     *int     `json:"favorites"`
	CreatedAt     *int64   `json:"created_at"`
	UpdatedAt     *int64   `json:"updated_at"`
}

type Source struct {
	Version   string      `json:"version"`
	Connector string      `json:"connector"`
	Name      string      `json:"name"`
	TsMs      int64       `json:"ts_ms"`
	Snapshot  string      `json:"snapshot"`
	Db        string      `json:"db"`
	Sequence  string      `json:"sequence"`
	Schema    string      `json:"schema"`
	Table     string      `json:"table"`
	TxId      int         `json:"txId"`
	Lsn       int         `json:"lsn"`
	Xmin      interface{} `json:"xmin"`
}

type Payload struct {
	Before *Schema `json:"before"`
	After  *Schema `json:"after"`
	Source Source  `json:"source"`
}
