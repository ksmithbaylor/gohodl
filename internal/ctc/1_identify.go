package ctc

import (
	"fmt"
	"os"
	"sync"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type cachedTxs struct {
	Network string   `json:"network"`
	Block   int      `json:"block"`
	Address string   `json:"address"`
	Txs     []string `json:"txs"`
}

func IdentifyTransactions(db *util.FileDB, clients generic.AllNodeClients) {
	cfg := config.Config

	txHashesDB := db.NewCollection("evm_tx_hashes")

	if os.Getenv("SKIP_IDENTIFY") != "" {
		fmt.Println("Skipping transaction identification step")
		return
	}

	indexers := generic.NewAllIndexers(cfg.AllNetworks())
	latestBlocks := clients.LatestBlocks()

	fmt.Println("Getting transaction hashes for each address...")

	var mu sync.Mutex
	txHashes := make(map[string]map[string][]string, 0) // address -> network -> list of tx hashes
	errors := make(map[string]map[string]error, 0)      // address -> network -> error
	for name, addr := range cfg.Ownership.Ethereum.Addresses {
		label := fmt.Sprintf("%s (%s)", addr.Hex(), name)
		txHashes[label] = make(map[string][]string, 0)
		errors[label] = make(map[string]error, 0)
	}

	var wg sync.WaitGroup

	for _, network := range cfg.AllNetworks() {
		wg.Add(1)
		go func(network core.Network) {
			defer wg.Done()

			indexer, found := indexers[network.GetName()]
			if !found {
				fmt.Printf("No indexer found for %s, skipping\n", network.GetName())
				return
			}

			latestBlock, found := latestBlocks[network.GetName()]
			if !found {
				fmt.Printf("No latest block found for %s, skipping\n", network.GetName())
				return
			}

			for name, addr := range cfg.Ownership.Ethereum.Addresses {
				label := fmt.Sprintf("%s (%s)", addr.Hex(), name)
				cacheKey := fmt.Sprintf("%s-%s", network.GetName(), addr.Hex())

				var firstBlock *int
				knownTxs := make([]string, 0)

				var cached cachedTxs
				cacheFound, err := txHashesDB.Read(cacheKey, &cached)
				if err != nil {
					fmt.Printf("Error reading cache for %s: %s\n", cacheKey, err.Error())
				} else if cacheFound {
					if cached.Address == addr.Hex() && cached.Network == network.GetName() {
						firstBlock = &cached.Block
						knownTxs = cached.Txs
					} else {
						fmt.Printf("Mismatched cache contents for %s, skipping\n", cacheKey)
					}
				}

				latestBlockInt := int(latestBlock)
				txs, err := indexer.GetAllTransactionHashes(addr.Hex(), firstBlock, &latestBlockInt)
				if err != nil {
					fmt.Printf("%s - %s: Error getting transactions: %s\n", label, network.GetName(), err.Error())
					mu.Lock()
					errors[label][network.GetName()] = err
					mu.Unlock()
					continue
				}

				allTxs := util.UniqueItems(knownTxs, txs)

				fmt.Printf(
					"%s - %s: %d txs already known, %d new txs found, %d total\n",
					label,
					network.GetName(),
					len(knownTxs),
					len(txs),
					len(allTxs),
				)
				err = txHashesDB.Write(cacheKey, cachedTxs{
					Network: network.GetName(),
					Address: addr.Hex(),
					Block:   latestBlockInt,
					Txs:     allTxs,
				})
				if err != nil {
					fmt.Printf("Failed to write cache for %s: %s\n", cacheKey, err.Error())
				}
				mu.Lock()
				txHashes[label][network.GetName()] = allTxs
				mu.Unlock()
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
