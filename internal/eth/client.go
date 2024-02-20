package eth

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/types"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type Client struct {
	Network types.EthNetwork

	connections map[string]*ethclient.Client
}

func NewClient(network types.EthNetwork) (*Client, error) {
	connections := make(map[string]*ethclient.Client, 0)

	for _, rpc := range network.Config.RPCs {
		client, err := ethclient.Dial(rpc)
		if err == nil {
			connections[rpc] = client
			util.Debugf("Connected to %s\n", rpc)
		}
	}

	if len(connections) < QUORUM {
		return nil, fmt.Errorf("Connected to less than quorum of %d clients for chain ID %d (only found %d)", QUORUM, network.Config.ChainID, len(connections))
	}

	return &Client{
		Network:     network,
		connections: connections,
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
