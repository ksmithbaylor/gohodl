package ctc

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	handlers "github.com/ksmithbaylor/gohodl/internal/handlers/kevin"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

var END_OF_2023 = 1704067199
var END_OF_2024 = 1735689599

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

	err = ctcCsvWriter.Write(ctc_util.CTC_HEADERS)
	if err != nil {
		fmt.Printf("Error writing CTC CSV headers: %s\n", err.Error())
		return
	}

	rowsToWrite := make([][]string, 0)

	privateImplementation := handlers.Implementation
	ctcWriter := func(rows ...[]string) error {
		rowsToWrite = append(rowsToWrite, rows...)
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
	unhandled := 0

	for {
		row, err := txCsvReader.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading txs CSV row: %s\n", err.Error())
			}
			break
		}

		// Skip header row
		if row[0] == "timestamp" {
			continue
		}

		timestamp, err := strconv.Atoi(row[0])
		if err != nil {
			panic("Invalid timestamp: " + row[0])
		}

		if timestamp <= END_OF_2023 {
			continue
		}

		totalTxs++
		info := evm.TxInfo{
			Time:      timestamp,
			Network:   row[1],
			Hash:      row[2],
			BlockHash: row[3],
			From:      row[4],
			To:        row[5],
			Method:    row[6],
			Value:     row[7],
			Success:   row[8] == "success",
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
			if err == nil {
				handledTxs++
			} else if err == handlers.NOT_HANDLED {
				fmt.Println("NOT_HANDLED:", info.Network, info.Hash)
				unhandled++
			}
		}
		if err != nil && err != handlers.NOT_HANDLED {
			fmt.Println(err.Error())
			return
		}
	}

	sort.Slice(rowsToWrite, func(i, j int) bool {
		return rowsToWrite[i][0] < rowsToWrite[j][0]
	})

	err = ctcCsvWriter.WriteAll(rowsToWrite)
	if err != nil {
		fmt.Printf("Error writing CTC CSV: %s\n", err.Error())
	}

	fmt.Printf("%d transactions handled out of %d (%.2f%%), %d remaining\n", handledTxs, totalTxs, 100.0*float32(handledTxs)/float32(totalTxs), totalTxs-handledTxs)
	if unhandled > 0 {
		fmt.Printf("%d transactions temporarily not handled (will be %.2f%% when done)\n", unhandled, 100.0*float32(handledTxs+unhandled)/float32(totalTxs))
	}
	fmt.Println("Finished exporting transactions!")
}

func getCtcCsvPath(db *util.FileDB) string {
	return db.Path + "/ctc.csv"
}
