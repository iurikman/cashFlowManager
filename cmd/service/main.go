package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/iurikman/cashFlowManager/internal/broker"
	"github.com/iurikman/cashFlowManager/internal/config"
	"github.com/iurikman/cashFlowManager/internal/converter"
	"github.com/iurikman/cashFlowManager/internal/jwtgenerator"
	"github.com/iurikman/cashFlowManager/internal/rest"
	"github.com/iurikman/cashFlowManager/internal/service"
	"github.com/iurikman/cashFlowManager/internal/store"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
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

	transactionsProducer := broker.NewTransactionsProducer()

	svc := service.NewService(db, xrConverter, transactionsProducer)

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
		if err := consumer.Start(ctx); err != nil {
			return fmt.Errorf("consumer stopped: %w", err)
		}

		return nil
	})
	log.Info("consumer started")

	eg.Go(func() error {
		if err := svc.StartCleaner(ctx); err != nil {
			return fmt.Errorf("cleaner stopped: %w", err)
		}

		return nil
	})
	log.Info("cleaner started")

	eg.Go(func() error {
		if err := srv.Start(ctx); err != nil {
			return fmt.Errorf("server stopped: %w", err)
		}

		return nil
	})
	log.Info("service started")

	if err := eg.Wait(); err != nil {
		log.Panicf("eg.Wait() err: %v", err)
	}
}
