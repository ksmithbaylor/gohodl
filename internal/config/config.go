package config

import (
	"fmt"
	"log"

	"github.com/ksmithbaylor/gohodl/internal/types"
	"github.com/spf13/viper"
)

// Global, read in at startup
var Config config

////////////////////////////////////////////////////////////////////////////////
// Top-level

type config struct {
	Ownership   blockchains                                     `mapstructure:"ownership"`
	EvmNetworks map[types.EthNetworkName]types.EthNetworkConfig `mapstructure:"evm_networks"`
}

type blockchains struct {
	Bitcoin  utxo     `mapstructure:"bitcoin"`
	Ethereum ethereum `mapstructure:"ethereum"`
	Solana   solana   `mapstructure:"solana"`
	Cosmos   cosmos   `mapstructure:"cosmos"`
	Dogecoin utxo     `mapstructure:"dogecoin"`
	Litecoin utxo     `mapstructure:"litecoin"`
}

type addresses[t any] map[string]t

////////////////////////////////////////////////////////////////////////////////
// UTXO networks

type utxo struct {
	Xpubs map[string]xpub `mapstructure:"xpubs"`
}

type xpub struct {
	Type types.UtxoWalletScheme `mapstructure:"type"`
	Key  string                 `mapstructure:"key"`
}

////////////////////////////////////////////////////////////////////////////////
// Ethereum

type ethereum struct {
	Addresses addresses[types.EthAddress]                          `mapstructure:"addresses"`
	Instadapp map[types.EthNetworkName]map[string]types.EthAddress `mapstructure:"instadapp"`
}

////////////////////////////////////////////////////////////////////////////////
// Solana

type solana struct {
	Addresses addresses[types.SolanaAddress] `mapstructure:"addresses"`
}

////////////////////////////////////////////////////////////////////////////////
// Cosmos

type cosmos map[types.CosmosNetwork]addresses[types.CosmosAddress]

////////////////////////////////////////////////////////////////////////////////
// Initialization

func init() {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.AddConfigPath(".")
	v.SetConfigFile("config.yml")

	if err := v.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("could not read in config: %w", err))
	}

	if err := v.Unmarshal(&Config); err != nil {
		log.Fatal(fmt.Errorf("invalid config: %w", err))
	}
}
