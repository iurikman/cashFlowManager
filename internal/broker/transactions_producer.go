package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

const transactionsTopic = "transactions"

type TransactionsProducer struct {
	kafkaWriter *kafka.Writer
}

func NewTransactionsProducer() *TransactionsProducer {
	cfg := config.NewConfig()

	address, err := net.ResolveTCPAddr("tcp", cfg.KafkaAddress)
	if err != nil {
		log.Warnf("Could not resolve Kafka address: %s", err)

		return nil
	}

	return &TransactionsProducer{kafkaWriter: &kafka.Writer{
		Addr:         address,
		Topic:        transactionsTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
		Async:        true,
	}}
}

func (p *TransactionsProducer) ProduceTransaction(ctx context.Context, transaction models.Transaction) error {
	key, err := json.Marshal(transaction.TransactionID)
	if err != nil {
		return fmt.Errorf("could not marshal transaction id: %w", err)
	}

	payload, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("could not marshal transaction: %w", err)
	}

	if err = p.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: payload,
	}); err != nil {
		return fmt.Errorf("could not write messages: %w", err)
	}

	log.Infof("transaction # %s produced", transaction.TransactionID)

	if err = p.kafkaWriter.Close(); err != nil {
		return fmt.Errorf("could not close kafkaWriter: %w", err)
	}

	return nil
}
