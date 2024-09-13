package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type Producer struct {
	kafkaWriter *kafka.Writer
}

func NewProducer() *Producer {
	cfg := config.NewConfig()

	address, err := net.ResolveTCPAddr("tcp", cfg.KafkaAddress)
	if err != nil {
		log.Warnf("Failed to resolve Kafka address: %s", err)

		return nil
	}

	return &Producer{kafkaWriter: &kafka.Writer{
		Addr:         address,
		Topic:        userUpdatesTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
		Async:        true,
	}}
}

func (p *Producer) Start(ctx context.Context) error {
	var user models.User

	for i := 1; i < 11; i++ {
		user.ID = uuid.New()
		user.Username = "User " + strconv.Itoa(i)
		user.Email = "email" + strconv.Itoa(i)
		user.Phone = "phone" + strconv.Itoa(i)
		user.Password = "userPassword" + strconv.Itoa(i)
		user.Wallets = nil
		user.CreatedAt = time.Now()
		user.Deleted = false

		key, err := json.Marshal(user.ID)
		if err != nil {
			return fmt.Errorf("failed to marshal user key: %w", err)
		}

		payload, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("could not marshal user: %w", err)
		}

		log.Printf("Writed message payload: %s", string(payload))

		if err := p.kafkaWriter.WriteMessages(
			ctx,
			kafka.Message{Value: payload, Key: key}); err != nil {
			return fmt.Errorf("could not write messages: %w", err)
		}
	}

	return nil
}

func (p *Producer) Stop() error {
	if err := p.kafkaWriter.Close(); err != nil {
		log.Warnf("failed to close kafka writer: %v", err)

		return nil
	}

	log.Println("producer stopped")

	return nil
}
