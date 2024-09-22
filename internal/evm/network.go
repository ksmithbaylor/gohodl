package evm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ksmithbaylor/gohodl/internal/core"
)

const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000"

type Network struct {
	Name              NetworkName `mapstructure:"name"`
	ChainID           uint        `mapstructure:"chain_id"`
	NativeAssetSymbol string      `mapstructure:"native_asset"`
	RPCs              []string    `mapstructure:"rpcs"`
	SettlesTo         NetworkName `mapstructure:"settles_to"`
	Etherscan         struct {
		URL string `mapstructure:"url"`
		Key string `mapstructure:"key"`
		RPS uint   `mapstructure:"rps"`
	} `mapstructure:"etherscan"`
}

func (n Network) GetKind() core.NetworkKind {
	return core.EvmNetworkKind
}

func (n Network) GetName() string {
	return n.Name.String()
}

func (n Network) NativeAsset() core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkName: n.Name.String(),
		Kind:        core.EvmNative,
		Identifier:  common.Address{}.String(),
		Symbol:      n.NativeAssetSymbol,
		Decimals:    18,
	}
}

func (n Network) Erc20TokenAsset(contractAddress, symbol string, decimals uint8) core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkName: n.Name.String(),
		Kind:        core.Erc20Token,
		Identifier:  contractAddress,
		Symbol:      symbol,
		Decimals:    decimals,
	}
}

func (n Network) Erc721NftAsset(contractAddress, symbol string, tokenID uint64) core.Asset {
	return core.Asset{
		NetworkKind: core.EvmNetworkKind,
		NetworkName: n.Name.String(),
		Kind:        core.Erc721Nft,
		Identifier:  fmt.Sprintf("%s/%d", contractAddress, tokenID),
		Symbol:      symbol,
		Decimals:    0,
	}
}
