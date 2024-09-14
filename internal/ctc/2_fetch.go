package ctc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func FetchTransactions(db *util.FileDB, clients generic.AllNodeClients) {
	txsDB := db.NewCollection("txs")
	receiptsDB := db.NewCollection("receipts")

	if os.Getenv("SKIP_FETCH") != "" {
		fmt.Println("Skipping transaction fetching step")
		return
	}

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

	txsToFetch := make(map[string][]string)

	// Get a list of transaction hashes to fetch for each network
	for _, key := range txHashesKeys {
		// Only actual cache keys
		if !strings.Contains(key, "-") {
			continue
		}

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

		txsToFetch[entry.Network] = util.UniqueItems(txsToFetch[entry.Network], entry.Txs)
	}

	// Fetch all transactions in parallel across networks
	var wg sync.WaitGroup
	for network, txs := range txsToFetch {
		client := clients[network]
		if client == nil {
			fmt.Printf("No client found for %s\n", network)
			continue
		}

		evmClient, ok := client.(*evm.Client)
		if !ok {
			fmt.Printf("Non-EVM networks (like %s) not implemented yet\n", network)
			continue
		}

		wg.Add(1)
		go fetch(&wg, evmClient, txsDB, receiptsDB, network, txs)
	}

	wg.Wait()
	fmt.Println("Done fetching all transactions!")
}

func fetch(
	wg *sync.WaitGroup,
	client *evm.Client,
	txsDB *util.FileDBCollection,
	receiptsDB *util.FileDBCollection,
	network string,
	txs []string,
) {
	defer wg.Done()

	fmt.Printf("Fetching %d transactions on %s\n", len(txs), network)

	for _, txHash := range txs {
		cacheKey := fmt.Sprintf("%s-%s", network, txHash)

		fetchTransaction(client, txsDB, cacheKey, network, txHash)
		fetchTransactionReceipt(client, receiptsDB, cacheKey, network, txHash)
	}

	fmt.Printf("Done fetching transactions for %s\n", network)
}

func fetchTransaction(
	client *evm.Client,
	txsDB *util.FileDBCollection,
	cacheKey string,
	network string,
	txHash string,
) {
	var cachedTx types.Transaction
	cacheFound, err := txsDB.Read(cacheKey, &cachedTx)
	if err != nil {
		fmt.Printf("Error reading cache for %s: %s\n", cacheKey, err.Error())
		return
	}

	if cacheFound {
		checkTransaction(&cachedTx, network, txHash)
		return
	}

	tx, err := client.GetTransaction(txHash)
	if err != nil {
		fmt.Printf("Error fetching %s tx %s: %s\n", network, txHash, err.Error())
		return
	}
	if tx == nil {
		fmt.Printf("Nil response for %s tx %s\n", network, txHash)
		return
	}

	if tx.IsDepositTx() {
		newTx := types.NewTransaction(tx.Nonce(), *tx.To(), tx.Value(), tx.Gas(), tx.GasPrice(), tx.Data())
		newTxJson, err := newTx.MarshalJSON()
		if err != nil {
			fmt.Printf("Error marshaling original deposit tx for %s on %s: %s\n", txHash, network, err.Error())
			return
		}

		var jsonToModify map[string]json.RawMessage
		err = json.Unmarshal(newTxJson, &jsonToModify)
		if err != nil {
			fmt.Printf("Error unmarshaling replacement deposit tx for %s on %s: %s\n", txHash, network, err.Error())
			return
		}

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			fmt.Printf("Error getting sender for deposit tx %s on %s: %s\n", txHash, network, err.Error())
			return
		}
		jsonToModify["from"] = json.RawMessage("\"" + strings.ToLower(from.Hex()) + "\"")
		jsonToModify["type"] = json.RawMessage("\"0x7e\"")
		jsonToModify["hash"] = json.RawMessage("\"" + txHash + "\"")
		jsonToModify["sourceHash"] = json.RawMessage("\"" + strings.ToLower(tx.SourceHash().Hex()) + "\"")

		finalJson, err := json.Marshal(jsonToModify)
		if err != nil {
			fmt.Printf("Error marshaling replacement deposit tx for %s on %s: %s\n", txHash, network, err.Error())
			return
		}

		err = txsDB.WriteRaw(cacheKey, finalJson)
		if err != nil {
			fmt.Printf("Error caching deposit transaction %s: %s\n", cacheKey, err.Error())
		}

		fmt.Printf("Fetched %s deposit transaction %s\n", network, txHash)
		return
	}

	checkTransaction(tx, network, txHash)

	err = txsDB.Write(cacheKey, tx)
	if err != nil {
		fmt.Printf("Error caching transaction %s: %s\n", cacheKey, err.Error())
	}

	fmt.Printf("Fetched %s transaction %s\n", network, txHash)
}

func fetchTransactionReceipt(
	client *evm.Client,
	receiptsDB *util.FileDBCollection,
	cacheKey string,
	network string,
	txHash string,
) {
	var cachedReceipt types.Receipt
	cacheFound, err := receiptsDB.Read(cacheKey, &cachedReceipt)
	if err != nil {
		fmt.Printf("Error reading cache for %s: %s\n", cacheKey, err.Error())
		fmt.Printf("%#v\n", cachedReceipt)
		return
	}

	if cacheFound {
		checkReceipt(&cachedReceipt, network, txHash)
		return
	}

	receipt, err := client.GetTransactionReceipt(txHash)
	if err != nil {
		fmt.Printf("Error fetching %s tx %s receipt: %s\n", network, txHash, err.Error())
		return
	}
	if receipt == nil {
		fmt.Printf("Nil response for %s tx %s receipt\n", network, txHash)
		return
	}

	checkReceipt(receipt, network, txHash)

	err = receiptsDB.Write(cacheKey, receipt)
	if err != nil {
		fmt.Printf("Error caching transaction %s receipt: %s\n", cacheKey, err.Error())
	}

	fmt.Printf("Fetched %s transaction %s receipt\n", network, txHash)
}

func checkReceipt(receipt *types.Receipt, network string, txHash string) {
	if receipt.GasUsed == 0 && len(receipt.Logs) == 0 {
		fmt.Printf("--- Receipt for %s tx %s appears to be empty\n", network, txHash)
	}
}

func checkTransaction(tx *types.Transaction, network string, txHash string) {
	if len(tx.Hash().String()) < 4 {
		fmt.Printf("--- Transaction %s on %s appears to be invalid\n", txHash, network)
	}
}
