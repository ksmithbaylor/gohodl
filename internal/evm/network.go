package evm

import (
	"fmt"

	"github.com/ksmithbaylor/gohodl/internal/core"
)

type EvmNetwork struct {
	Name        EvmNetworkName `mapstructure:"name"`
	ChainID     uint           `mapstructure:"chain_id"`
	NativeAsset string         `mapstructure:"native_asset"`
	RPCs        []string       `mapstructure:"rpcs"`
}

func (n EvmNetwork) NativeEvmAsset() core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        core.EvmNative,
		Identifier:  EvmNullAddress.String(),
		Symbol:      n.NativeAsset,
		Decimals:    18,
	}
}

func (n EvmNetwork) Erc20TokenAsset(contractAddress, symbol string, decimals uint8) core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        core.Erc20Token,
		Identifier:  EvmAddress(contractAddress).String(),
		Symbol:      symbol,
		Decimals:    decimals,
	}
}

func (n EvmNetwork) Erc721NftAsset(contractAddress, symbol string, tokenID uint64) core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        core.Erc721Nft,
		Identifier:  fmt.Sprintf("%s/%d", EvmAddress(contractAddress).String(), tokenID),
		Symbol:      symbol,
		Decimals:    0,
	}
}
