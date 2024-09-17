package ctc

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func AnalyzeTransactions(db *util.FileDB) {
	if os.Getenv("SKIP_ANALYZE") != "" {
		fmt.Println("Skipping transaction analyze step")
		return
	}

	txsDB, txsFound := db.Collections["txs"]
	receiptsDB, receiptsFound := db.Collections["receipts"]
	if !txsFound || txsDB == nil || !receiptsFound || receiptsDB == nil {
		fmt.Println("Cannot analyze transactions without first fetching them")
		return
	}

	// network => tx hash => tx/receipt
	txs := make(map[string]map[string]*types.Transaction)
	receipts := make(map[string]map[string]*types.Receipt)

	txKeys, err := txsDB.List()
	if err != nil {
		fmt.Printf("Error reading tx entries: %s\n", err.Error())
		return
	}

	for _, key := range txKeys {
		if !strings.Contains(key, "-") {
			continue
		}

		var tx types.Transaction
		found, err := txsDB.Read(key, &tx)
		if err != nil {
			fmt.Printf("Error reading tx from cache for %s: %s\n", key, err.Error())
			continue
		}
		if !found {
			fmt.Printf("Tx cache disappeared for %s\n", key)
			continue
		}

		var receipt types.Receipt
		found, err = receiptsDB.Read(key, &receipt)
		if err != nil {
			fmt.Printf("Error reading receipt from cache for %s: %s\n", key, err.Error())
			continue
		}
		if !found {
			fmt.Printf("Receipt cache disappeared for %s\n", key)
			continue
		}

		keyParts := strings.Split(key, "-")
		network := keyParts[0]
		hash := keyParts[1]

		if txs[network] == nil {
			txs[network] = make(map[string]*types.Transaction)
		}
		if receipts[network] == nil {
			receipts[network] = make(map[string]*types.Receipt)
		}

		txs[network][hash] = &tx
		receipts[network][hash] = &receipt
	}

	for network, networkTxs := range txs {
		fmt.Printf("%s: %d txs\n", network, len(networkTxs))
	}
}
