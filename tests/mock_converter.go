package tests

import (
	"context"
	"github.com/iurikman/cashFlowManager/internal/converter"
)

type MockConverter struct{}

func (c MockConverter) Convert(ctx context.Context, currencyFrom, currencyTo converter.Currency) (float64, error) {
	if currencyFrom == currencyTo {
		return 1, nil
	}

	return 1.5, nil
}
