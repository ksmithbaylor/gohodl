package evm

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

const (
	QUORUM            int = 2
	CONSENSUS_RETRIES int = 5
)

func ensureAgreementWithRetry[R any, C comparable](
	connections map[string]*ethclient.Client,
	action func(*ethclient.Client) (R, C, error),
) (R, error) {
	var ret R
	var err error

	for i := 0; i < CONSENSUS_RETRIES; i++ {
		ret, err = ensureAgreement(connections, action)

		if err == nil {
			return ret, nil
		} else {
			util.Debug(err)
			time.Sleep(time.Millisecond * 500)
		}
	}

	return ret, err
}

func ensureAgreement[R any, C comparable](
	connections map[string]*ethclient.Client,
	getUsing func(*ethclient.Client) (R, C, error),
) (R, error) {
	tried := make(map[string]bool, len(connections))
	var rpcs []string
	var clients []*ethclient.Client
	var wg sync.WaitGroup
	var mu sync.Mutex

	for rpc, client := range connections {
		clients = append(clients, client)
		rpcs = append(rpcs, rpc)
		tried[rpc] = true
		if len(clients) == QUORUM {
			break
		}
	}

	votes := make(map[C]int, 0)
	answers := make(map[C]R, 0)

	for i, client := range clients {
		wg.Add(1)
		go func(rpc string, c *ethclient.Client) {
			defer wg.Done()

			result, compKey, err := getUsing(c)
			if err == nil {
				mu.Lock()
				answers[compKey] = result
				votes[compKey]++
				util.Debugf("Success from %s: %#+v\n", rpc, result)
				mu.Unlock()
			} else {
				util.Debugf("Problem with %s: %s", rpc, err.Error())
			}
		}(rpcs[i], client)
	}
	wg.Wait()

	for rpc, client := range connections {
		for compKey, voteCount := range votes {
			if voteCount >= QUORUM {
				return answers[compKey], nil
			}
		}

		if tried[rpc] {
			continue
		}
		tried[rpc] = true
		result, compKey, err := getUsing(client)
		if err == nil {
			mu.Lock()
			answers[compKey] = result
			votes[compKey]++
			util.Debugf("Success from %s: %#+v\n", rpc, result)
			mu.Unlock()
		} else {
			util.Debugf("Problem with %s: %s", rpc, err.Error())
		}
	}

	var nothing R
	return nothing, fmt.Errorf("No quorum was successful and agreed")
}
