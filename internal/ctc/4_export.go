package ctc

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/private"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func ExportTransactions(db *util.FileDB) {
	txCsvFile, err := os.Open(getTxsCsvPath(db))
	if txCsvFile == nil || err != nil {
		fmt.Println("Transactions CSV not written yet, please run analyze step")
		return
	}
	defer txCsvFile.Close()

	ctcCsvFile, err := os.Create(getCtcCsvPath(db))
	if err != nil {
		fmt.Printf("Error creating CTC CSV file: %s\n", err.Error())
		return
	}
	defer ctcCsvFile.Close()

	txCsvReader := csv.NewReader(txCsvFile)
	ctcCsvWriter := csv.NewWriter(ctcCsvFile)
	defer ctcCsvWriter.Flush()

	privateImplementation := private.Implementation
	ctcWriter := ctcCsvWriter.Write
	txReader := func(network, hash string) (*types.Transaction, *types.Receipt, error) {
		return readTransaction(db, network, hash)
	}

	totalTxs := 0
	handledTxs := 0

	for {
		row, err := txCsvReader.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading txs CSV row: %s\n", err.Error())
			}
			break
		}

		totalTxs++
		info := evm.TxInfo{
			Network: row[0],
			Hash:    row[1],
			From:    row[2],
			To:      row[3],
			Method:  row[4],
			Value:   row[5],
			Success: row[6] == "success",
		}

		handled, err := privateImplementation.HandleTransaction(&info, txReader, ctcWriter)
		if handled {
			handledTxs++
		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Printf("%d transactions handled out of %d (%.2f%%)\n", handledTxs, totalTxs, 100.0*float32(handledTxs)/float32(totalTxs))
	fmt.Println("Finished reading txs from CSV!")
}

func readTransaction(db *util.FileDB, network, hash string) (*types.Transaction, *types.Receipt, error) {
	txsDB, txsFound := db.Collections["txs"]
	receiptsDB, receiptsFound := db.Collections["receipts"]
	if !txsFound || txsDB == nil || !receiptsFound || receiptsDB == nil {
		return nil, nil, errors.New("Cannot export transactions without first fetching them")
	}

	key := fmt.Sprintf("%s-%s", network, hash)

	var tx types.Transaction
	found, err := txsDB.Read(key, &tx)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading tx for %s: %w", key, err)
	}
	if !found {
		return nil, nil, fmt.Errorf("Transaction not found: %s", key)
	}

	var receipt types.Receipt
	found, err = receiptsDB.Read(key, &receipt)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading tx receipt for %s: %w", key, err)
	}
	if !found {
		return nil, nil, fmt.Errorf("Transaction receipt not found: %s", key)
	}

	return &tx, &receipt, nil
}

func getCtcCsvPath(db *util.FileDB) string {
	return db.Path + "/ctc.csv"
}
