package config

import (
	"fmt"
	"net"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	BindAddress string `env:"BIND_ADDRESS" env-default:":8080"`

	PostgresHost     string `env:"POSTGRES_HOST" env-default:"localhost"`
	PostgresPort     string `env:"POSTGRES_PORT" env-default:"5432"`
	PostgresDatabase string `env:"POSTGRES_DATABASE" env-default:"postgres"`
	PostgresUser     string `env:"POSTGRES_USER" env-default:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" env-default:"postgres"`

	KafkaBrokers  []string       `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	KafkaTopic    string         `env:"KAFKA_TOPIC" env-default:"wallets"`
	KafkaGroupID  string         `env:"KAFKA_GROUP_ID" env-default:"wallets_group_id"`
	KafkaBalancer kafka.Balancer `env:"KAFKA_BALANCER" env-default:"least_bytes"`
	KafkaAddress  net.Addr       `env:"KAFKA_ADDRESS" env-default:"localhost:9092"`
}

func NewConfig() Config {
	config := Config{}
	if err := cleanenv.ReadEnv(&config); err != nil {
		panic(fmt.Errorf("error reading config: %w", err))
	}

	return config
}
