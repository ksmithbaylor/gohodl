package ctc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func readTransactionBundle(db *util.FileDB, network, hash string) (
	*types.Transaction,
	*types.Receipt,
	*types.Header,
	error,
) {
	txsDB, txsFound := db.Collections["txs"]
	receiptsDB, receiptsFound := db.Collections["receipts"]
	blocksDB, blocksFound := db.Collections["blocks"]
	if !txsFound || txsDB == nil ||
		!receiptsFound || receiptsDB == nil ||
		!blocksFound || blocksDB == nil {
		return nil, nil, nil, fmt.Errorf("Cannot read transactions without first fetching them")
	}

	key := fmt.Sprintf("%s-%s", network, hash)

	var tx types.Transaction
	found, err := txsDB.Read(key, &tx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading tx for %s: %w", key, err)
	}
	if !found {
		return nil, nil, nil, fmt.Errorf("Transaction not found: %s", key)
	}

	var receipt types.Receipt
	found, err = receiptsDB.Read(key, &receipt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading tx receipt for %s: %w", key, err)
	}
	if !found {
		return nil, nil, nil, fmt.Errorf("Transaction receipt not found: %s", key)
	}

	var block types.Header
	blockHash := receipt.BlockHash.Hex()
	blockKey := fmt.Sprintf("%s-%s", network, blockHash)
	found, err = blocksDB.Read(blockKey, &block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading block from cache for %s: %w", blockKey, err)
	}
	if !found {
		return nil, nil, nil, fmt.Errorf("Block cache disappeared for %s", blockKey)
	}

	return &tx, &receipt, &block, nil
}
