package ctc

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	handlers "github.com/ksmithbaylor/gohodl/internal/handlers/kevin"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

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
		for _, row := range rows {
			rowsToWrite = append(rowsToWrite, row)
		}
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
			if err == nil {
				handledTxs++
			} else if err == handlers.NOT_HANDLED {
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

	fmt.Printf("%d transactions handled out of %d (%.2f%%)\n", handledTxs, totalTxs, 100.0*float32(handledTxs)/float32(totalTxs))
	if unhandled > 0 {
		fmt.Printf("%d transactions temporarily not handled (will be %.2f%% when done)\n", unhandled, 100.0*float32(handledTxs+unhandled)/float32(totalTxs))
	}
	fmt.Println("Finished exporting transactions!")
}

func getCtcCsvPath(db *util.FileDB) string {
	return db.Path + "/ctc.csv"
}
