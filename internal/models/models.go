package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	Owner     uuid.UUID `json:"owner"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Deleted   bool      `json:"deleted"`
}

type WalletDTO struct {
	Owner    uuid.UUID `json:"owner"`
	Currency string    `json:"currency"`
	Balance  float64   `json:"balance"`
	Deleted  bool      `json:"deleted"`
}

func (w Wallet) Validate() error {
	if w.Currency == "" {
		return ErrCurrencyIsEmpty
	}

	if w.Balance < 0 {
		return ErrBalanceBelowZero
	}

	if w.Owner == uuid.Nil {
		return ErrOwnerIsEmpty
	}

	return nil
}

func (w WalletDTO) ValidateWalletDTO() error {
	if w.Currency == "" {
		return ErrCurrencyIsEmpty
	}

	if w.Balance < 0 {
		return ErrBalanceBelowZero
	}

	if w.Owner == uuid.Nil {
		return ErrOwnerIsEmpty
	}

	return nil
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Wallets   []Wallet  `json:"wallets"`
	CreatedAt time.Time `json:"createdAt"`
	Deleted   bool      `json:"deleted"`
}
