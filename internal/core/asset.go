package core

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type AssetKind string

const (
	UtxoNative AssetKind = "utxo_native"
	EvmNative  AssetKind = "evm_native"
	Erc20Token AssetKind = "erc20"
	Erc721Nft  AssetKind = "erc721"
	SvmNative  AssetKind = "svm_native"
	SplToken   AssetKind = "spl_token"
	SplNft     AssetKind = "spl_nft"
)

// An Asset uniquely identifies an asset across all networks. It is used in
// combination with a decimal value to track amounts for calculations, and
// ensure proper usage of these amounts with one another.
type Asset struct {
	NetworkKind NetworkKind // Category of network (EVM, UTXO, Cosmos, Solana, etc)
	NetworkName string      // Common name of the network
	Kind        AssetKind   // Kind of asset (token, nft, native asset, etc)
	Identifier  string      // Exact identifier for the asset (token contract, etc)
	Symbol      string      // Human-readable symbol for the asset
	Decimals    uint8       // Number of decimal places used for formatting
}

func (a Asset) String() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		a.NetworkKind,
		a.NetworkName,
		a.Kind,
		a.Identifier,
		a.Symbol,
	)
}

func (a Asset) FungibleWith(other Asset) bool {
	if a == other {
		return true
	}

	bothEvm := a.NetworkKind == EvmNetworkKind && other.NetworkKind == EvmNetworkKind
	bothNative := a.Kind == EvmNative && other.Kind == EvmNative

	if bothEvm && bothNative {
		return a.Symbol == other.Symbol
	}

	return false
}

func (a Asset) WithDecimalStringValue(s string) (Amount, error) {
	return NewAmountFromDecimalString(a, s)
}

func (a Asset) WithAtomicStringValue(s string) (Amount, error) {
	return NewAmountFromAtomicString(a, s)
}

func (a Asset) WithDecimalValue(d decimal.Decimal) (Amount, error) {
	return NewAmountFromDecimal(a, d)
}

func (a Asset) WithAtomicDecimalValue(d decimal.Decimal) Amount {
	return NewAmountFromAtomicDecimal(a, d)
}

func (a Asset) WithAtomicValue(v uint64) Amount {
	return NewAmountFromAtomicValue(a, v)
}
