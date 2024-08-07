package kafka

import (
	"cashFlowManager/internal/models"
	"context"
	"net"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Address  net.Addr
	Topic    string
	Balancer kafka.Balancer
	GroupID  string
	Brokers  []string
}

type Kafka struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func (k *Kafka) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	// TODO implement me
	panic("implement me")
}

func NewKafka(config Config) (*Kafka, error) {
	writer := &kafka.Writer{
		Addr:     config.Address,
		Topic:    config.Topic,
		Balancer: config.Balancer,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		GroupID: config.GroupID,
		Topic:   config.Topic,
	})

	return &Kafka{writer: writer, reader: reader}, nil
}
