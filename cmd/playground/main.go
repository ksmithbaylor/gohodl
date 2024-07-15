package main

import (
	"fmt"
	"sync"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

func main() {
	allNetworks := config.Config.EvmNetworks

	clients := make(map[evm.NetworkName]*evm.Client)
	var clientsMu sync.Mutex
	var wg sync.WaitGroup

	longestName := 0
	for _, network := range allNetworks {
		longestName = max(longestName, len(network.Name))
	}

	fmt.Println("Connecting...")

	for _, network := range allNetworks {
		network := network
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := evm.NewClient(network)
			if err != nil {
				fmt.Printf("%-*s  %s\n", longestName, network.Name, err.Error())
			}
			clientsMu.Lock()
			clients[network.Name] = client
			clientsMu.Unlock()
		}()
	}
	wg.Wait()

	forAll("Block numbers", clients, func(networkName evm.NetworkName, client *evm.Client) {
		block, err := client.LatestBlock()
		if err != nil {
			fmt.Printf("%-*s  %s\n", longestName, networkName, err.Error())
		}

		fmt.Printf("%-*s  %d\n", longestName, networkName, block)
	})

	for label, addr := range config.Config.Ownership.Ethereum.Addresses {
		a := addr
		forAll(fmt.Sprintf("Balance of %s", label), clients, func(networkName evm.NetworkName, client *evm.Client) {
			balance, err := client.Balance(a)
			if err != nil {
				fmt.Printf("%-*s  %s\n", longestName, networkName, err.Error())
			}

			fmt.Printf("%-*s  %s (%s)\n", longestName, networkName, balance, balance.Asset.Symbol)
		})
	}
}

func forAll(label string, clients map[evm.NetworkName]*evm.Client, action func(evm.NetworkName, *evm.Client)) {
	var wg sync.WaitGroup

	fmt.Printf("\n%s:\n", label)
	for n, c := range clients {
		networkName := n
		client := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			action(networkName, client)
		}()
	}
	wg.Wait()
}
