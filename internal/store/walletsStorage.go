package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *Postgres) CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error) {
	createdWallet := new(models.Wallet)

	query := `INSERT INTO wallets (id, owner, currency, balance, created_at, updated_at, deleted) 
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, owner, currency, balance, created_at, updated_at, deleted
				`

	err := p.db.QueryRow(
		ctx,
		query,
		wallet.ID,
		wallet.Owner,
		wallet.Currency,
		wallet.Balance,
		time.Now(),
		time.Now(),
		wallet.Deleted,
	).Scan(
		&createdWallet.ID,
		&createdWallet.Owner,
		&createdWallet.Currency,
		&createdWallet.Balance,
		&createdWallet.CreatedAt,
		&createdWallet.UpdatedAt,
		&createdWallet.Deleted,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, models.ErrDuplicateWallet
		}

		return nil, fmt.Errorf("creating wallet error: %w", err)
	}

	return createdWallet, nil
}

func (p *Postgres) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet

	query := `SELECT * FROM wallets WHERE id = $1 and deleted = false`

	err := p.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&wallet.ID,
		&wallet.Owner,
		&wallet.Currency,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
		&wallet.Deleted,
	)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, models.ErrWalletNotFound
	case err != nil:
		return nil, fmt.Errorf("getting wallet by id error: %w", err)
	}

	return &wallet, nil
}

func (p *Postgres) UpdateWallet(ctx context.Context, id uuid.UUID, walletDTO models.WalletDTO) (*models.Wallet, error) {
	var updatedWallet models.Wallet

	query := `	UPDATE wallets SET owner = $2, currency = $3, balance = $4, updated_at = $5
                WHERE id = $1 and deleted = false 
				RETURNING id, owner, currency, balance, created_at, updated_at, deleted
				`

	err := p.db.QueryRow(
		ctx,
		query,
		id,
		walletDTO.Owner,
		walletDTO.Currency,
		walletDTO.Balance,
		time.Now(),
	).Scan(
		&updatedWallet.ID,
		&updatedWallet.Owner,
		&updatedWallet.Currency,
		&updatedWallet.Balance,
		&updatedWallet.CreatedAt,
		&updatedWallet.UpdatedAt,
		&updatedWallet.Deleted,
	)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, models.ErrWalletNotFound
	case err != nil:
		return nil, fmt.Errorf("updating wallet error: %w", err)
	}

	return &updatedWallet, nil
}

func (p *Postgres) DeleteWallet(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE wallets SET deleted = true WHERE id = $1 and deleted = false`

	result, err := p.db.Exec(ctx, query, id)
	if result.RowsAffected() == 0 {
		return models.ErrWalletNotFound
	}

	if err != nil {
		return fmt.Errorf("deleting wallet error: %w", err)
	}

	return nil
}
