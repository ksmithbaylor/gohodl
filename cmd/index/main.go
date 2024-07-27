package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/indexing"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type cachedTxs struct {
	Network string   `json:"network"`
	Block   int      `json:"block"`
	Address string   `json:"address"`
	Txs     []string `json:"txs"`
}

func main() {
	cfg := config.Config

	cache, err := util.NewFileCache("evm_tx_hashes")
	if err != nil {
		log.Fatal(err)
	}

	clients, errs := evm.AllClients(cfg.EvmNetworks)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println("Getting transaction hashes for each address...")

	indexers := make(map[evm.NetworkName]core.Indexer, 0)
	latestBlocks := make(map[evm.NetworkName]int)

	for _, network := range cfg.EvmNetworks {
		indexer, err := indexing.GetIndexerForNetwork(network)
		if err != nil {
			log.Fatal(err)
		}
		indexers[network.Name] = indexer

		block, err := clients[network.Name].LatestBlock()
		if err != nil {
			log.Fatal(err)
		}
		latestBlocks[network.Name] = int(block)
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

			indexer, found := indexers[network.Name]
			if !found {
				fmt.Printf("No indexer found for %s, skipping\n", network.Name)
				return
			}

			latestBlock, found := latestBlocks[network.Name]
			if !found {
				fmt.Printf("No latest block found for %s, skipping\n", network.Name)
				return
			}

			for name, addr := range cfg.Ownership.Ethereum.Addresses {
				label := fmt.Sprintf("%s (%s)", addr.Hex(), name)
				cacheKey := fmt.Sprintf("%s-%s", network.Name, addr.Hex())

				var firstBlock *int
				knownTxs := make([]string, 0)

				var cached cachedTxs
				cacheFound, err := cache.Read(cacheKey, &cached)
				if err != nil {
					fmt.Printf("Error reading cache for %s: %s\n", cacheKey, err.Error())
				} else if cacheFound {
					if cached.Address == addr.Hex() && cached.Network == string(network.Name) {
						firstBlock = &cached.Block
						knownTxs = cached.Txs
					} else {
						fmt.Printf("Mismatched cache contents for %s, skipping\n", cacheKey)
					}
				}

				txs, err := indexer.GetAllTransactionHashes(addr.Hex(), firstBlock, &latestBlock)
				if err != nil {
					fmt.Printf("%s - %s: Error getting transactions: %s\n", label, network.Name, err.Error())
					errors[label][string(network.Name)] = err
					continue
				}

				allTxs := util.UniqueItems(knownTxs, txs)

				fmt.Printf(
					"%s - %s: %d txs already known, %d new txs found, %d total\n",
					label,
					network.Name,
					len(knownTxs),
					len(txs),
					len(allTxs),
				)
				err = cache.Write(cacheKey, cachedTxs{
					Network: string(network.Name),
					Address: addr.Hex(),
					Block:   latestBlock,
					Txs:     allTxs,
				})
				if err != nil {
					fmt.Printf("Failed to write cache for %s: %s\n", cacheKey, err.Error())
				}
				txHashes[label][string(network.Name)] = allTxs
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
