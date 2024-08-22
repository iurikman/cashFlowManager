package store

import (
	"context"
	"fmt"
	"time"

	"github.com/iurikman/cashFlowManager/internal/models"
)

func (p *Postgres) UpsertUser(ctx context.Context, user models.User) error {
	query := `	INSERT INTO users (id, name, created_at, deleted) VALUES ($1, $2, $3, $4) 
				ON CONFLICT (id) DO UPDATE 
				SET name = $2, deleted = $4;
				`

	_, err := p.db.Exec(
		ctx,
		query,
		user.ID,
		user.Username,
		time.Now(),
		user.Deleted,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
