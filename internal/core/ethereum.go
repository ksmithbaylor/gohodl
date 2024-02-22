package core

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

////////////////////////////////////////////////////////////////////////////////

type EvmAddress string

const EvmNullAddress EvmAddress = "0x0000000000000000000000000000000000000000"

func (a EvmAddress) ToGeth() common.Address {
	return common.HexToAddress(string(a))
}

func (a EvmAddress) String() string {
	// Round-trip to get casing/checksum
	return a.ToGeth().String()
}

////////////////////////////////////////////////////////////////////////////////

type EvmNetworkName string

const (
	Ethereum  EvmNetworkName = "ethereum"
	Polygon   EvmNetworkName = "polygon"
	Avalanche EvmNetworkName = "avalanche"
	Base      EvmNetworkName = "base"
	Optimism  EvmNetworkName = "optimism"
	Moonbeam  EvmNetworkName = "moonbeam"
	Moonriver EvmNetworkName = "moonriver"
	Fantom    EvmNetworkName = "fantom"
	Evmos     EvmNetworkName = "evmos"
)

func (n EvmNetworkName) String() string {
	return string(n)
}

////////////////////////////////////////////////////////////////////////////////

type EvmNetwork struct {
	Name        EvmNetworkName `mapstructure:"name"`
	ChainID     uint           `mapstructure:"chain_id"`
	NativeAsset string         `mapstructure:"native_asset"`
	RPCs        []string       `mapstructure:"rpcs"`
}

func (n EvmNetwork) NativeEvmAsset() Asset {
	return Asset{
		NetworkKind: EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        EvmNative,
		Identifier:  EvmNullAddress.String(),
		Symbol:      n.NativeAsset,
		Decimals:    18,
	}
}

func (n EvmNetwork) Erc20TokenAsset(contractAddress, symbol string, decimals uint8) Asset {
	return Asset{
		NetworkKind: EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        Erc20Token,
		Identifier:  EvmAddress(contractAddress).String(),
		Symbol:      symbol,
		Decimals:    decimals,
	}
}

func (n EvmNetwork) Erc721NftAsset(contractAddress, symbol string, tokenID uint64) Asset {
	return Asset{
		NetworkKind: EvmNetworkKind,
		NetworkID:   n.Name,
		Kind:        Erc721Nft,
		Identifier:  fmt.Sprintf("%s/%d", EvmAddress(contractAddress).String(), tokenID),
		Symbol:      symbol,
		Decimals:    0,
	}
}
