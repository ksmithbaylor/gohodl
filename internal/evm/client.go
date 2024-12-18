package evm

import (
	"context"
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/util"
	"github.com/nanmu42/etherscan-api"
)

const AAVE_NATIVE_ASSET = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"

var POLYGON_STAKING_TOKENS = []common.Address{
	common.HexToAddress("0xa6e768fEf2D1aF36c0cfdb276422E7881a83e951"),
	common.HexToAddress("0x503B36441618e61135adE1Fa6Aa8e5345DA7Ce75"),
}

type Client struct {
	Network   Network          // The network the client is for
	Etherscan *EtherscanClient // A client for the etherscan-compatible explorer

	connections     map[string]*ethclient.Client // Maps RPC URL to the corresponding eth client
	symbolCache     map[common.Address]string    // Caches token contract `symbol()` lookups
	decimalCache    map[common.Address]uint8     // Caches token contract `decimals()` lookups
	tokenDataCache  *util.FileDBCollection       // File cache for token data
	internalTxCache *util.FileDBCollection       // File cache for etherscan internal txs
}

func NewClient(network Network) (*Client, error) {
	connections := make(map[string]*ethclient.Client, 0)
	symbolCache := make(map[common.Address]string, 0)
	decimalCache := make(map[common.Address]uint8, 0)
	tokenDataCache := util.NewFileDB("data").NewCollection("token_data")
	internalTxCache := util.NewFileDB("data").NewCollection("internal_txs")
	etherscanClient, err := NewEtherscanClient(network)
	if err != nil {
		return nil, err
	}

	return &Client{
		Network:         network,
		Etherscan:       etherscanClient,
		connections:     connections,
		symbolCache:     symbolCache,
		decimalCache:    decimalCache,
		tokenDataCache:  tokenDataCache,
		internalTxCache: internalTxCache,
	}, nil
}

func (c *Client) Connect() error {
	if len(c.connections) >= QUORUM {
		return nil
	}

	for _, rpc := range c.Network.RPCs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			chainID, err := client.ChainID(context.Background())
			if err == nil && chainID != nil {
				if chainID.Int64() == int64(c.Network.ChainID) {
					c.connections[rpc] = client
					util.Debugf("Connected to %s\n", rpc)
				}
			}
		}
	}

	if len(c.connections) < QUORUM {
		return fmt.Errorf("Connected to less than quorum of %d clients for chain ID %d (only found %d)", QUORUM, c.Network.ChainID, len(c.connections))
	}

	return nil
}

func (c *Client) LatestBlock() (uint64, error) {
	err := c.Connect()
	if err != nil {
		return 0, err
	}

	return ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (uint64, uint64, error) {
		num, err := client.BlockNumber(context.Background())
		return num, num, err
	})
}

func (c *Client) GetTransaction(hash string) (*types.Transaction, error) {
	err := c.Connect()
	if err != nil {
		return nil, err
	}

	return ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (*types.Transaction, string, error) {
		tx, _, err := client.TransactionByHash(context.Background(), common.HexToHash(hash))
		if err != nil {
			return nil, "", err
		}

		json, err := tx.MarshalJSON()
		if err != nil {
			return nil, "", fmt.Errorf("Unable to marshal tx to json: %w", err)
		}

		return tx, common.Bytes2Hex(json), nil
	})
}

func (c *Client) GetTransactionReceipt(hash string) (*types.Receipt, error) {
	err := c.Connect()
	if err != nil {
		return nil, err
	}

	return ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (*types.Receipt, string, error) {
		receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(hash))
		if err != nil {
			return nil, "", err
		}

		json, err := receipt.MarshalJSON()
		if err != nil {
			return nil, "", fmt.Errorf("Unable to marshal tx receipt to json: %w", err)
		}

		return receipt, common.Bytes2Hex(json), nil
	})
}

func (c *Client) GetInternalTransactions(hash string) ([]etherscan.InternalTx, bool, error) {
	cacheKey := fmt.Sprintf("%s-%s", c.Network.Name, hash)

	var txs []etherscan.InternalTx
	cacheFound, err := c.internalTxCache.Read(cacheKey, &txs)
	if err != nil {
		return nil, false, err
	}
	if cacheFound {
		return txs, true, nil
	}

	txs, err = c.Etherscan.GetInternalTransfers(hash)
	if err != nil {
		return nil, false, err
	}

	err = c.internalTxCache.Write(cacheKey, txs)
	if err != nil {
		return nil, false, err
	}

	return txs, false, nil
}

func (c *Client) GetBlock(hash string) (*types.Header, error) {
	err := c.Connect()
	if err != nil {
		return nil, err
	}

	return ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (*types.Header, string, error) {
		blockHeader, err := client.HeaderByHash(context.Background(), common.HexToHash(hash))
		if err != nil {
			return nil, "", err
		}

		json, err := blockHeader.MarshalJSON()
		if err != nil {
			return nil, "", fmt.Errorf("Unable to marshal block header to json: %w", err)
		}

		return blockHeader, common.Bytes2Hex(json), nil
	})
}

