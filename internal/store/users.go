package store

import (
	"context"
	"fmt"
	"time"

	"github.com/iurikman/cashFlowManager/internal/models"
)

func (p *Postgres) UpsertUser(ctx context.Context, user models.User) error {
	query := `	INSERT INTO users (id, name, email, phone, password, created_at, updated_at, deleted) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
				ON CONFLICT (id) DO UPDATE 
				SET name = $2, email = $3, phone = $4, password = $5, updated_at = $7, deleted = $8;
				`

	_, err := p.db.Exec(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Phone,
		user.Password,
		time.Now(),
		time.Now(),
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
