package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iurikman/cashFlowManager/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *Postgres) CreateWallet(ctx context.Context, wallet models.Wallet) (*models.Wallet, error) {
	createdWallet := new(models.Wallet)

	query := `INSERT INTO wallets (id, owner, currency, balance, created_at, deleted) 
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id, owner, currency, balance, created_at, deleted
				`

	err := p.db.QueryRow(
		ctx,
		query,
		wallet.ID,
		wallet.Owner,
		wallet.Currency,
		wallet.Balance,
		wallet.CreatedAt,
		wallet.Deleted,
	).Scan(
		&createdWallet.ID,
		&createdWallet.Owner,
		&createdWallet.Currency,
		&createdWallet.Balance,
		&createdWallet.CreatedAt,
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
