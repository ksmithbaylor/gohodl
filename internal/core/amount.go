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

	return NewAmountFromDecimal(asset, value)
}

func NewAmountFromAtomicString(asset Asset, atomicString string) (Amount, error) {
	value, err := decimal.NewFromString(atomicString)
	if err != nil {
		return Amount{}, fmt.Errorf("Not a valid amount ('%s') of %s: %w", atomicString, asset, err)
	}

	if strings.Contains(atomicString, ".") || value.Exponent() != 0 {
		return Amount{}, fmt.Errorf("Fractional atomic units ('%s') of %s not allowed", atomicString, asset)
	}

	if value.LessThan(decimal.Zero) {
		return Amount{}, fmt.Errorf("Negative amount ('%s') of %s not allowed", atomicString, asset)
	}

	return Amount{
		Value: value.Shift(-int32(asset.Decimals)),
		Asset: asset,
	}, nil
}

func NewAmountFromDecimal(asset Asset, decimalValue decimal.Decimal) (Amount, error) {
	if decimalValue.LessThan(decimal.Zero) {
		return Amount{}, fmt.Errorf("Negative amount ('%s') of %s not allowed", decimalValue.String(), asset)
	}

	value := decimalValue.Truncate(int32(asset.Decimals))

	return Amount{
		Value: value,
		Asset: asset,
	}, nil
}

func NewAmountFromAtomicValue(asset Asset, atomicValue uint64) Amount {
	value := decimal.New(int64(atomicValue), -int32(asset.Decimals))

	return Amount{
		Value: value,
		Asset: asset,
	}

}

func NewAmountFromAtomicDecimal(asset Asset, atomicDecimal decimal.Decimal) Amount {
	value := atomicDecimal.Truncate(0).Shift(-int32(asset.Decimals))

	return Amount{
		Value: value,
		Asset: asset,
	}
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.Value.StringFixed(int32(a.Asset.Decimals)), a.Asset)
}

func (a Amount) Add(other Amount) (Amount, error) {
	if a.Asset != other.Asset {
		return a, fmt.Errorf("Cannot add %s amount and %s amount", a.Asset, other.Asset)
	}

	return Amount{
		Value: a.Value.Add(other.Value),
		Asset: a.Asset,
	}, nil
}

func (a Amount) Sub(other Amount) (Amount, error) {
	if a.Asset != other.Asset {
		return a, fmt.Errorf("Cannot subtract %s amount and %s amount", a.Asset, other.Asset)
	}

	return Amount{
		Value: a.Value.Sub(other.Value),
		Asset: a.Asset,
	}, nil
}

func (a Amount) Neg() Amount {
	return Amount{
		Value: a.Value.Neg(),
		Asset: a.Asset,
	}
}
