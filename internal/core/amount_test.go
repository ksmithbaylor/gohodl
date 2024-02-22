package core_test

import (
	"fmt"
	"testing"

	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/stretchr/testify/assert"
)

var eth core.Asset = core.Asset{
	NetworkKind: core.EvmNetworkKind,
	NetworkID:   evm.Ethereum,
	Kind:        core.EvmNative,
	Identifier:  evm.EvmNullAddress.String(),
	Symbol:      "ETH",
	Decimals:    18,
}

var usdc core.Asset = core.Asset{
	NetworkKind: core.EvmNetworkKind,
	NetworkID:   evm.Ethereum,
	Kind:        core.Erc20Token,
	Identifier:  evm.EvmAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48").String(),
	Symbol:      "USDC",
	Decimals:    6,
}

func TestFromDecimalString_Exact(t *testing.T) {
	input := "64.234173882933845091"
	amount, err := core.NewAmountFromDecimalString(eth, input)
	assert.Nil(t, err)
	assert.Equal(t, input, amount.Value.String())
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

func TestFromDecimalString_Invalid(t *testing.T) {
	input := "abc"
	_, err := core.NewAmountFromDecimalString(eth, input)
	expected := fmt.Sprintf("Not a valid amount ('abc') of %s: can't convert abc to decimal", eth)
	assert.EqualError(t, err, expected)
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

func TestFromDecimalString_Negative(t *testing.T) {
	input := "-123"
	_, err := core.NewAmountFromDecimalString(usdc, input)
	expected := fmt.Sprintf("Negative amount ('-123') of %s not allowed", usdc)
	assert.EqualError(t, err, expected)
}

func TestFromCentsString_Exact(t *testing.T) {
	input := "64234173882933845091"
	amount, err := core.NewAmountFromCentsString(eth, input)
	assert.Nil(t, err)
	assert.Equal(t, "64.234173882933845091", amount.Value.String())
}

func TestFromCentsString_Zero(t *testing.T) {
	input := "0"
	amount, err := core.NewAmountFromCentsString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "0", amount.Value.String())
}

func TestFromCentsString_WholeNumber(t *testing.T) {
	input := "1"
	amount, err := core.NewAmountFromCentsString(usdc, input)
	assert.Nil(t, err)
	assert.Equal(t, "0.000001", amount.Value.String())
}

func TestFromCentsString_Invalid(t *testing.T) {
	input := "abc"
	_, err := core.NewAmountFromCentsString(eth, input)
	expected := fmt.Sprintf("Not a valid amount ('abc') of %s: can't convert abc to decimal", eth)
	assert.EqualError(t, err, expected)
}

func TestFromCentsString_Negative(t *testing.T) {
	input := "-123"
	_, err := core.NewAmountFromCentsString(usdc, input)
	expected := fmt.Sprintf("Negative amount ('-123') of %s not allowed", usdc)
	assert.EqualError(t, err, expected)
}

func TestFromCents_Exact(t *testing.T) {
	input := uint64(206824325)
	amount := core.NewAmountFromCents(eth, input)
	assert.Equal(t, "0.000000000206824325", amount.Value.String())
}

func TestFromCents_Zero(t *testing.T) {
	input := uint64(0)
	amount := core.NewAmountFromCents(usdc, input)
	assert.Equal(t, "0", amount.Value.String())
}
