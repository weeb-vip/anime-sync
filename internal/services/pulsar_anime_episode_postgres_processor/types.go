package pulsar_anime_postgres_processor

import "time"

type Schema struct {
	Id            string     `json:"id"`
	AnimeId       *string    `json:"anime_id"`
	Episode       *int       `json:"episode"`
	TitleEn       *string    `json:"title_en"`
	TitleJp       *string    `json:"title_jp"`
	Aired         *time.Time `json:"aired"`
	Synopsis      *string    `json:"synopsis"`
	TitleSynonyms *string    `json:"title_synonyms"`
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
