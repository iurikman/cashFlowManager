package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/iurikman/cashFlowManager/internal/broker"
	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/service"
	"github.com/iurikman/cashFlowManager/internal/store"

	log "github.com/sirupsen/logrus"

	"github.com/iurikman/cashFlowManager/internal/converter"
	"github.com/iurikman/cashFlowManager/internal/jwtgenerator"
	"github.com/iurikman/cashFlowManager/internal/rest"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"golang.org/x/sync/errgroup"
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

	xrConverter := converter.NewConverter(cfg.XRConverterHost)

	svc := service.NewService(db, xrConverter)

	jwtGenerator := jwtgenerator.NewJWTGenerator()

	srv, err := rest.NewServer(
		rest.ServerConfig{BindAddress: cfg.BindAddress},
		svc,
		jwtGenerator.GetPublicKey(),
	)
	if err != nil {
		log.Panicf("rest.NewServer(cfg) err: %v", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	consumer := broker.NewConsumer(
		db,
		broker.ConsumerConfig{
			KafkaBrokers: cfg.KafkaBrokers,
			KafkaGroupID: cfg.KafkaGroupID,
		})

	eg.Go(func() error {
		err = consumer.Start(ctx)

		return fmt.Errorf("consumer stopped: %w", err)
	})
	log.Info("consumer started")

	eg.Go(func() error {
		err = srv.Start(ctx)

		return fmt.Errorf("service stopped: %w", err)
	})
	log.Info("service started")

	if err := eg.Wait(); err != nil {
		log.Panicf("eg.Wait() err: %v", err)
	}
}
