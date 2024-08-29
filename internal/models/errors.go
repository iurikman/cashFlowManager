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
)
