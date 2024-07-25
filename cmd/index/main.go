package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/indexing"
)

func main() {
	cfg := config.Config

	fmt.Println("Getting transaction hashes for each address...")

	indexers := make(map[string]core.Indexer, 0)

	for _, network := range cfg.EvmNetworks {
		indexer, err := indexing.GetIndexerForNetwork(network)
		if err != nil {
			log.Fatal(err)
		}
		indexers[string(network.Name)] = indexer
	}

	txHashes := make(map[string]map[string][]string, 0) // address -> network -> list of tx hashes
	errors := make(map[string]map[string]error, 0)      // address -> network -> error
	for name, addr := range cfg.Ownership.Ethereum.Addresses {
		label := fmt.Sprintf("%s (%s)", addr.Hex(), name)
		txHashes[label] = make(map[string][]string, 0)
		errors[label] = make(map[string]error, 0)
	}

	var wg sync.WaitGroup

	for _, network := range cfg.EvmNetworks {
		wg.Add(1)
		go func(network evm.Network) {
			defer wg.Done()

			indexer, found := indexers[string(network.Name)]
			if !found {
				fmt.Printf("No indexer found for %s, skipping\n", network.Name)
				return
			}

			for name, addr := range cfg.Ownership.Ethereum.Addresses {
				label := fmt.Sprintf("%s (%s)", addr.Hex(), name)
				txs, err := indexer.GetAllTransactionHashes(addr.Hex())
				if err != nil {
					fmt.Printf("%s - %s: Error getting transactions: %s\n", label, network.Name, err.Error())
					errors[label][string(network.Name)] = err
					continue
				}

				fmt.Printf("%s - %s: %d transactions found\n", label, network.Name, len(txs))
				txHashes[label][string(network.Name)] = txs
			}
		}(network)
	}

	wg.Wait()

	fmt.Printf("\n------------------------------------------------------------\n\n")

	totalTxs := 0

	for addrLabel, txsByNetwork := range txHashes {
		totalAddressTxs := 0
		for _, txs := range txsByNetwork {
			totalAddressTxs += len(txs)
		}
		totalTxs += totalAddressTxs
		fmt.Printf("%s: %d total txs\n", addrLabel, totalAddressTxs)
		for network, txs := range txsByNetwork {
			fmt.Printf("  %s: %d txs\n", network, len(txs))
		}
		fmt.Println()
	}

	fmt.Printf("%d total txs across all addresses and networks\n", totalTxs)

	fmt.Printf("\n------------------------------------------------------------\n\n")

	for addrLabel, errorsByNetwork := range errors {
		errCount := 0
		for _, err := range errorsByNetwork {
			if err != nil {
				errCount += 1
			}
		}
		if errCount == 0 {
			break
		}
		fmt.Printf("%s: %d total errors\n", addrLabel, errCount)
		for network, err := range errorsByNetwork {
			if err != nil {
				fmt.Printf("  %s: %s\n", network, err.Error())
			}
		}
		fmt.Println()
	}
}
