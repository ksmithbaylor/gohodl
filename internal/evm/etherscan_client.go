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
	// TODO
	return []string{}, nil
}

// func (c *EtherscanClient) request(url string) {
//   response, err := c.client.Get(url)
//   if err != nil {
//     return nil, err
//   }
//
//   // TODO make this generic and parse the response into a struct specified
// }
//
// func (c *EtherscanClient) getAllPages(url string) {
//   c.client.Get(url)
//
//   // TODO cycle through pages and return a slice of whatever
// }
