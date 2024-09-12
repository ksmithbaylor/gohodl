package ctc

import (
	"fmt"

	// "github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func FetchTransactions(db *util.FileDB, clients generic.AllNodeClients) {
	txHashesDB, found := db.Collections["evm_tx_hashes"]
	if !found || txHashesDB == nil {
		fmt.Println("Cannot fetch transactions without first running identification step")
		return
	}

	txHashesKeys, err := txHashesDB.List()
	if err != nil {
		fmt.Printf("Error listing keys in tx hashes collection: %s\n", err.Error())
		return
	}

	for _, key := range txHashesKeys {
		var entry cachedTxs
		cacheFound, err := txHashesDB.Read(key, &entry)
		if err != nil {
			fmt.Printf("Error reading tx hash cache for %s: %s\n", key, err.Error())
			continue
		}
		if !cacheFound {
			fmt.Printf("Tx hash cache disappeared for %s\n", key)
			continue
		}

		if len(entry.Txs) == 0 {
			continue
		}

		client := clients[entry.Network]
		if client == nil {
			fmt.Printf("No client found for %s\n", entry.Network)
			continue
		}

		// TODO actually fetch and cache all transaction headers and receipts
		fmt.Printf("Fetching %d transactions for %s on %s\n", len(entry.Txs), entry.Address, entry.Network)

		// evmClient, ok := client.(*evm.Client)
		// if !ok {
		//   fmt.Printf("Non-EVM networks (like %s) not implemented yet\n", entry.Network)
		//   continue
		// }
	}
}
