package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const UserInfoKey ctxKey = "userInfo"

type ctxKey string

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	Owner     uuid.UUID `json:"owner"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Deleted   bool      `json:"deleted"`
}

func (w Wallet) Validate() error {
	if _, ok := allowedCurrencies[w.Currency]; !ok {
		return ErrCurrencyNotAllowed
	}

	if w.Owner == uuid.Nil {
		return ErrOwnerIsEmpty
	}

	if w.Name == "" {
		return ErrNameIsRequired
	}

	return nil
}

type WalletDTO struct {
	Name     *string `json:"name,omitempty"`
	Currency *string `json:"currency,omitempty"`
}

func (w WalletDTO) Validate() error {
	if w.Name != nil {
		if *w.Name == "" {
			return ErrNameIsEmpty
		}
	}

	if w.Currency != nil {
		if *w.Currency == "" {
			return ErrCurrencyIsEmpty
		}

		if _, ok := allowedCurrencies[*w.Currency]; !ok {
			return ErrCurrencyNotAllowed
		}
	}

	return nil
}

type Transaction struct {
	TransactionID   uuid.UUID `json:"id"`
	WalletID        uuid.UUID `json:"walletId"`
	OwnerID         uuid.UUID `json:"ownerId"`
	TargetWalletID  uuid.UUID `json:"targetWalletId"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ConvertedAmount float64   `json:"convertedAmount"`
	ExRate          float64   `json:"exRate"`
	OperationType   string    `json:"transactionType"`
	ExecutedAt      time.Time `json:"executedAt"`
}

func (t Transaction) Validate() error {
	if t.WalletID == uuid.Nil {
		return ErrWalletIDIsEmpty
	}

	if _, ok := allowedCurrencies[t.Currency]; !ok {
		return ErrCurrencyNotAllowed
	}

	if _, ok := allowedOperationTypes[t.OperationType]; !ok {
		return ErrOperationTypeNotAllowed
	}

	if t.Amount <= 0 {
		return ErrAmountIsZero
	}

	if t.OperationType == "" {
		return ErrTransactionTypeIsEmpty
	}

	return nil
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Password  string    `json:"password"`
	Wallets   []Wallet  `json:"wallets"`
	CreatedAt time.Time `json:"createdAt"`
	Deleted   bool      `json:"deleted"`
}

type UserInfo struct {
	ID uuid.UUID
}

//nolint:gochecknoglobals
var allowedOperationTypes = map[string]struct{}{
	"deposit":  {},
	"transfer": {},
	"withdraw": {},
}

//nolint:gochecknoglobals
var allowedCurrencies = map[string]string{
	"RUR": "",
	"CHY": "R01375",
	"AED": "R01230",
	"INR": "R01270",
}

func GetCurrencyCode(code string) (*string, error) {
	if code, ok := allowedCurrencies[code]; ok {
		return &code, nil
	}

	return nil, ErrCurrencyNotAllowed
}

type Claims struct {
	jwt.RegisteredClaims
	UUID uuid.UUID `json:"uuid"`
}

type Params struct {
	Offset         int    `schema:"offset,omitempty"`
	Limit          int    `schema:"limit,omitempty"`
	Sorting        string `schema:"sorting,omitempty"`
	Descending     bool   `schema:"descending,omitempty"`
	FilterDateFrom string `schema:"filterFrom,omitempty"`
	FilterDateTo   string `schema:"filterTo,omitempty"`
	FilterType     string `schema:"filterCurrency,omitempty"`
}
