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
	// Strings, not epoch micros. work.created_at is timestamptz, which Debezium
	// sends as io.debezium.time.ZonedTimestamp -- "2026-08-29T21:14:47.291952Z".
	// Only a plain `timestamp` column arrives as an int64, which is what the
	// anime_seasons processor this was modelled on happens to have.
	//
	// Neither is read: parseToEntity stamps time.Now(), matching the other
	// processors, because these record when this store learned of the change.
	// They are declared so the payload unmarshals at all -- an int64 here made
	// every work event fail on decode and killed the consumer.
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
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

// ProducerPayload is what goes onto the search subject. The consumer keys off
// Action, so a delete carries the row as it last was rather than nothing.
type ProducerPayload struct {
	Action Action  `json:"action"`
	Data   *Schema `json:"data"`
}

type Payload struct {
	Before *Schema `json:"before"`
	After  *Schema `json:"after"`
	Source Source  `json:"source"`
}

// The image-sync contract. Mirrors anime_processor's copy rather than importing
// it: they are two producers speaking one wire format, and coupling the packages
// so one can borrow a struct would make a change to either a change to both.
type DataType = string

// DataTypeWork files the cover under /works/<id> in the bucket, away from the
// root where anime posters live.
const DataTypeWork DataType = "Work"

type ImageSchema struct {
	// ID is what image-sync keys the object by. Name is sent alongside because
	// the consumer still falls back to it for messages published before ids.
	ID   string   `json:"id"`
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Type DataType `json:"type"`
}

type ImagePayload struct {
	Data ImageSchema `json:"data"`
}
