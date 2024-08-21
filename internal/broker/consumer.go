package broker

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/iurikman/cashFlowManager/internal/store"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

const (
	userUpdatesTopic = "users_updates"
)

type Consumer struct {
	reader *kafka.Reader
	db     *store.Postgres
}

func NewConsumer(db *store.Postgres) *Consumer {
	cfg := config.NewConfig()

	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.KafkaBrokers,
			Topic:   userUpdatesTopic,
			GroupID: cfg.KafkaGroupID,
		}),
		db: db,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	var user models.User

	for {
		select {
		case <-ctx.Done():
			log.Warn("StartConsumer(ctx context.Context) case <-ctx.Done()")

			return nil
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Warnf("failed to read message from Kafka (c.reader.ReadMessage(ctx) err): %s", err)

				continue
			}

			log.Infof("received message: Topic: %s \n message: %s \n key = %s \n",
				message.Topic,
				string(message.Value),
				string(message.Key),
			)

			if err = json.NewDecoder(bytes.NewReader(message.Value)).Decode(&user); err != nil {
				log.Warnf("failed to decode message from Kafka: %s", err)

				continue
			}

			if err = c.db.UpsertUser(ctx, user); err != nil {
				log.Warnf("c.Db.UpsertUser(...) err: %s", err)
			}
		}
	}
}
