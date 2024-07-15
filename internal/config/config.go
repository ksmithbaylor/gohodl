package config

import (
	"fmt"
	"log"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Global, read in at startup
var Config config

////////////////////////////////////////////////////////////////////////////////
// Top-level

type config struct {
	Ownership   blockchains   `mapstructure:"ownership"`
	EvmNetworks []evm.Network `mapstructure:"evm_networks"`
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
	Type core.UtxoWalletScheme `mapstructure:"type"`
	Key  string                `mapstructure:"key"`
}

////////////////////////////////////////////////////////////////////////////////
// Ethereum

type ethereum struct {
	Addresses addresses[common.Address] `mapstructure:"addresses"`
}

////////////////////////////////////////////////////////////////////////////////
// Solana

type solana struct {
	Addresses addresses[core.SolanaAddress] `mapstructure:"addresses"`
}

////////////////////////////////////////////////////////////////////////////////
// Cosmos

type cosmos map[core.CosmosNetwork]addresses[core.CosmosAddress]

////////////////////////////////////////////////////////////////////////////////
// Initialization

func init() {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.AddConfigPath(".")
	v.SetConfigFile("config.yml")

	if err := v.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("could not read in config: %w", err))
	}

	if err := v.Unmarshal(&Config, viper.DecodeHook(CustomDecoder())); err != nil {
		log.Fatal(fmt.Errorf("invalid config: %w", err))
	}
}

func CustomDecoder() mapstructure.DecodeHookFunc {
	return func(dataType reflect.Type, targetType reflect.Type, data any) (any, error) {
		if dataType.Kind() != reflect.String {
			return data, nil
		}

		if targetType != reflect.TypeOf(common.Address{}) {
			return data, nil
		}

		return common.HexToAddress(data.(string)), nil
	}
}
