package evm

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ksmithbaylor/gohodl/internal/util"
	"github.com/nanmu42/etherscan-api"
)

const PER_PAGE = 100

type EtherscanClient struct {
	network Network
	client  *etherscan.Client
	ticks   chan struct{}
	rps     uint
}

func NewEtherscanClient(network Network) (*EtherscanClient, error) {
	e := network.Etherscan

	if e.URL == "" {
		return nil, fmt.Errorf("Invalid etherscan config for %s", network.Name)
	}

	if e.RPS == 0 {
		e.RPS = 5
	}

	client := EtherscanClient{
		network: network,
		ticks:   make(chan struct{}),
		rps:     e.RPS,
	}

	client.client = etherscan.NewCustomized(etherscan.Customization{
		BaseURL: e.URL + "?",
		Key:     e.Key,
		BeforeRequest: func(_, _ string, params map[string]any) error {
			client.wait(params)
			return nil
		},
	})

	go func() {
		client.ticks <- struct{}{}
	}()

	return &client, nil
}

type labeledGetter struct {
	label  string
	getTxs func(string, int, int, bool) ([]any, error)
}

func (c *EtherscanClient) GetAllTransactionHashes(address string, startBlock, endBlock *int) ([]string, error) {
	s := 0
	if startBlock != nil {
		s = *startBlock
	}
	e := 0
	if endBlock != nil {
		e = *endBlock
	}
	fmt.Printf("Getting txs for %s on %s from %d to %d\n", address, c.network.Name, s, e)

	return c.getAllTypesOfTransactionHash(address,
		labeledGetter{"normal", withStartEnd(startBlock, endBlock, withAnyReturn(c.client.NormalTxByAddress))},
		labeledGetter{"internal", withStartEnd(startBlock, endBlock, withAnyReturn(c.client.InternalTxByAddress))},
		labeledGetter{"erc20", withStartEnd(startBlock, endBlock, withAnyReturn(withAnyContractAddress(c.client.ERC20Transfers)))},
		labeledGetter{"erc721", withStartEnd(startBlock, endBlock, withAnyReturn(withAnyContractAddress(c.client.ERC721Transfers)))},
		labeledGetter{"erc1155", withStartEnd(startBlock, endBlock, withAnyReturn(withAnyContractAddress(c.client.ERC1155Transfers)))},
	)
}

func (c *EtherscanClient) wait(params map[string]any) {
	<-c.ticks
	go func() {
		time.Sleep(time.Duration(1000/(c.rps)) * time.Millisecond)
		c.ticks <- struct{}{}
	}()
}

func (c *EtherscanClient) getAllTypesOfTransactionHash(address string, txGetters ...labeledGetter) ([]string, error) {
	hashLists := make([][]string, len(txGetters))

	for i, getter := range txGetters {
		hashes, err := c.getTransactionHashes(address, getter.label, getter.getTxs)
		if err != nil {
			return nil, fmt.Errorf("Could not get %s txs for %s: %w", getter.label, address, err)
		}
		hashLists[i] = hashes
	}

	return util.UniqueItems(hashLists...), nil
}

func (c *EtherscanClient) getTransactionHashes(
	address string,
	label string,
	getTxs func(string, int, int, bool) ([]any, error),
) ([]string, error) {
	hashes := make([]string, 0)

	if c.network.Name == "fantom" && label == "erc1155" {
		// FtmScan doesn't support this endpoint, so if there are any, they are
		// lost to the mists of time.
		return hashes, nil
	}

	page := 1
	rateLimitWaitSeconds := 1

	for {
		txs, err := getTxs(address, page, PER_PAGE, true)
		if err != nil {
			if strings.Contains(err.Error(), "No transactions found") {
				// We've gone through all pages with transactions!
				break
			} else if strings.Contains(err.Error(), "NOTOK") || strings.Contains(err.Error(), "502") {
				// There was some other error (likely rate limiting), so retry with
				// back-off up to 10 times.
				time.Sleep(time.Duration(rateLimitWaitSeconds) * time.Second)
				rateLimitWaitSeconds++
				if rateLimitWaitSeconds > 10 {
					return nil, err
				}
				continue
			} else {
				// There was some non-etherscan error, return it
				return nil, err
			}
		}

		if len(txs) == 0 {
			break
		}

		for _, tx := range txs {
			hash := reflect.ValueOf(tx).FieldByName("Hash")
			if !hash.IsValid() {
				return nil, fmt.Errorf("%t has no Hash field", reflect.TypeOf(tx))
			}
			hashes = append(hashes, hash.String())
		}

		page++
	}

	return hashes, nil
}

func withAnyContractAddress[T any](
	getTxs func(*string, *string, *int, *int, int, int, bool) ([]T, error),
) func(string, *int, *int, int, int, bool) ([]T, error) {
	return func(a string, s, e *int, p, o int, d bool) ([]T, error) {
		return getTxs(nil, &a, s, e, p, o, d)
	}
}

func withAnyReturn[T any](
	getTxs func(string, *int, *int, int, int, bool) ([]T, error),
) func(string, *int, *int, int, int, bool) ([]any, error) {
	return func(a string, s, e *int, p, o int, d bool) ([]any, error) {
		results, err := getTxs(a, s, e, p, o, d)
		if err != nil {
			return nil, err
		}
		anys := make([]any, len(results))
		for i, result := range results {
			anys[i] = result
		}
		return anys, err
	}
}

func withStartEnd(
	startBlock, endBlock *int,
	getTxs func(string, *int, *int, int, int, bool) ([]any, error),
) func(string, int, int, bool) ([]any, error) {
	return func(a string, p, o int, d bool) ([]any, error) {
		return getTxs(a, startBlock, endBlock, p, o, d)
	}
}
