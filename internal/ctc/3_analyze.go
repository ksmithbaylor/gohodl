package ctc

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func AnalyzeTransactions(db *util.FileDB, txHashes map[string][]string) {
	if os.Getenv("SKIP_ANALYZE") != "" {
		fmt.Println("Skipping transaction analyze step")
		return
	}

	txsDB, txsFound := db.Collections["txs"]
	if !txsFound || txsDB == nil {
		fmt.Println("Cannot analyze transactions without first fetching them")
		return
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
		"timestamp",
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

	for network, networkTxHashes := range txHashes {
		for _, txHash := range networkTxHashes {
			tx, receipt, block, err := readTransactionBundle(db, network, txHash)
			if err != nil {
				fmt.Printf("Could not read %s transaction %s: %s\n", network, txHash, err.Error())
				continue
			}

			if block == nil {
				panic("nil block for " + network + " / " + receipt.BlockHash.Hex())
			}

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

			to := evm.ZERO_ADDRESS
			if tx.To() != nil {
				to = tx.To().Hex()
			}

			err = txCsvWriter.Write([]string{
				strconv.Itoa(int(block.Time)),
				network,
				txHash,
				block.Hash().Hex(),
				from.Hex(),
				to,
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
