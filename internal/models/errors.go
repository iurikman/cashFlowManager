package models

import "errors"

var (
	ErrDuplicateWallet         = errors.New("duplicate wallet")
	ErrBalanceBelowZero        = errors.New("balance is below zero")
	ErrOwnerIsEmpty            = errors.New("owner is empty")
	ErrWalletNotFound          = errors.New("wallet not found")
	ErrWalletIDIsEmpty         = errors.New("wallet ID is empty")
	ErrAmountIsZero            = errors.New("amount is zero")
	ErrTransactionTypeIsEmpty  = errors.New("transaction type is empty")
	ErrChangeBalanceData       = errors.New("change balance data is wrong")
	ErrCurrencyNotAllowed      = errors.New("currency not allowed")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidAccessToken      = errors.New("invalid access token")
	ErrHeaderIsEmpty           = errors.New("header is empty")
	ErrNameIsRequired          = errors.New("name is required")
	ErrTransactionsNotFound    = errors.New("transactions not found")
	ErrOperationTypeNotAllowed = errors.New("operation type not allowed")
	ErrNameIsEmpty             = errors.New("name is empty")
	ErrCurrencyIsEmpty         = errors.New("currency is empty")
	ErrBalanceIsEmpty          = errors.New("balance is empty")
)
