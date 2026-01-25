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

func FetchTransactions(db *util.FileDB, clients generic.AllNodeClients) map[string][]string {
	txsDB := db.NewCollection("txs")
	receiptsDB := db.NewCollection("receipts")
	blocksDB := db.NewCollection("blocks")

	txHashesDB, found := db.Collections["evm_tx_hashes"]
	if !found || txHashesDB == nil {
		fmt.Println("Cannot fetch transactions without first running identification step")
		return nil
	}

	txHashesKeys, err := txHashesDB.List()
	if err != nil {
		fmt.Printf("Error listing keys in tx hashes collection: %s\n", err.Error())
		return nil
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

	if os.Getenv("SKIP_FETCH") != "" {
		fmt.Println("Skipping transaction fetching step")
		return txsToFetch
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

		if evmClient.Network.GetDeprecated() {
			fmt.Printf("Skipping fetch step for deprecated network %s\n", network)
			continue
		}

		wg.Add(1)
		go fetch(&wg, evmClient, txsDB, receiptsDB, blocksDB, network, txs)
	}

	wg.Wait()
	fmt.Println("Done fetching all transactions!")

	return txsToFetch
}

func fetch(
	wg *sync.WaitGroup,
	client *evm.Client,
	txsDB *util.FileDBCollection,
	receiptsDB *util.FileDBCollection,
	blocksDB *util.FileDBCollection,
	network string,
	txs []string,
) {
	defer wg.Done()

	fmt.Printf("Fetching %d transactions on %s\n", len(txs), network)

	unfetched := txs

	for {
		retryTx := make([]string, 0)
		retryReceipt := make([]string, 0)
		retryBlock := make([]string, 0)
		retryInternal := make([]string, 0)

		for _, txHash := range unfetched {
			cacheKey := fmt.Sprintf("%s-%s", network, txHash)

			txSuccess := fetchTransaction(client, txsDB, cacheKey, network, txHash)
			receiptSuccess, blockHash := fetchTransactionReceipt(client, receiptsDB, cacheKey, network, txHash)
			blockSuccess := fetchBlock(client, blocksDB, network, blockHash)
			internalSuccess := fetchInternalTxs(client, network, txHash)

			if !txSuccess {
				retryTx = append(retryTx, txHash)
			}
			if !receiptSuccess {
				retryReceipt = append(retryReceipt, txHash)
			}
			if !blockSuccess {
				retryBlock = append(retryBlock, txHash)
			}
			if !internalSuccess {
				retryInternal = append(retryInternal, txHash)
			}
		}

		if len(retryTx) > 0 || len(retryReceipt) > 0 || len(retryBlock) > 0 || len(retryInternal) > 0 {
			unfetched = util.UniqueItems(retryTx, retryReceipt, retryBlock, retryInternal)
		} else {
			break
		}
	}

	fmt.Printf("Done fetching transactions for %s\n", network)
}

func fetchTransaction(
	client *evm.Client,
	txsDB *util.FileDBCollection,
	cacheKey string,
	network string,
	txHash string,
) bool {
	var cachedTx types.Transaction
	cacheFound, err := txsDB.Read(cacheKey, &cachedTx)
	if err != nil {
		fmt.Printf("Error reading tx cache for %s: %s\n", cacheKey, err.Error())
		return true
	}

	if cacheFound {
		checkTransaction(&cachedTx, network, txHash)
		return true
	}

	tx, err := client.GetTransaction(txHash)
	if err != nil {
		fmt.Printf("Error fetching %s tx %s: %s\n", network, txHash, err.Error())
		return false
	}
	if tx == nil {
		fmt.Printf("Nil response for %s tx %s\n", network, txHash)
		return false
	}

	if tx.IsDepositTx() {
		newTx := types.NewTransaction(tx.Nonce(), *tx.To(), tx.Value(), tx.Gas(), tx.GasPrice(), tx.Data())
		newTxJson, err := newTx.MarshalJSON()
		if err != nil {
			fmt.Printf("Error marshaling original deposit tx for %s on %s: %s\n", txHash, network, err.Error())
			return false
		}

		var jsonToModify map[string]json.RawMessage
		err = json.Unmarshal(newTxJson, &jsonToModify)
		if err != nil {
			fmt.Printf("Error unmarshaling replacement deposit tx for %s on %s: %s\n", txHash, network, err.Error())
			return false
		}

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			fmt.Printf("Error getting sender for deposit tx %s on %s: %s\n", txHash, network, err.Error())
			return false
		}
		jsonToModify["from"] = json.RawMessage("\"" + strings.ToLower(from.Hex()) + "\"")
		jsonToModify["type"] = json.RawMessage("\"0x7e\"")
		jsonToModify["hash"] = json.RawMessage("\"" + txHash + "\"")
		jsonToModify["sourceHash"] = json.RawMessage("\"" + strings.ToLower(tx.SourceHash().Hex()) + "\"")

		finalJson, err := json.Marshal(jsonToModify)
		if err != nil {
			fmt.Printf("Error marshaling replacement deposit tx for %s on %s: %s\n", txHash, network, err.Error())
			return false
		}

		err = txsDB.WriteRaw(cacheKey, finalJson)
		if err != nil {
			fmt.Printf("Error caching deposit transaction %s: %s\n", cacheKey, err.Error())
		}

		fmt.Printf("Fetched %s deposit transaction %s\n", network, txHash)
		return true
	}

	checkTransaction(tx, network, txHash)

	err = txsDB.Write(cacheKey, tx)
	if err != nil {
		fmt.Printf("Error caching transaction %s: %s\n", cacheKey, err.Error())
	}

	fmt.Printf("Fetched %s transaction %s\n", network, txHash)
	return true
}

