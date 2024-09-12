package evm

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type Client struct {
	Network Network // The network the client is for

	connections  map[string]*ethclient.Client // Maps RPC URL to the corresponding eth client
	symbolCache  map[common.Address]string    // Caches token contract `symbol()` lookups
	decimalCache map[common.Address]uint8     // Caches token contract `decimals()` lookups
}

func NewClient(network Network) (*Client, error) {
	connections := make(map[string]*ethclient.Client, 0)

	for _, rpc := range network.RPCs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			chainID, err := client.ChainID(context.Background())
			if err == nil && chainID != nil {
				if chainID.Int64() == int64(network.ChainID) {
					connections[rpc] = client
					util.Debugf("Connected to %s\n", rpc)
				}
			}
		}
	}

	if len(connections) < QUORUM {
		return nil, fmt.Errorf("Connected to less than quorum of %d clients for chain ID %d (only found %d)", QUORUM, network.ChainID, len(connections))
	}

	symbolCache := make(map[common.Address]string, 0)
	decimalCache := make(map[common.Address]uint8, 0)

	return &Client{
		Network:      network,
		connections:  connections,
		symbolCache:  symbolCache,
		decimalCache: decimalCache,
	}, nil
}

func (c *Client) LatestBlock() (uint64, error) {
	return ensureAgreementWithRetry(c.connections, func(client *ethclient.Client) (uint64, uint64, error) {
		num, err := client.BlockNumber(context.Background())
		return num, num, err
	})
}

func (c *Client) Balance(address common.Address) (core.Amount, error) {
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
	if dec, ok := c.decimalCache[token]; ok {
		return dec, nil
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
	c.decimalCache[token] = decimals
	return decimals, nil
}

func (c *Client) TokenSymbol(token common.Address) (string, error) {
	if sym, ok := c.symbolCache[token]; ok {
		return sym, nil
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
	c.symbolCache[token] = symbol
	return symbol, nil
}

func (c *Client) Erc20Balance(token common.Address, address common.Address) (core.Amount, error) {
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
