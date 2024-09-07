package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const UserInfoKey ctxKey = "userInfo"

type UserRegisterData struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (u UserRegisterData) Validate() error {
	switch {
	case u.Phone == "":
		return ErrPhoneIsRequired
	case u.Email == "":
		return ErrEmailIsRequired
	case u.Password == "":
		return ErrPasswordIsRequired
	case u.Name == "":
		return ErrNameIsRequired
	}

	return nil
}

type UserLoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ctxKey string

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	Owner     uuid.UUID `json:"owner"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Deleted   bool      `json:"deleted"`
}

type Transaction struct {
	TransactionID   uuid.UUID `json:"id"`
	WalletID        uuid.UUID `json:"walletId"`
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

	if _, ok := AllowedCurrencies[t.Currency]; !ok {
		return ErrCurrencyNotAllowed
	}

	if t.Amount <= 0 {
		return ErrAmountIsZero
	}

	if t.OperationType == "" {
		return ErrTransactionTypeIsEmpty
	}

	return nil
}

func (w Wallet) Validate() error {
	if _, ok := AllowedCurrencies[w.Currency]; !ok {
		return ErrCurrencyNotAllowed
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
var AllowedCurrencies = map[string]string{
	"RUR": "",
	"CHY": "R01375",
	"AED": "R01230",
	"INR": "R01270",
}

type Claims struct {
	jwt.RegisteredClaims
	UUID uuid.UUID `json:"uuid"`
}
