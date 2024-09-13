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

	timeNow := time.Now()

	query := `INSERT INTO wallets (id, owner, currency, balance, created_at, updated_at, deleted) 
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, owner, currency, balance, created_at, updated_at, deleted
				`

	err := p.db.QueryRow(
		ctx,
		query,
		uuid.New(),
		wallet.Owner,
		wallet.Currency,
		wallet.Balance,
		timeNow,
		timeNow,
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

		switch {
		case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
			return nil, models.ErrDuplicateWallet
		case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation:
			return nil, models.ErrUserNotFound
		}

		return nil, fmt.Errorf("creating wallet error: %w", err)
	}

	return createdWallet, nil
}

func (p *Postgres) GetWalletByID(ctx context.Context, id, ownerID uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet

	query := `	SELECT id, owner, currency, balance, created_at, updated_at, deleted 
				FROM wallets 
				WHERE id = $1 and owner = $2 and deleted = false`

	err := p.db.QueryRow(
		ctx,
		query,
		id,
		ownerID,
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

func (p *Postgres) updateWalletBalance(
	ctx context.Context, tx pgx.Tx,
	walletID, ownerID uuid.UUID,
	amount float64,
) error {
	var updatedWallet models.Wallet

	query := `	UPDATE wallets SET balance = balance + $3, updated_at = $4
                WHERE id = $1 and owner = $2 and deleted = false 
				RETURNING id, balance
				`

	err := tx.QueryRow(
		ctx,
		query,
		walletID,
		ownerID,
		amount,
		time.Now(),
	).Scan(
		&updatedWallet.ID,
		&updatedWallet.Balance,
	)

	var pgErr *pgconn.PgError

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.ErrWalletNotFound
	case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation:
		return models.ErrBalanceBelowZero
	case err != nil:
		return fmt.Errorf("updating wallet error: %w", err)
	}

	return nil
}

func (p *Postgres) DeleteWallet(ctx context.Context, id, ownerID uuid.UUID) error {
	query := `UPDATE wallets SET deleted = true WHERE id = $1 and owner = $2 and deleted = false`

	result, err := p.db.Exec(ctx, query, id, ownerID)

	switch {
	case result.RowsAffected() == 0:
		return models.ErrWalletNotFound
	case err != nil:
		return fmt.Errorf("deleting wallet error: %w", err)
	}

	return nil
}
