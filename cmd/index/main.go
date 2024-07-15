package main

import (
	"fmt"
	"log"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
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

	for name, addr := range cfg.Ownership.Ethereum.Addresses {
		fmt.Printf("%s (%s):\n", addr.Hex(), name)

		for _, network := range cfg.EvmNetworks {
			fmt.Printf("  %s:\n", network.Name)

			indexer, found := indexers[string(network.Name)]
			if !found {
				fmt.Println("    No indexer found, skipping")
				continue
			}

			txs, err := indexer.GetAllTransactionHashes(addr.Hex())
			if err != nil {
				fmt.Printf("    Error getting transactions: %s", err.Error())
			}

			fmt.Printf("    %d transactions found\n", len(txs))
		}
	}
}
