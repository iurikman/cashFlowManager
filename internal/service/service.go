package service

import (
	"context"

	"github.com/iurikman/cashFlowManager/internal/models"
)

type Service struct {
	db db
}

func NewService(db db) *Service {
	return &Service{
		db: db,
	}
}

type db interface {
	UpsertUser(ctx context.Context, user models.User) error
}
