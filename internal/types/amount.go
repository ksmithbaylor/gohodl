package types

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Amount struct {
	Value    decimal.Decimal
	Currency string
}

func NewAmount(strValue, currency string) (Amount, error) {
	value, err := decimal.NewFromString(strValue)
	if err != nil {
		return Amount{}, fmt.Errorf("Not a valid amount ('%s'): %w", value, err)
	}

	return Amount{
		Value:    value,
		Currency: currency,
	}, nil
}

func NewAmountWithDecimals(strValue string, decimals int, currency string) (Amount, error) {
	value, err := decimal.NewFromString(strValue)
	if err != nil {
		return Amount{}, fmt.Errorf("Not a valid amount ('%s'): %w", value, err)
	}

	return Amount{
		Value:    value.Shift(-int32(decimals)),
		Currency: currency,
	}, nil
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.Value.String(), a.Currency)
}

func (a Amount) Add(other Amount) (Amount, error) {
	if a.Currency != other.Currency {
		return a, fmt.Errorf("Cannot add %s amount and %s amount", a.Currency, other.Currency)
	}

	return Amount{
		Value:    a.Value.Add(other.Value),
		Currency: a.Currency,
	}, nil
}

func (a Amount) Sub(other Amount) (Amount, error) {
	if a.Currency != other.Currency {
		return a, fmt.Errorf("Cannot subtract %s amount and %s amount", a.Currency, other.Currency)
	}

	return Amount{
		Value:    a.Value.Add(other.Value),
		Currency: a.Currency,
	}, nil
}
