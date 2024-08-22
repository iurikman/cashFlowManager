package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	BindAddress string `env:"BIND_ADDRESS" env-default:":8080"`

	PostgresHost     string `env:"POSTGRES_HOST" env-default:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" env-default:"5432"`
	PostgresDatabase string `env:"POSTGRES_DATABASE" env-default:"postgres"`
	PostgresUser     string `env:"POSTGRES_USER" env-default:"admin"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" env-default:"admin"`

	KafkaBrokers  []string `env:"KAFKA_BROKERS" env-default:"127.0.0.1:9092"`
	KafkaTopic    string   `env:"KAFKA_TOPIC" env-default:"users_updates"`
	KafkaGroupID  string   `env:"KAFKA_GROUP_ID" env-default:"users_group_id"`
	KafkaBalancer string   `env:"KAFKA_BALANCER" env-default:"least_bytes"`
	KafkaAddress  string   `env:"KAFKA_ADDRESS" env-default:"127.0.0.1:9092"`
}

func NewConfig() Config {
	config := Config{}
	if err := cleanenv.ReadEnv(&config); err != nil {
		panic(fmt.Errorf("error reading config: %w", err))
	}

	log.Infof("Config loaded %+v", config)

	return config
}
