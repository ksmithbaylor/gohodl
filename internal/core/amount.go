package core

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Amount struct {
	Value decimal.Decimal
	Asset Asset
}

func NewAmountFromDecimalString(asset Asset, decimalString string) (Amount, error) {
	value, err := decimal.NewFromString(decimalString)
	if err != nil {
		return Amount{}, fmt.Errorf("Not a valid amount ('%s') of %s: %w", decimalString, asset, err)
	}

	if value.LessThan(decimal.Zero) {
		return Amount{}, fmt.Errorf("Negative amount ('%s') of %s not allowed", decimalString, asset)
	}

	return Amount{
		Value: value.Truncate(int32(asset.Decimals)),
		Asset: asset,
	}, nil
}

func NewAmountFromCentsString(asset Asset, centsString string) (Amount, error) {
	value, err := decimal.NewFromString(centsString)
	if err != nil {
		return Amount{}, fmt.Errorf("Not a valid amount ('%s') of %s: %w", centsString, asset, err)
	}

	if strings.Contains(centsString, ".") || value.Exponent() != 0 {
		return Amount{}, fmt.Errorf("Fractional cents ('%s') of %s not allowed", centsString, asset)
	}

	if value.LessThan(decimal.Zero) {
		return Amount{}, fmt.Errorf("Negative amount ('%s') of %s not allowed", centsString, asset)
	}

	return Amount{
		Value: value.Shift(-int32(asset.Decimals)),
		Asset: asset,
	}, nil
}

func NewAmountFromCents(asset Asset, cents uint64) Amount {
	value := decimal.New(int64(cents), -int32(asset.Decimals))

	return Amount{
		Value: value,
		Asset: asset,
	}
}

func NewAmountFromCentsDecimal(asset Asset, cents decimal.Decimal) Amount {
	value := cents.Shift(-int32(asset.Decimals))

	return Amount{
		Value: value,
		Asset: asset,
	}
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.Value.StringFixed(int32(a.Asset.Decimals)), a.Asset)
}

//
// func (a Amount) Add(other Amount) (Amount, error) {
//   if a.Currency != other.Currency {
//     return a, fmt.Errorf("Cannot add %s amount and %s amount", a.Currency, other.Currency)
//   }
//
//   return Amount{
//     Value:    a.Value.Add(other.Value),
//     Currency: a.Currency,
//   }, nil
// }
//
// func (a Amount) Sub(other Amount) (Amount, error) {
//   if a.Currency != other.Currency {
//     return a, fmt.Errorf("Cannot subtract %s amount and %s amount", a.Currency, other.Currency)
//   }
//
//   return Amount{
//     Value:    a.Value.Add(other.Value),
//     Currency: a.Currency,
//   }, nil
// }
