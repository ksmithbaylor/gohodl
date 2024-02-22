package eth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/types"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type Client struct {
	Network types.EvmNetwork // The network the client is for

	connections  map[string]*ethclient.Client // Maps RPC URL to the corresponding eth client
	symbolCache  map[string]string            // Caches token contract `symbol()` lookups
	decimalCache map[string]uint8             // Caches token contract `decimals()` lookups
}

func NewClient(network types.EvmNetwork) (*Client, error) {
	connections := make(map[string]*ethclient.Client, 0)
	symbolCache := make(map[string]string, 0)
	decimalCache := make(map[string]uint8, 0)

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

	return &Client{
		Network:      network,
		connections:  connections,
		symbolCache:  symbolCache,
		decimalCache: decimalCache,
	}, nil
}

func (c *Client) LatestBlock() (uint64, error) {
	return withRetry(c.connections, func(client *ethclient.Client) (uint64, error) {
		return client.BlockNumber(context.Background())
	})
}

func (c *Client) Balance(a types.EvmAddress) (types.Amount, error) {
	address := a.ToGeth()

	balance, err := withRetry(c.connections, func(client *ethclient.Client) (string, error) {
		bal, e := client.BalanceAt(context.Background(), address, nil)
		if e != nil {
			return "", e
		}
		return bal.String(), nil
	})
	if err != nil {
		return types.Amount{}, fmt.Errorf("Could not get balance: %w", err)
	}

	return types.NewAmountFromCentsString(c.Network.NativeEvmAsset(), balance)
}

func (c *Client) Erc20Decimals(token types.EvmAddress) (uint8, error) {
	if dec, ok := c.decimalCache[token.String()]; ok {
		return dec, nil
	}

	contract := token.ToGeth()

	decimals, err := withRetry(c.connections, func(client *ethclient.Client) (uint8, error) {
		result, e := client.CallContract(context.Background(), decimalsCall(contract), nil)
		if e != nil {
			return 0, e
		}
		return decodeUint8(result), nil
	})
	if err != nil {
		return 0, fmt.Errorf("Could not get token decimals: %w", err)
	}
	c.decimalCache[token.String()] = decimals
	return decimals, nil
}

func (c *Client) TokenSymbol(token types.EvmAddress) (string, error) {
	if sym, ok := c.symbolCache[token.String()]; ok {
		return sym, nil
	}

	contract := token.ToGeth()

	symbol, err := withRetry(c.connections, func(client *ethclient.Client) (string, error) {
		result, e := client.CallContract(context.Background(), symbolCall(contract), nil)
		if e != nil {
			return "", e
		}
		sym, e := decodeString(result)
		if e != nil {
			return "", e
		}
		return sym, nil
	})
	if err != nil {
		return "", fmt.Errorf("Could not get token symbol: %w", err)
	}
	c.symbolCache[token.String()] = symbol
	return symbol, nil
}

func (c *Client) Erc20Balance(token types.EvmAddress, a types.EvmAddress) (types.Amount, error) {
	decimals, err := c.Erc20Decimals(token)
	if err != nil {
		return types.Amount{}, fmt.Errorf("Could not get token balance: %w", err)
	}

	symbol, err := c.TokenSymbol(token)
	if err != nil {
		return types.Amount{}, fmt.Errorf("Could not get token symbol: %w", err)
	}

	asset := c.Network.Erc20TokenAsset(token.String(), symbol, decimals)

	contract := token.ToGeth()
	address := a.ToGeth()

	balanceStr, err := withRetry(c.connections, func(client *ethclient.Client) (string, error) {
		result, e := client.CallContract(context.Background(), balanceCall(contract, address), nil)
		if e != nil {
			return "", e
		}
		cents := decodeBigInt(result)
		return cents.String(), nil
	})
	if err != nil {
		return types.Amount{}, fmt.Errorf("Could not get token balance: %w", err)
	}
	balance, err := types.NewAmountFromCentsString(asset, balanceStr)
	if err != nil {
		return types.Amount{}, fmt.Errorf("Could not get token balance: %w", err)
	}
	return balance, nil
}
