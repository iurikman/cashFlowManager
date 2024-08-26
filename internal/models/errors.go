package models

import "errors"

var (
	ErrDuplicateWallet  = errors.New("duplicate wallet")
	ErrCurrencyIsEmpty  = errors.New("currency is empty")
	ErrBalanceBelowZero = errors.New("balance is below zero")
	ErrOwnerIsEmpty     = errors.New("owner is empty")
	ErrWalletNotFound   = errors.New("wallet not found")
)
