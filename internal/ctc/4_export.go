package ctc

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	"github.com/ksmithbaylor/gohodl/internal/private"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

var CTC_HEADERS = []string{
	"Timestamp (UTC)",
	"Type",
	"Base Currency",
	"Base Amount",
	"Quote Currency (Optional)",
	"Quote Amount (Optional)",
	"Fee Currency (Optional)",
	"Fee Amount (Optional)",
	"From (Optional)",
	"To (Optional)",
	"Blockchain (Optional)",
	"ID (Optional)",
	"Description (Optional)",
	"Reference Price Per Unit (Optional)",
	"Reference Price Currency (Optional)",
}

func ExportTransactions(db *util.FileDB, clients generic.AllNodeClients) {
	if os.Getenv("SKIP_EXPORT") != "" {
		fmt.Println("Skipping transaction export step")
		return
	}

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

	err = ctcCsvWriter.Write(CTC_HEADERS)
	if err != nil {
		fmt.Printf("Error writing CTC CSV headers: %s\n", err.Error())
		return
	}

	rowsToWrite := make([][]string, 0)

	privateImplementation := private.Implementation
	ctcWriter := func(row []string) error {
		rowsToWrite = append(rowsToWrite, row)
		return nil
	}
	txReader := func(network, hash string) (
		*types.Transaction,
		*types.Receipt,
		*types.Header,
		error,
	) {
		return readTransactionBundle(db, network, hash)
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

		// Skip header row
		if row[0] == "network" {
			continue
		}

		totalTxs++
		info := evm.TxInfo{
			Network:   row[0],
			Hash:      row[1],
			BlockHash: row[2],
			From:      row[3],
			To:        row[4],
			Method:    row[5],
			Value:     row[6],
			Success:   row[7] == "success",
		}

		client, ok := clients[info.Network]
		if !ok {
			fmt.Printf("No client for network %s\n", info.Network)
			return
		}
		evmClient, ok := client.(*evm.Client)
		if !ok {
			fmt.Printf("Non-EVM networks (like %s) not implemented yet\n", info.Network)
			return
		}

		handled, err := privateImplementation.HandleTransaction(&info, evmClient, txReader, ctcWriter)
		if handled {
			handledTxs++
		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	sort.Sort(byFirstField(rowsToWrite))

	err = ctcCsvWriter.WriteAll(rowsToWrite)
	if err != nil {
		fmt.Printf("Error writing CTC CSV: %s\n", err.Error())
	}

	fmt.Printf("%d transactions handled out of %d (%.2f%%)\n", handledTxs, totalTxs, 100.0*float32(handledTxs)/float32(totalTxs))
	fmt.Println("Finished reading txs from CSV!")
}

type byFirstField [][]string

func (data byFirstField) Len() int {
	return len(data)
}

func (data byFirstField) Swap(i, j int) {
	data[i], data[j] = data[j], data[i]
}

func (data byFirstField) Less(i, j int) bool {
	return data[i][0] < data[j][0]
}

func getCtcCsvPath(db *util.FileDB) string {
	return db.Path + "/ctc.csv"
}
