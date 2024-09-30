package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/converter"
	"github.com/iurikman/cashFlowManager/internal/models"
)

type Service struct {
	db                   db
	xrConverter          xrConverter
	transactionsProducer transactionsProducer
}

func NewService(db db, xrConverter xrConverter, transactionsProducer transactionsProducer) *Service {
	return &Service{
		db:                   db,
		xrConverter:          xrConverter,
		transactionsProducer: transactionsProducer,
	}
}

type xrConverter interface {
	Convert(ctx context.Context, currencyFrom, currencyTo converter.Currency) (float64, error)
}

//go:generate mockery --name transactionsProducer --exported
type transactionsProducer interface {
	ProduceTransaction(ctx context.Context, transactions models.Transaction) error
}

type db interface {
	CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id, ownerID uuid.UUID) (*models.Wallet, error)
	UpdateWallet(ctx context.Context, id, ownerID uuid.UUID, name, currency *string, balance float64) (*models.Wallet, error)
	DeleteWallet(ctx context.Context, id, ownerID uuid.UUID) error
	Withdraw(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error
	Deposit(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error
	Transfer(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error
	GetTransactions(ctx context.Context, ID uuid.UUID, params models.Params) ([]*models.Transaction, error)
}

func (s *Service) CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error) {
	createdWallet, err := s.db.CreateWallet(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("s.db.CreateWallet(ctx, wallet) err: %w", err)
	}

	return createdWallet, nil
}

func (s *Service) GetWalletByID(ctx context.Context, id, ownerID uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.db.GetWalletByID(ctx, id, ownerID)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetWalletByID(id) err: %w", err)
	}

	return wallet, nil
}

func (s *Service) UpdateWallet(ctx context.Context, id, ownerID uuid.UUID, walletDTO models.WalletDTO) (*models.Wallet, error) {
	wallet, err := s.db.GetWalletByID(ctx, id, ownerID)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetWalletByID(id) err: %w", err)
	}

	newBalance := wallet.Balance
	newCurrency := &wallet.Currency

	if walletDTO.Currency != nil {
		newCurrency = walletDTO.Currency

		if wallet.Currency != *walletDTO.Currency {
			convertedAmount, err := s.xrConverter.Convert(
				ctx,
				converter.Currency{Amount: wallet.Balance, Name: wallet.Currency},
				converter.Currency{Amount: wallet.Balance, Name: *walletDTO.Currency},
			)
			if err != nil {
				return nil, fmt.Errorf("s.xrConverter.Convert(...) err: %w", err)
			}

			newBalance = convertedAmount
		}
	}

	newName := &wallet.Name
	if walletDTO.Name != nil {
		newName = walletDTO.Name
	}

	updatedWallet, err := s.db.UpdateWallet(ctx, id, ownerID, newName, newCurrency, newBalance)
	if err != nil {
		return nil, fmt.Errorf("s.db.UpdateWallet(ctx, id, walletDTO) err: %w", err)
	}

	return updatedWallet, nil
}

func (s *Service) DeleteWallet(ctx context.Context, id, ownerID uuid.UUID) error {
	if err := s.db.DeleteWallet(ctx, id, ownerID); err != nil {
		return fmt.Errorf("s.db.DeleteWallet(id) err: %w", err)
	}

	return nil
}

//nolint:dupl
func (s *Service) Withdraw(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error {
	wallet, err := s.db.GetWalletByID(ctx, transaction.WalletID, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.GetWalletByID(walletID) err: %w", err)
	}

	if wallet.Currency != transaction.Currency {
		convertedAmount, err := s.xrConverter.Convert(
			ctx,
			converter.Currency{Amount: transaction.Amount, Name: transaction.Currency},
			converter.Currency{Amount: wallet.Balance, Name: wallet.Currency},
		)
		if err != nil {
			return fmt.Errorf("s.xrConverter.Convert(...) err: %w", err)
		}

		transaction.ConvertedAmount = convertedAmount
	}

	err = s.db.Withdraw(ctx, transaction, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.Withdraw() err: %w", err)
	}

	if err = s.transactionsProducer.ProduceTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("s.transactionsProducer.ProduceTransaction() err: %w", err)
	}

	return nil
}

//nolint:dupl
func (s *Service) Deposit(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error {
	wallet, err := s.db.GetWalletByID(ctx, transaction.WalletID, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.GetWalletByID(walletID) err: %w", err)
	}

	if wallet.Currency != transaction.Currency {
		convertedAmount, err := s.xrConverter.Convert(
			ctx,
			converter.Currency{Amount: transaction.Amount, Name: transaction.Currency},
			converter.Currency{Amount: wallet.Balance, Name: wallet.Currency},
		)
		if err != nil {
			return fmt.Errorf("s.xrConverter.Convert(...) err: %w", err)
		}

		transaction.ConvertedAmount = convertedAmount
	}

	err = s.db.Deposit(ctx, transaction, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.Deposit() err: %w", err)
	}

	if err = s.transactionsProducer.ProduceTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("s.transactionsProducer.ProduceTransaction() err: %w", err)
	}

	return nil
}

func (s *Service) Transfer(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error {
	walletFrom, err := s.GetWalletByID(ctx, transaction.WalletID, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.GetWalletByID(walletID) err: %w", err)
	}

	walletTo, err := s.GetWalletByID(ctx, transaction.TargetWalletID, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.GetWalletByID(walletID) err: %w", err)
	}

	if walletFrom.Currency != walletTo.Currency {
		convertedAmount, err := s.xrConverter.Convert(
			ctx,
			converter.Currency{Amount: transaction.Amount, Name: walletFrom.Currency},
			converter.Currency{Amount: walletTo.Balance, Name: walletTo.Currency},
		)
		if err != nil {
			return fmt.Errorf("s.xrConverter.Convert(...) err: %w", err)
		}

		transaction.ConvertedAmount = convertedAmount
	}

	err = s.db.Transfer(ctx, transaction, ownerID)
	if err != nil {
		return fmt.Errorf("s.db.Transfer() err: %w", err)
	}

	if err = s.transactionsProducer.ProduceTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("s.transactionsProducer.ProduceTransaction() err: %w", err)
	}

	return nil
}

func (s *Service) GetTransactions(ctx context.Context, id uuid.UUID, params models.Params) (
	[]*models.Transaction, error,
) {
	var transactions []*models.Transaction

	transactions, err := s.db.GetTransactions(ctx, id, params)

	switch {
	case errors.Is(err, models.ErrTransactionsNotFound):
		return nil, models.ErrTransactionsNotFound

	case err != nil:
		return nil, fmt.Errorf("s.db.GetTransactions(walletID) err: %w", err)
	}

	return transactions, nil
}
