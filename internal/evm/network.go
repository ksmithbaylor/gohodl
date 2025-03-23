package evm

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

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
	ExplorerURLs      struct {
		Tx   string `mapstructure:"tx"`
		Addr string `mapstructure:"addr"`
	} `mapstructure:"explorer_urls"`
	Etherscan struct {
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

func (n Network) OpenTransactionInExplorer(hash string, wait ...bool) {
	url := strings.Replace(n.ExplorerURLs.Tx, "TX", hash, 1)
	time.Sleep(time.Millisecond * 100)
	cmd := exec.Command("/usr/bin/open", "-u", url, "-a", "Google Chrome")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(cmd.String())
		fmt.Println(string(output))
		fmt.Println(err.Error())
	}
	if len(wait) > 0 && wait[0] {
		_, _ = fmt.Scanln()
	}
}
