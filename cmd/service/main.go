package main

import (
	"cashFlowManager/internal/config"
	"cashFlowManager/internal/rest"
	"cashFlowManager/internal/service"
	"context"
	"log"
)

func main() {
	cfg := config.NewConfig()
	svc := service.NewService()
	srv := rest.NewServer(rest.Config{BindAddress: cfg.BindAddress}, svc)

	err := srv.Start(context.Background())
	if err != nil {
		log.Panicf("start server error: %v", err)
	}
}
