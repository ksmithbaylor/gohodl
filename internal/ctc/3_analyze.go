package ctc

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func AnalyzeTransactions(db *util.FileDB) {
	if os.Getenv("SKIP_ANALYZE") != "" {
		fmt.Println("Skipping transaction analyze step")
		return
	}

	txsDB, txsFound := db.Collections["txs"]
	if !txsFound || txsDB == nil {
		fmt.Println("Cannot analyze transactions without first fetching them")
		return
	}

	// network => tx/block hash => tx/receipt/block
	txs := make(map[string]map[string]*types.Transaction)
	receipts := make(map[string]map[string]*types.Receipt)
	blocks := make(map[string]map[string]*types.Header)

	txKeys, err := txsDB.List()
	if err != nil {
		fmt.Printf("Error reading tx entries: %s\n", err.Error())
		return
	}

	for _, key := range txKeys {
		if !strings.Contains(key, "-") {
			continue
		}

		keyParts := strings.Split(key, "-")
		network := keyParts[0]
		hash := keyParts[1]

		tx, receipt, block, err := readTransactionBundle(db, network, hash)
		if err != nil {
			fmt.Printf("Could not read %s transaction %s: %s\n", network, hash, err.Error())
			continue
		}

		if txs[network] == nil {
			txs[network] = make(map[string]*types.Transaction)
		}
		if receipts[network] == nil {
			receipts[network] = make(map[string]*types.Receipt)
		}
		if blocks[network] == nil {
			blocks[network] = make(map[string]*types.Header)
		}

		txs[network][hash] = tx
		receipts[network][hash] = receipt
		blocks[network][block.Hash().Hex()] = block
	}

	txCsvFile, err := os.Create(getTxsCsvPath(db))
	if err != nil {
		fmt.Printf("Error creating csv file: %s\n", err.Error())
		return
	}
	defer txCsvFile.Close()

	txCsvWriter := csv.NewWriter(txCsvFile)
	defer txCsvWriter.Flush()

	err = txCsvWriter.Write([]string{
		"network",
		"hash",
		"blockhash",
		"from",
		"to",
		"method",
		"value",
		"success",
	})
	if err != nil {
		fmt.Printf("Error writing csv header row: %s\n", err.Error())
		return
	}

	for network, networkTxs := range txs {
		fmt.Printf("%s: %d txs\n", network, len(networkTxs))

		for txHash, tx := range networkTxs {
			receipt := receipts[network][txHash]
			block := blocks[network][receipt.BlockHash.Hex()]

			from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
			if err != nil {
				from = common.HexToAddress(evm.ZERO_ADDRESS)
			}

			method := ""
			if len(tx.Data()) >= 4 {
				method = "0x" + common.Bytes2Hex(tx.Data()[:4])
			}

			status := "success"
			if receipt.Status == 0 {
				status = "failed"
			}

			err = txCsvWriter.Write([]string{
				network,
				tx.Hash().Hex(),
				block.Hash().Hex(),
				from.Hex(),
				tx.To().Hex(),
				method,
				tx.Value().String(),
				status,
			})
			if err != nil {
				fmt.Printf("Error writing csv row for %s tx %s: %s\n", network, txHash, err.Error())
			}
		}
	}

	fmt.Println("Done analyzing transactions!")
}

func getTxsCsvPath(db *util.FileDB) string {
	return db.Path + "/txs.csv"
}
