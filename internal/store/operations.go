package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
)

func (p *Postgres) Deposit(
	ctx context.Context, changeBalanceData models.Transaction,
) (*models.Transaction, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("p.db.Begin(ctx) err: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warnf("deposit tx.Rollback(ctx) err: %v", err)
		}
	}()

	err = p.updateWalletBalance(ctx, tx, changeBalanceData.WalletID, models.WalletDTO{Balance: changeBalanceData.Amount})
	if err != nil {
		return nil, models.ErrChangeBalanceData
	}

	transaction, err := saveTransaction(ctx, tx, changeBalanceData)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit err: %w", err)
	}

	return transaction, nil
}

func (p *Postgres) Transfer(
	ctx context.Context, changeBalanceData models.Transaction,
) (*models.Transaction, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("p.db.Begin(ctx) err: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warnf("transfer tx.Rollback(ctx) err: %v", err)
		}
	}()

	err = p.updateWalletBalance(
		ctx,
		tx,
		changeBalanceData.WalletID,
		models.WalletDTO{Balance: -changeBalanceData.Amount})

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		return nil, models.ErrWalletNotFound
	case err != nil:
		return nil, fmt.Errorf("owner walletp.db.UpdateWallet(ctx) err: %w", err)
	}

	err = p.updateWalletBalance(
		ctx,
		tx,
		changeBalanceData.TargetWalletID,
		models.WalletDTO{Balance: changeBalanceData.Amount})

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		return nil, models.ErrWalletNotFound
	case err != nil:
		return nil, fmt.Errorf("target wallet p.db.UpdateWallet(ctx) err: %w", err)
	}

	transaction, err := saveTransaction(ctx, tx, changeBalanceData)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit err: %w", err)
	}

	return transaction, nil
}

func (p *Postgres) Withdraw(
	ctx context.Context, changeBalanceData models.Transaction,
) (*models.Transaction, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("p.db.Begin(ctx) err: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Warnf("withdraw tx.Rollback(ctx) err: %v", err)
		}
	}()

	if err = p.updateWalletBalance(
		ctx,
		tx,
		changeBalanceData.WalletID,
		models.WalletDTO{Balance: -changeBalanceData.Amount}); err != nil {
		return nil, models.ErrChangeBalanceData
	}

	transaction, err := saveTransaction(ctx, tx, changeBalanceData)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit err: %w", err)
	}

	return transaction, nil
}

func saveTransaction(ctx context.Context, tx pgx.Tx, transaction models.Transaction) (*models.Transaction, error) {
	var executedOperation models.Transaction

	query := `INSERT INTO transactions_history
    (id, wallet_id, target_wallet_id, amount, currency, transaction_type, executed_at)
    		VALUES ($1, $2, $3, $4, $5, $6, $7)
           	RETURNING 
           	    id, wallet_id, target_wallet_id, amount, currency, transaction_type, executed_at`

	err := tx.QueryRow(
		ctx,
		query,
		transaction.TransactionID,
		transaction.WalletID,
		transaction.TargetWalletID,
		transaction.Amount,
		transaction.Currency,
		transaction.OperationType,
		time.Now(),
	).Scan(
		&executedOperation.TransactionID,
		&executedOperation.WalletID,
		&executedOperation.TargetWalletID,
		&executedOperation.Amount,
		&executedOperation.Currency,
		&executedOperation.OperationType,
		&executedOperation.ExecutedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("transaction writing to base err: %w", err)
	}

	return &executedOperation, nil
}
