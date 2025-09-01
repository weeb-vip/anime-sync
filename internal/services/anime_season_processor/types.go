package anime_season_processor

type Action = string

const (
	CreateAction Action = "create"
	UpdateAction Action = "update"
	DeleteAction Action = "delete"
)

type Schema struct {
	ID           string  `json:"id"`
	Season       string  `json:"season"`
	Status       string  `json:"status"`
	EpisodeCount *int    `json:"episode_count"`
	Notes        *string `json:"notes"`
	CreatedAt    *int64  `json:"created_at"`
	UpdatedAt    *int64  `json:"updated_at"`
	AnimeID      *string `json:"anime_id"`
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