package types

import (
	"fmt"
)

type Asset struct {
	NetworkKind NetworkKind // Category of network (EVM, UTXO, Cosmos, Solana, etc)
	NetworkID   NetworkID   // Network identifier (chain id or name)
	AssetSymbol string      // Human-readable symbol for the asset
	AssetID     string      // Exact identifier for the asset (token contract, etc)
}

func (a Asset) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", a.NetworkKind, a.NetworkID, a.AssetSymbol, a.AssetID)
}
