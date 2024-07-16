package evm

import (
	"fmt"
	"net/http"
	"time"
)

type EtherscanClient struct {
	BaseURL           string
	APIKey            string
	RequestsPerSecond uint

	network     Network
	http        *http.Client
	lastRequest time.Time
}

func NewEtherscanClient(network Network) (*EtherscanClient, error) {
	e := network.Etherscan

	if e.URL == "" || e.Key == "" {
		return nil, fmt.Errorf("Invalid etherscan config for %s", network.Name)
	}

	if e.RPS == 0 {
		e.RPS = 5
	}

	httpClient := http.Client{}

	return &EtherscanClient{
		BaseURL:           e.URL,
		APIKey:            e.Key,
		RequestsPerSecond: e.RPS,
		network:           network,
		http:              &httpClient,
	}, nil
}

func (c *EtherscanClient) GetAllTransactionHashes(address string) ([]string, error) {
	hashes := make(map[string]any)

	normalHashes, err := c.getNormalTransactionHashes(address)
	if err != nil {
		return nil, fmt.Errorf("Could not get normal transactions for %s: %w", address, err)
	}
	for _, hash := range normalHashes {
		hashes[hash] = struct{}{}
	}

	hashesList := make([]string, 0, len(hashes))
	for hash := range hashes {
		hashesList = append(hashesList, hash)
	}

	return hashesList, nil
}

func (c *EtherscanClient) getNormalTransactionHashes(address string) ([]string, error) {
	hashes := make([]string, 0)

	hashes = append(hashes, "asdf")
	hashes = append(hashes, "foo")
	hashes = append(hashes, "asdf")

	return hashes, nil
}
