package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/iurikman/cashFlowManager/internal/broker"
	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/service"
	"github.com/iurikman/cashFlowManager/internal/store"

	log "github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	defer cancel()

	cfg := config.NewConfig()

	db, err := store.New(ctx, store.Config{
		PGUser:     cfg.PostgresUser,
		PGPassword: cfg.PostgresPassword,
		PGHost:     cfg.PostgresHost,
		PGPort:     cfg.PostgresPort,
		PGDatabase: cfg.PostgresDatabase,
	})
	if err != nil {
		log.Panicf("store.NewPostgres(context.Background(), store.ServerConfig{...} err: %v", err)
	}

	if err := db.Migrate(migrate.Up); err != nil {
		log.Panicf("pgStore.Migrate: %v", err)
	}

	log.Info("successful migration")

	svc := service.NewService(db)

	consumer := broker.NewConsumer(db)

	err = consumer.Start(ctx)
	if err != nil {
		log.Panicf("consumer.StartConsumer(ctx) err: %v", err)
	}

	log.Infof("sercise: %v", svc)
}
