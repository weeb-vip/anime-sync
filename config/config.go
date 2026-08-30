package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig   AppConfig
	DBConfig    DBConfig
	KafkaConfig KafkaConfig
	NatsConfig  NatsConfig
}

type AppConfig struct {
	APPName string `default:"anime-api"`
	Port    int    `env:"PORT" default:"3000"`
	Version string `default:"x.x.x"`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"weeb" env:"DBNAME"`
	User     string `default:"weeb" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"5432" env:"DBPORT"`
	SSLMode  string `default:"require" env:"DBSSL"`
}

type KafkaConfig struct {
	ConsumerGroupName string `default:"image-sync-group" env:"KAFKA_CONSUMER_GROUP_NAME"`
	BootstrapServers  string `default:"localhost:9092" env:"KAFKA_BOOTSTRAP_SERVERS"`
	Offset            string `default:"earliest" env:"KAFKA_OFFSET"`
	Topic             string `default:"anime-db.public.anime" env:"KAFKA_TOPIC"`
	ProducerTopic     string `default:"image-sync" env:"KAFKA_PRODUCER_TOPIC"`
	AlgoliaTopic      string `default:"algolia-sync" env:"KAFKA_ALGOLIA_TOPIC"`
}

// NatsConfig mirrors KafkaConfig field for field, so a service moving between
// the two has one obvious substitution per setting rather than a translation.
//
// The names differ where the systems genuinely differ. A NATS "subject" is what
// Kafka calls a topic, and Debezium publishes to subjects named exactly like the
// topics it used to write, so the values carry over unchanged --
// anime-db-staging.public.anime is both.
type NatsConfig struct {
	URL string `default:"nats://localhost:4222" env:"NATSURL"`

	// The durable consumer name. Like a Kafka consumer group, every instance
	// sharing it shares one subscription's workload and ack state; unlike one,
	// leaving it empty makes the consumer ephemeral and its position is lost on
	// restart.
	ConsumerGroupName string `default:"anime-sync-nats" env:"NATSCONSUMERGROUPNAME"`

	// The stream to bind to, rather than one derived from the subject.
	//
	// Debezium owns the CDC stream and declares it over anime-db-staging.>.
	// JetStream refuses two streams whose subjects overlap, so a consumer that
	// created its own per-subject stream would fail against it. Naming the
	// stream makes the driver bind to the existing one instead.
	StreamName string `default:"ANIMEDBSTAGING" env:"NATSSTREAMNAME"`

	Offset string `default:"earliest" env:"NATSOFFSET"`

	Subject string `default:"anime-db-staging.public.anime" env:"NATSSUBJECT"`

	// Outbound subjects. These are not CDC, so nothing else declares a stream
	// over them and the driver creates one per subject as needed.
	ProducerSubject string `default:"image-sync" env:"NATSPRODUCERSUBJECT"`
	AlgoliaSubject  string `default:"algolia-sync" env:"NATSALGOLIASUBJECT"`
	// Works are indexed separately from anime, so they get their own subject
	// rather than sharing AlgoliaSubject with a type discriminator. The two
	// records have almost no fields in common -- a work has volumes, chapters,
	// authors and a serialization; an anime has episodes, studios and a
	// broadcast -- so one index would be half-empty in both directions and
	// every anime search would have to filter manga out.
	//
	// A distinct default rather than reusing AlgoliaSubject: if this were left
	// to an environment variable and that variable went missing, works would
	// publish onto the anime subject and quietly corrupt the anime index.
	AlgoliaWorkSubject string `default:"algolia-sync-work" env:"NATSALGOLIAWORKSUBJECT"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}
