//nolint
package tests

import (
	"context"

	"github.com/iurikman/cashFlowManager/internal/converter"
)

var AllowedCurrencies = map[string]float64{
	"RUR": 1,
	"CHY": 12,
	"AED": 24,
	"INR": 2,
}

type MockConverter struct{}

func (c MockConverter) Convert(ctx context.Context, currencyFrom converter.Currency, currencyTo converter.Currency) (float64, error) {
	changeRateCurrFrom := AllowedCurrencies[currencyFrom.Name]
	changeRateCurrTo := AllowedCurrencies[currencyTo.Name]

	switch {
	case currencyTo.Name == "RUR":
		result := currencyFrom.Amount * changeRateCurrFrom

		return result, nil
	default:
		result := (currencyFrom.Amount * changeRateCurrFrom) / changeRateCurrTo

		return result, nil
	}
}
