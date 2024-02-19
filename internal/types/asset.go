package types

import (
	"fmt"
)

type AssetKind string

const (
	UTXOAsset  AssetKind = "utxo_asset"
	EVMNative  AssetKind = "evm_native"
	ERC20Token AssetKind = "erc20"
	ERC721NFT  AssetKind = "erc721"
	SVMNative  AssetKind = "svm_native"
	SPLToken   AssetKind = "spl_token"
	SPLNFT     AssetKind = "spl_nft"
)

type Asset struct {
	NetworkKind NetworkKind  // Category of network (EVM, UTXO, Cosmos, Solana, etc)
	NetworkID   fmt.Stringer // Network identifier (chain id or name)
	Kind        AssetKind    // Kind of asset (token, nft, native asset, etc)
	Identifier  string       // Exact identifier for the asset (token contract, etc)
	Symbol      string       // Human-readable symbol for the asset
	Decimals    uint8        // Number of decimal places used for formatting
}

func (a Asset) String() string {
	return fmt.Sprintf("%s/%s/%s/%s/%s",
		a.NetworkKind,
		a.NetworkID,
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
	bothNative := a.Kind == EVMNative && other.Kind == EVMNative

	if bothEvm && bothNative {
		return a.Symbol == other.Symbol
	}

	return false
}
