package pulsar_anime_postgres_processor

type Action = string

type DataType = string

const (
	// DataTypeImage represents an image data type
	DataTypeAnime     DataType = "Anime"
	DataTypeCharacter DataType = "Character"
	DataTypeStaff     DataType = "Staff"
)

const (
	CreateAction Action = "create"
	UpdateAction Action = "update"
	DeleteAction Action = "delete"
)

type Schema struct {
	ID            string  `json:"id"`
	AnidbID       *string `json:"anidbid"`
	TheTVDBID     *string `json:"thetvdbid"`
	TitleEn       *string `json:"title_en"`
	TitleJp       *string `json:"title_jp"`
	TitleRomaji   *string `json:"title_romaji"`
	TitleKanji    *string `json:"title_kanji"`
	Type          *string `json:"type"`
	ImageUrl      *string `json:"image_url"`
	Synopsis      *string `json:"synopsis"`
	Episodes      *int    `json:"episodes"`
	Status        *string `json:"status"`
	Duration      *string `json:"duration"`
	Broadcast     *string `json:"broadcast"`
	Source        *string `json:"source"`
	CreatedAt     *int64  `json:"created_at"`
	UpdatedAt     *int64  `json:"updated_at"`
	Rating        *string `json:"rating"`
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	TitleSynonyms *string `json:"title_synonyms"`
	Genres        *string `json:"genres"`
	Licensors     *string `json:"licensors"`
	Studios       *string `json:"studios"`
	Ranking       *int    `json:"ranking"`
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

type ProducerPayload struct {
	Action string  `json:"action"`
	Data   *Schema `json:"data"`
}

type ImageSchema struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Type DataType `json:"type"`
}
type ImagePayload struct {
	Data ImageSchema `json:"data"`
}
