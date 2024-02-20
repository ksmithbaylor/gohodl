package eth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/types"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type Client struct {
	Network types.EthNetwork // The network the client is for

	connections  map[string]*ethclient.Client // Maps RPC URL to the corresponding eth client
	symbolCache  map[string]string            // Caches token contract `symbol()` lookups
	decimalCache map[string]uint8             // Caches token contract `decimals()` lookups
}

func NewClient(network types.EthNetwork) (*Client, error) {
	connections := make(map[string]*ethclient.Client, 0)
	symbolCache := make(map[string]string, 0)
	decimalCache := make(map[string]uint8, 0)

	for _, rpc := range network.Config.RPCs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			chainID, err := client.ChainID(context.Background())
			if err == nil && chainID != nil {
				if chainID.Int64() == int64(network.Config.ChainID) {
					connections[rpc] = client
					util.Debugf("Connected to %s\n", rpc)
				}
			}
		}
	}

	if len(connections) < QUORUM {
		return nil, fmt.Errorf("Connected to less than quorum of %d clients for chain ID %d (only found %d)", QUORUM, network.Config.ChainID, len(connections))
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

func (c *Client) Balance(a types.EthAddress) (types.Amount, error) {
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
