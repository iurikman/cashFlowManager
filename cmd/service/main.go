package main

import (
	"cashFlowManager/internal/config"
	"cashFlowManager/internal/kafka"
	"cashFlowManager/internal/rest"
	"cashFlowManager/internal/service"
	"cashFlowManager/internal/store"
	"context"
	"crypto/rsa"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	cfg := config.NewConfig()

	db, err := store.NewPostgres(ctx, store.Config{
		PGUser:     cfg.PostgresUser,
		PGPassword: cfg.PostgresPassword,
		PGHost:     cfg.PostgresHost,
		PGPort:     cfg.PostgresPort,
		PGDatabase: cfg.PostgresDatabase,
	})
	if err != nil {
		log.Panicf("store.NewPostgres(context.Background(), store.Config{...} err: %v", err)
	}

	kafkaConfig := kafka.Config{
		Address:  cfg.KafkaAddress,
		Topic:    cfg.KafkaTopic,
		Balancer: cfg.KafkaBalancer,
		GroupID:  cfg.KafkaGroupID,
		Brokers:  cfg.KafkaBrokers,
	}

	kafkaInit, err := kafka.NewKafka(kafkaConfig)
	if err != nil {
		log.Panicf("kafka2.NewKafka(kafka2.Config{} err: %v", err)
	}

	svc := service.NewService(db, kafkaInit)
	key := rsa.PublicKey{
		N: nil,
		E: 0,
	}
	srv := rest.NewServer(rest.Config{BindAddress: cfg.BindAddress}, svc, &key)

	err = srv.Start(context.Background())
	if err != nil {
		log.Panicf("start server error: %v", err)
	}
}
