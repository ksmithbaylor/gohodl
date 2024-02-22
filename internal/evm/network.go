package evm

import (
	"fmt"

	"github.com/ksmithbaylor/gohodl/internal/core"
)

type Network struct {
	Name        NetworkName `mapstructure:"name"`
	ChainID     uint        `mapstructure:"chain_id"`
	NativeAsset string      `mapstructure:"native_asset"`
	RPCs        []string    `mapstructure:"rpcs"`
}

func (n Network) NativeEvmAsset() core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        core.EvmNative,
		Identifier:  NullAddress.String(),
		Symbol:      n.NativeAsset,
		Decimals:    18,
	}
}

func (n Network) Erc20TokenAsset(contractAddress, symbol string, decimals uint8) core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        core.Erc20Token,
		Identifier:  Address(contractAddress).String(),
		Symbol:      symbol,
		Decimals:    decimals,
	}
}

func (n Network) Erc721NftAsset(contractAddress, symbol string, tokenID uint64) core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        core.Erc721Nft,
		Identifier:  fmt.Sprintf("%s/%d", Address(contractAddress).String(), tokenID),
		Symbol:      symbol,
		Decimals:    0,
	}
}
