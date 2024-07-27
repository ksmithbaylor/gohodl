package main

import (
	"fmt"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/generic"
)

func main() {
	allNetworks := config.Config.EvmNetworks

	longestName := 0
	for _, network := range allNetworks {
		longestName = max(longestName, len(network.Name))
	}

	fmt.Println("Connecting...")

	clients := generic.NewAllNodeClients(config.Config.AllNetworks())

	fmt.Printf("\nBlock numbers:\n")
	clients.ForEach(func(networkName string, client core.NodeClient) {
		block, err := client.LatestBlock()
		if err != nil {
			fmt.Printf("%-*s  %s\n", longestName, networkName, err.Error())
		}

		fmt.Printf("%-*s  %d\n", longestName, networkName, block)
	})

	for label, addr := range config.Config.Ownership.Ethereum.Addresses {
		a := addr
		fmt.Printf("\nBalance of %s:\n", label)
		clients.ForEach(func(networkName string, client core.NodeClient) {
			evmClient := client.(*evm.Client)
			balance, err := evmClient.Balance(a)
			if err != nil {
				fmt.Printf("%-*s  %s\n", longestName, networkName, err.Error())
			}
			fmt.Printf("%-*s  %s (%s)\n", longestName, networkName, balance, balance.Asset.Symbol)
		})
	}
}