func fetchTransactionReceipt(
	client *evm.Client,
	receiptsDB *util.FileDBCollection,
	cacheKey string,
	network string,
	txHash string,
) (bool, string) {
	var cachedReceipt types.Receipt
	cacheFound, err := receiptsDB.Read(cacheKey, &cachedReceipt)
	if err != nil {
		fmt.Printf("Error reading receipt cache for %s: %s\n", cacheKey, err.Error())
		fmt.Printf("%#v\n", cachedReceipt)
		return true, "<invalid-cache>"
	}

	if cacheFound {
		checkReceipt(&cachedReceipt, network, txHash)
		return true, cachedReceipt.BlockHash.String()
	}

	receipt, err := client.GetTransactionReceipt(txHash)
	if err != nil {
		fmt.Printf("Error fetching %s tx %s receipt: %s\n", network, txHash, err.Error())
		return false, ""
	}
	if receipt == nil {
		fmt.Printf("Nil response for %s tx %s receipt\n", network, txHash)
		return false, ""
	}

	checkReceipt(receipt, network, txHash)

	err = receiptsDB.Write(cacheKey, receipt)
	if err != nil {
		fmt.Printf("Error caching transaction %s receipt: %s\n", cacheKey, err.Error())
	}

	fmt.Printf("Fetched %s transaction %s receipt\n", network, txHash)
	return true, receipt.BlockHash.String()
}

func fetchBlock(
	client *evm.Client,
	blocksDB *util.FileDBCollection,
	network string,
	blockHash string,
) bool {
	cacheKey := fmt.Sprintf("%s-%s", network, blockHash)

	var cachedBlock types.Header
	cacheFound, err := blocksDB.Read(cacheKey, &cachedBlock)
	if err != nil {
		fmt.Printf("Error reading cache for block %s: %s\n", cacheKey, err.Error())
		return true
	}

	if cacheFound {
		checkBlock(&cachedBlock, network, blockHash)
		return true
	}

	block, err := client.GetBlock(blockHash)
	if err != nil {
		fmt.Printf("Error fetching %s block %s: %s\n", network, blockHash, err.Error())
		return false
	}
	if block == nil {
		fmt.Printf("Nil response for %s block %s\n", network, blockHash)
		return false
	}

	checkBlock(block, network, blockHash)

	err = blocksDB.Write(cacheKey, block)
	if err != nil {
		fmt.Printf("Error caching block %s: %s\n", cacheKey, err.Error())
	}

	fmt.Printf("Fetched %s block %s\n", network, blockHash)
	return true
}

func fetchInternalTxs(client *evm.Client, network, txHash string) bool {
	_, cached, err := client.GetInternalTransactions(txHash)
	if err != nil {
		fmt.Printf("Error fetching internal txs for %s tx %s: %s\n", network, txHash, err.Error())
		return false
	}

	if !cached {
		fmt.Printf("Fetched internal txs for %s tx %s\n", network, txHash)
	}
	return true
}

func checkTransaction(tx *types.Transaction, network string, txHash string) {
	if tx == nil || len(tx.Hash().String()) < 4 {
		fmt.Printf("--- Transaction %s on %s appears to be invalid\n", txHash, network)
	}
}

func checkReceipt(receipt *types.Receipt, network string, txHash string) {
	if receipt == nil || receipt.GasUsed == 0 && len(receipt.Logs) == 0 {
		fmt.Printf("--- Receipt for %s tx %s appears to be empty\n", network, txHash)
	}
}

func checkBlock(block *types.Header, network string, blockHash string) {
	if block == nil || block.SanityCheck() != nil {
		fmt.Printf("--- Block %s on %s appears to be empty\n", blockHash, network)
	}
}
