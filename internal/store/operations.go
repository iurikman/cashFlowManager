package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

func (p *Postgres) Deposit(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("p.db.Begin(ctx) err: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warnf("deposit tx.Rollback(ctx) err: %v", err)
		}
	}()

	err = p.updateWalletBalance(ctx, tx, transaction.WalletID, ownerID, transaction.Amount)
	if err != nil {
		return models.ErrChangeBalanceData
	}

	err = saveTransaction(ctx, tx, transaction)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit err: %w", err)
	}

	return nil
}

func (p *Postgres) Transfer(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("p.db.Begin(ctx) err: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warnf("transfer tx.Rollback(ctx) err: %v", err)
		}
	}()

	err = p.updateWalletBalance(ctx, tx, transaction.WalletID, ownerID, -transaction.Amount)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		return models.ErrWalletNotFound
	case errors.Is(err, models.ErrBalanceBelowZero):
		return models.ErrBalanceBelowZero
	case err != nil:
		return fmt.Errorf("owner walletp.db.UpdateWallet(ctx) err: %w", err)
	}

	err = p.updateWalletBalance(ctx, tx, transaction.TargetWalletID, ownerID, transaction.ConvertedAmount)

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		return models.ErrWalletNotFound
	case err != nil:
		return fmt.Errorf("target wallet p.db.UpdateWallet(ctx) err: %w", err)
	}

	err = saveTransaction(ctx, tx, transaction)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit err: %w", err)
	}

	return nil
}

func (p *Postgres) Withdraw(ctx context.Context, transaction models.Transaction, ownerID uuid.UUID) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("p.db.Begin(ctx) err: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warnf("withdraw tx.Rollback(ctx) err: %v", err)
		}
	}()

	err = p.updateWalletBalance(ctx, tx, transaction.WalletID, ownerID, -transaction.Amount)

	switch {
	case errors.Is(err, models.ErrBalanceBelowZero):
		return models.ErrBalanceBelowZero
	case err != nil:
		return models.ErrChangeBalanceData
	}

	err = saveTransaction(ctx, tx, transaction)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit err: %w", err)
	}

	return nil
}

func saveTransaction(ctx context.Context, tx pgx.Tx, transaction models.Transaction) error {
	var executedOperation models.Transaction

	query := `INSERT INTO transactions_history
    (id, wallet_id, target_wallet_id, amount, converted_amount, currency, transaction_type, executed_at)
    		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
           	RETURNING 
           	    id, wallet_id, target_wallet_id, amount, converted_amount, currency, transaction_type, executed_at`

	err := tx.QueryRow(
		ctx,
		query,
		transaction.TransactionID,
		transaction.WalletID,
		transaction.TargetWalletID,
		transaction.Amount,
		transaction.ConvertedAmount,
		transaction.Currency,
		transaction.OperationType,
		time.Now(),
	).Scan(
		&executedOperation.TransactionID,
		&executedOperation.WalletID,
		&executedOperation.TargetWalletID,
		&executedOperation.Amount,
		&executedOperation.ConvertedAmount,
		&executedOperation.Currency,
		&executedOperation.OperationType,
		&executedOperation.ExecutedAt,
	)
	if err != nil {
		return fmt.Errorf("transaction writing to base err: %w", err)
	}

	return nil
}
