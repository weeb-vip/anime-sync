package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig    AppConfig
	DBConfig     DBConfig
	PulsarConfig PulsarConfig
	KafkaConfig  KafkaConfig
	FFConfig     FFConfig
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
	Port     uint   `default:"3306" env:"DBPORT"`
	SSLMode  string `default:"false" env:"DBSSL"`
}

type PulsarConfig struct {
	URL                  string `default:"pulsar://localhost:6650" env:"PULSARURL"`
	Topic                string `default:"public/default/myanimelist.public.anime" env:"PULSARTOPIC"`
	SubscribtionName     string `default:"my-sub" env:"PULSARSUBSCRIPTIONNAME"`
	ProducerAlgoliaTopic string `default:"public/default/myanimelist.public.anime-algolia" env:"PULSARALGOLIATOPIC"`
	ProducerImageTopic   string `default:"public/default/myanimelist.public.anime-image" env:"PULSARIMAGETOPIC"`
}

type KafkaConfig struct {
	ConsumerGroupName string `default:"image-sync-group" env:"KAFKA_CONSUMER_GROUP_NAME"`
	BootstrapServers  string `default:"localhost:9092" env:"KAFKA_BOOTSTRAP_SERVERS"`
	Topic             string `default:"image-sync-topic" env:"KAFKA_TOPIC"`
}

type FFConfig struct {
	APIKey  string `default:"" env:"FF_API_KEY"`
	BaseURL string `default:"http://flagsmith-api.weeb.svc.cluster.local" env:"FF_BASE_URL"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}
