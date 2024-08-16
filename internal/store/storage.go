package store

import (
	"cashFlowManager/internal/models"
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db  *pgxpool.Pool
	dsn string
}

type Config struct {
	PGUser     string
	PGPassword string
	PGHost     string
	PGPort     string
	PGDatabase string
}

func NewPostgres(ctx context.Context, config Config) (*Postgres, error) {
	urlScheme := url.URL{
		Scheme:   "Postgres",
		User:     url.UserPassword(config.PGUser, config.PGPassword),
		Host:     fmt.Sprintf("%s:%s", config.PGHost, config.PGPort),
		Path:     config.PGDatabase,
		RawQuery: (url.Values{"sslmode": []string{"disable"}}).Encode(),
	}
	dsn := urlScheme.String()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New(ctx, urlScheme.String()) err: %w", err)
	}

	return &Postgres{db: db, dsn: dsn}, nil
}

func (p *Postgres) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	return nil
}
