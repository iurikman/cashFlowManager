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

	query := `INSERT INTO wallets (id, owner, name, currency, balance, created_at, updated_at, deleted) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id, owner, name, currency, balance, created_at, updated_at, deleted
				`

	err := p.db.QueryRow(
		ctx,
		query,
		uuid.New(),
		wallet.Owner,
		wallet.Name,
		wallet.Currency,
		0,
		timeNow,
		timeNow,
		wallet.Deleted,
	).Scan(
		&createdWallet.ID,
		&createdWallet.Owner,
		&createdWallet.Name,
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

	query := `	SELECT id, owner, name, currency, balance, created_at, updated_at, deleted 
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
		&wallet.Name,
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
	query := `	UPDATE wallets SET balance = balance + $3, updated_at = $4
                WHERE id = $1 and owner = $2 and deleted = false 
				RETURNING id, balance
				`

	_, err := tx.Exec(
		ctx,
		query,
		walletID,
		ownerID,
		amount,
		time.Now(),
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

func (p *Postgres) UpdateWallet(ctx context.Context, id, ownerID uuid.UUID, name, currency *string, balance float64) (*models.Wallet, error) {
	var updatedWallet models.Wallet

	query := `UPDATE wallets SET name = $3, currency = $4, balance = $5, updated_at = $6 
               WHERE id = $1 AND owner = $2 AND deleted = false
				RETURNING id, owner, name, currency, balance, created_at, updated_at, deleted
               `

	err := p.db.QueryRow(
		ctx,
		query,
		id,
		ownerID,
		name,
		currency,
		balance,
		time.Now(),
	).Scan(
		&updatedWallet.ID,
		&updatedWallet.Owner,
		&updatedWallet.Name,
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
