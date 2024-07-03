package core_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var eth core.Asset = core.Asset{
	NetworkKind: core.EvmNetworkKind,
	NetworkName: evm.Ethereum.String(),
	Kind:        core.EvmNative,
	Identifier:  common.Address{}.String(),
	Symbol:      "ETH",
	Decimals:    18,
}

var usdc core.Asset = core.Asset{
	NetworkKind: core.EvmNetworkKind,
	NetworkName: evm.Ethereum.String(),
	Kind:        core.Erc20Token,
	Identifier:  common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48").String(),
	Symbol:      "USDC",
	Decimals:    6,
}

var zeroDecimalAsset core.Asset = core.Asset{
	NetworkKind: core.EvmNetworkKind,
	NetworkName: evm.Ethereum.String(),
	Kind:        core.Erc20Token,
	Identifier:  common.Address{}.String(), // lies
	Symbol:      "ZDA",
	Decimals:    0,
}

// NewAmountFromDecimalString tests

func TestFromDecimalString_Exact(t *testing.T) {
	input := "64.234173882933845091"
	amount, err := core.NewAmountFromDecimalString(eth, input)
	assert.Nil(t, err)
	assert.Equal(t, input, amount.Value.String())
}

func TestFromDecimalString_Exact_ZeroDecimals(t *testing.T) {
	input := "64.234173882933845091"
	amount, err := core.NewAmountFromDecimalString(zeroDecimalAsset, input)
	assert.Nil(t, err)
	assert.Equal(t, "64", amount.Value.String())
}

func TestFromDecimalString_TooPrecise(t *testing.T) {
	input := "64.234173882933845091"
	amount, err := core.NewAmountFromDecimalString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "64.234173", amount.Value.String())
}

func TestFromDecimalString_Zero(t *testing.T) {
	input := "0.0"
	amount, err := core.NewAmountFromDecimalString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "0", amount.Value.String())
}

func TestFromDecimalString_WholeNumber(t *testing.T) {
	input := "1"
	amount, err := core.NewAmountFromDecimalString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "1.000000", amount.Value.StringFixed(6))
}

func TestFromDecimalString_Commas(t *testing.T) {
	input := "123,456"
	_, err := core.NewAmountFromDecimalString(usdc, input)
	expected := fmt.Sprintf("Not a valid amount ('123,456') of %s: can't convert 123,456 to decimal", usdc)
	assert.EqualError(t, err, expected)
}

func TestFromDecimalString_MultiDecimals(t *testing.T) {
	input := "123.456.789"
	_, err := core.NewAmountFromDecimalString(usdc, input)
	expected := fmt.Sprintf("Not a valid amount ('123.456.789') of %s: can't convert 123.456.789 to decimal: too many .s", usdc)
	assert.EqualError(t, err, expected)
}

func TestFromDecimalString_Invalid(t *testing.T) {
	input := "abc"
	_, err := core.NewAmountFromDecimalString(eth, input)
	expected := fmt.Sprintf("Not a valid amount ('abc') of %s: can't convert abc to decimal", eth)
	assert.EqualError(t, err, expected)
}

func TestFromDecimalString_Negative(t *testing.T) {
	input := "-123"
	_, err := core.NewAmountFromDecimalString(usdc, input)
	expected := fmt.Sprintf("Negative amount ('-123') of %s not allowed", usdc)
	assert.EqualError(t, err, expected)
}

// NewAmountFromAtomicString tests

func TestFromAtomicString_Exact(t *testing.T) {
	input := "64234173882933845091"
	amount, err := core.NewAmountFromAtomicString(eth, input)
	assert.Nil(t, err)
	assert.Equal(t, "64.234173882933845091", amount.Value.String())
}

func TestFromAtomicString_Zero(t *testing.T) {
	input := "0"
	amount, err := core.NewAmountFromAtomicString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "0", amount.Value.String())
}

