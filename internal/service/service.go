package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
)

type Service struct {
	db db
}

func NewService(db db) *Service {
	return &Service{
		db: db,
	}
}

type db interface {
	CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
}

func (s *Service) CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error) {
	if err := wallet.Validate(); err != nil {
		return nil, fmt.Errorf("wallet.Validate() err: %w", err)
	}

	createdWallet, err := s.db.CreateWallet(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("s.db.CreateWallet(ctx, wallet) err: %w", err)
	}

	return createdWallet, nil
}

func (s *Service) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.db.GetWalletByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetWalletByID(id) err: %w", err)
	}

	return wallet, nil
}