func (c *Client) Balance(address common.Address) (core.Amount, error) {
	err := c.Connect()
	if err != nil {
		return core.Amount{}, err
	}

	balance, err := ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (string, string, error) {
		bal, e := client.BalanceAt(context.Background(), address, nil)
		if e != nil {
			return "", "", e
		}
		return bal.String(), bal.String(), nil
	})
	if err != nil {
		return core.Amount{}, fmt.Errorf("Could not get balance: %w", err)
	}

	return core.NewAmountFromAtomicString(c.Network.NativeAsset(), balance)
}

func (c *Client) Erc20Decimals(token common.Address) (uint8, error) {
	if slices.Contains(POLYGON_STAKING_TOKENS, token) {
		return 0, nil
	}

	if dec, ok := c.decimalCache[token]; ok {
		return dec, nil
	}

	cacheKey := fmt.Sprintf("%s-%s-decimals", c.Network.Name, token.Hex())
	var cachedDecimals uint8
	cacheFound, err := c.tokenDataCache.Read(cacheKey, &cachedDecimals)
	if err != nil {
		fmt.Printf("Error reading from token decimal cache: %s\n", err.Error())
	}
	if cacheFound {
		c.decimalCache[token] = cachedDecimals
		return cachedDecimals, nil
	}

	err = c.Connect()
	if err != nil {
		return 0, err
	}

	decimals, err := ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (uint8, uint8, error) {
		result, e := client.CallContract(context.Background(), decimalsCall(token), nil)
		if e != nil {
			return 0, 0, e
		}

		decoded := decodeUint8(result)
		return decoded, decoded, nil
	})
	if err != nil {
		return 0, fmt.Errorf("Could not get token decimals: %w", err)
	}

	err = c.tokenDataCache.Write(cacheKey, decimals)
	if err != nil {
		fmt.Printf("Error writing to token decimal cache: %s\n", err.Error())
	}

	c.decimalCache[token] = decimals
	return decimals, nil
}

func (c *Client) TokenSymbol(token common.Address) (string, error) {
	if slices.Contains(POLYGON_STAKING_TOKENS, token) {
		return "PST", nil
	}

	if sym, ok := c.symbolCache[token]; ok {
		return sym, nil
	}

	cacheKey := fmt.Sprintf("%s-%s-symbol", c.Network.Name, token.Hex())
	var cachedSymbol string
	cacheFound, err := c.tokenDataCache.Read(cacheKey, &cachedSymbol)
	if err != nil {
		fmt.Printf("Error reading from token symbol cache: %s\n", err.Error())
	}
	if cacheFound {
		c.symbolCache[token] = cachedSymbol
		return cachedSymbol, nil
	}

	err = c.Connect()
	if err != nil {
		return "", err
	}

	symbol, err := ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (string, string, error) {
		result, e := client.CallContract(context.Background(), symbolCall(token), nil)
		if e != nil {
			return "", "", e
		}
		sym, e := decodeString(result)
		if e != nil {
			return "", "", e
		}
		return sym, sym, nil
	})
	if err != nil {
		return "", fmt.Errorf("Could not get token symbol: %w", err)
	}

	err = c.tokenDataCache.Write(cacheKey, symbol)
	if err != nil {
		fmt.Printf("Error writing to token symbol cache: %s\n", err.Error())
	}

	c.symbolCache[token] = symbol
	return symbol, nil
}

func (c *Client) NativeAsset() (core.Asset, error) {
	return c.Network.NativeAsset(), nil
}

func (c *Client) OpenTransactionInExplorer(hash string, wait ...bool) {
	c.Network.OpenTransactionInExplorer(hash, wait...)
}

func (c *Client) TokenAsset(token common.Address, instadapp ...bool) (core.Asset, error) {
	if len(instadapp) > 0 && instadapp[0] && token.Hex() == AAVE_NATIVE_ASSET {
		return c.NativeAsset()
	}

	symbol, err := c.TokenSymbol(token)
	if err != nil {
		return core.Asset{}, fmt.Errorf("Could not get token symbol for %s on %s: %w", token, c.Network.Name, err)
	}

	decimals, err := c.Erc20Decimals(token)
	if err != nil {
		return core.Asset{}, fmt.Errorf("Could not get token decimals for %s on %s: %w", token, c.Network.Name, err)
	}

	return c.Network.Erc20TokenAsset(token.Hex(), symbol, decimals), nil
}

func (c *Client) Erc20Balance(token common.Address, address common.Address) (core.Amount, error) {
	err := c.Connect()
	if err != nil {
		return core.Amount{}, err
	}

	decimals, err := c.Erc20Decimals(token)
	if err != nil {
		return core.Amount{}, fmt.Errorf("Could not get token decimals: %w", err)
	}

	symbol, err := c.TokenSymbol(token)
	if err != nil {
		return core.Amount{}, fmt.Errorf("Could not get token symbol: %w", err)
	}

	asset := c.Network.Erc20TokenAsset(token.String(), symbol, decimals)

	balanceStr, err := ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (string, string, error) {
		result, e := client.CallContract(context.Background(), balanceCall(token, address), nil)
		if e != nil {
			return "", "", e
		}
		cents := decodeBigInt(result)
		return cents.String(), cents.String(), nil
	})
	if err != nil {
		return core.Amount{}, fmt.Errorf("Could not get token balance: %w", err)
	}
	balance, err := core.NewAmountFromAtomicString(asset, balanceStr)
	if err != nil {
		return core.Amount{}, fmt.Errorf("Could not get token balance: %w", err)
	}
	return balance, nil
}
