package broker

import (
	"context"
	"encoding/json"
	"errors"
	"io"

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
				if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
					log.Warn("Consumer context canceled or reached EOF")

					return nil
				}

				log.Warnf("Failed to read message from Kafka (c.reader.ReadMessage(ctx) err): %s", err)

				continue
			}

			if len(message.Value) == 0 {
				log.Warn("Received empty message.")

				continue
			}

			if err := json.Unmarshal(message.Value, &user); err != nil && !errors.Is(err, io.EOF) {
				log.Warnf("failed to decode message from Kafka: %s", err)

				continue
			}

			log.Infof("\n Received user: \n ID: %v \n Name: %v \n Created at: %v \n Deleted: %v \n \n",
				user.ID,
				user.Username,
				user.CreatedAt,
				user.Deleted,
			)

			if err = c.db.UpsertUser(ctx, user); err != nil {
				log.Warnf("c.db.UpsertUser(...) err: %s", err)
			}
		}
	}
}
