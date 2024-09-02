package service

import (
	"context"
	"fmt"
	"github.com/iurikman/cashFlowManager/internal/converter"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
)

type Service struct {
	db        db
	converter xrConverter
}

func NewService(db db, converter xrConverter) *Service {
	return &Service{
		db:        db,
		converter: converter,
	}
}

type db interface {
	CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	UpdateWallet(ctx context.Context, id uuid.UUID, dto models.WalletDTO) (*models.Wallet, error)
	DeleteWallet(ctx context.Context, id uuid.UUID) error
	Withdraw(ctx context.Context, transaction models.Transaction) (*models.Transaction, error)
	Deposit(ctx context.Context, transaction models.Transaction) (*models.Transaction, error)
	Transfer(ctx context.Context, transaction models.Transaction) (*models.Transaction, error)
}

type xrConverter interface {
	Convert(ctx context.Context, currencyFrom, currencyTo converter.Currency) (float64, error)
}

func (s *Service) CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error) {
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

func (s *Service) UpdateWallet(ctx context.Context, id uuid.UUID, walletDTO models.WalletDTO) (*models.Wallet, error) {
	updatedWallet, err := s.db.UpdateWallet(ctx, id, walletDTO)
	if err != nil {
		return nil, fmt.Errorf("s.db.UpdateWallet(id, walletDTO) err: %w", err)
	}

	return updatedWallet, nil
}

func (s *Service) DeleteWallet(ctx context.Context, id uuid.UUID) error {
	if err := s.db.DeleteWallet(ctx, id); err != nil {
		return fmt.Errorf("s.db.DeleteWallet(id) err: %w", err)
	}

	return nil
}

func (s *Service) Withdraw(ctx context.Context, transaction models.Transaction) (*models.Transaction, error) {
	executedTransaction, err := s.db.Withdraw(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("s.db.Withdraw() err: %w", err)
	}

	return executedTransaction, nil
}

func (s *Service) Deposit(ctx context.Context, transaction models.Transaction) (*models.Transaction, error) {
	executedTransaction, err := s.db.Deposit(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("s.db.Deposit() err: %w", err)
	}

	return executedTransaction, nil
}

func (s *Service) Transfer(ctx context.Context, transaction models.Transaction) (*models.Transaction, error) {
	executedTransaction, err := s.db.Transfer(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("s.db.Transfer() err: %w", err)
	}

	return executedTransaction, nil
}