func TestFromAtomicString_WholeNumber(t *testing.T) {
	input := "1"
	amount, err := core.NewAmountFromAtomicString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "0.000001", amount.Value.String())
}

func TestFromAtomicString_Invalid(t *testing.T) {
	input := "abc"
	_, err := core.NewAmountFromAtomicString(eth, input)
	expected := fmt.Sprintf("Not a valid amount ('abc') of %s: can't convert abc to decimal", eth)
	assert.EqualError(t, err, expected)
}

func TestFromAtomicString_WithDecimals(t *testing.T) {
	input := "123.456"
	_, err := core.NewAmountFromAtomicString(usdc, input)
	expected := fmt.Sprintf("Fractional atomic units ('123.456') of %s not allowed", usdc)
	assert.EqualError(t, err, expected)
}

func TestFromAtomicString_Negative(t *testing.T) {
	input := "-123"
	_, err := core.NewAmountFromAtomicString(usdc, input)
	expected := fmt.Sprintf("Negative amount ('-123') of %s not allowed", usdc)
	assert.EqualError(t, err, expected)
}

// NewAmountFromDecimal tests

func TestFromDecimal(t *testing.T) {
	input := decimal.NewFromInt(123)
	amount, err := core.NewAmountFromDecimal(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "123.000000", amount.Value.StringFixed(6))
}

func TestFromDecimal_TooPrecise(t *testing.T) {
	input, _ := decimal.NewFromString("123.4567890123456789")
	amount, err := core.NewAmountFromDecimal(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "123.456789", amount.Value.String())
}

// NewAmountFromAtomicValue tests

func TestFromAtomic_Exact(t *testing.T) {
	input := uint64(206824325)
	amount := core.NewAmountFromAtomicValue(eth, input)
	assert.Equal(t, "0.000000000206824325", amount.Value.String())
}

func TestFromAtomic_Zero(t *testing.T) {
	input := uint64(0)
	amount := core.NewAmountFromAtomicValue(usdc, input)
	assert.Equal(t, "0", amount.Value.String())
}

// NewAmountFromAtomicDecimal tests

func TestFromAtomicDecimal_Exact(t *testing.T) {
	input, _ := decimal.NewFromString("1234567")
	amount := core.NewAmountFromAtomicDecimal(usdc, input)
	assert.Equal(t, "1.234567", amount.Value.String())
}

func TestFromAtomicDecimal_TooPrecise(t *testing.T) {
	input, _ := decimal.NewFromString("123.456")
	amount := core.NewAmountFromAtomicDecimal(usdc, input)
	assert.Equal(t, "0.000123", amount.Value.String())
}

// Add tests

func TestAdd(t *testing.T) {
	a, _ := core.NewAmountFromDecimalString(usdc, "123.456789")
	b, _ := core.NewAmountFromDecimalString(usdc, "654.320988")
	result, err := a.Add(b)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("777.777777 %s", usdc), result.String())
}

func TestAdd_Mismatch(t *testing.T) {
	a, _ := core.NewAmountFromDecimalString(usdc, "123.456789")
	b, _ := core.NewAmountFromDecimalString(eth, "1")
	_, err := a.Add(b)
	expected := fmt.Sprintf("Cannot add %s amount and %s amount", usdc, eth)
	assert.EqualError(t, err, expected)
}

// Sub tests

func TestSub(t *testing.T) {
	a, _ := core.NewAmountFromDecimalString(usdc, "123.456789")
	b, _ := core.NewAmountFromDecimalString(usdc, "654.320988")
	result, err := b.Sub(a)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("530.864199 %s", usdc), result.String())
}

func TestSub_Mismatch(t *testing.T) {
	a, _ := core.NewAmountFromDecimalString(usdc, "123.456789")
	b, _ := core.NewAmountFromDecimalString(eth, "1")
	_, err := b.Sub(a)
	expected := fmt.Sprintf("Cannot subtract %s amount and %s amount", eth, usdc)
	assert.EqualError(t, err, expected)
}
