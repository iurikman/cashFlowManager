package service

import (
	"cashFlowManager/internal/models"
	"context"
)

type db interface {
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
}

type kafka interface {
	CreateWallet(ctx context.Context, wallet *models.Wallet) error
}

type Service struct {
	db    db
	kafka kafka
}

func (s Service) CreateWallet(context context.Context, wallet models.Wallet) {
	// TODO implement me
	panic("implement me")
}

func NewService(db db, kafka kafka) *Service {
	return &Service{
		db:    db,
		kafka: kafka,
	}
}
