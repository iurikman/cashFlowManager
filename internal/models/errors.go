package models

import "errors"

var (
	ErrDuplicateWallet        = errors.New("duplicate wallet")
	ErrBalanceBelowZero       = errors.New("balance is below zero")
	ErrOwnerIsEmpty           = errors.New("owner is empty")
	ErrWalletNotFound         = errors.New("wallet not found")
	ErrWalletIDIsEmpty        = errors.New("wallet ID is empty")
	ErrAmountIsZero           = errors.New("amount is zero")
	ErrTransactionTypeIsEmpty = errors.New("transaction type is empty")
	ErrChangeBalanceData      = errors.New("change balance data is wrong")
	ErrCurrencyNotAllowed     = errors.New("currency not allowed")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidAccessToken     = errors.New("invalid access token")
	ErrHeaderIsEmpty          = errors.New("header is empty")
	ErrDuplicateEmail         = errors.New("duplicate email")
	ErrPhoneIsRequired        = errors.New("phone is required")
	ErrEmailIsRequired        = errors.New("email is required")
	ErrPasswordIsRequired     = errors.New("password is required")
	ErrNameIsRequired         = errors.New("name is required")
)
