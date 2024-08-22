package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/iurikman/cashFlowManager/internal/broker"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	producer := broker.NewProducer()
	if err := producer.Start(ctx); err != nil {
		log.Panicf("failed to start producer: %v", err)
	}

	<-ctx.Done()

	if err := producer.Stop(); err != nil {
		log.Panicf("failed to stop producer: %v", err)
	}
}
