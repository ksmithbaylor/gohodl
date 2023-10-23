package eth

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func withRetry[R comparable](
	connections map[string]*ethclient.Client,
	action func(*ethclient.Client) (R, error),
) (R, error) {
	var ret R
	var err error

	for i := 0; i < CONSENSUS_RETRIES; i++ {
		ret, err = ensureAgreement(connections, action)
		if err == nil {
			return ret, nil
		} else {
			util.Debug(err)
		}
	}

	return ret, err
}

func ensureAgreement[R comparable](
	connections map[string]*ethclient.Client,
	getUsing func(*ethclient.Client) (R, error),
) (R, error) {
	tried := make(map[string]bool, len(connections))
	var clients []*ethclient.Client
	var wg sync.WaitGroup
	var mu sync.Mutex

	for rpc, client := range connections {
		clients = append(clients, client)
		tried[rpc] = true
		if len(clients) == QUORUM {
			break
		}
	}

	answers := make(map[R]int, 0)

	for _, client := range clients {
		wg.Add(1)
		go func(c *ethclient.Client) {
			defer wg.Done()

			result, err := getUsing(c)
			if err == nil {
				mu.Lock()
				answers[result]++
				util.Debugf("got answer: %#+v\n", result)
				mu.Unlock()
			} else {
				util.Debug(err)
			}
		}(client)
	}
	wg.Wait()

	for rpc, client := range connections {
		for answer, votes := range answers {
			if votes >= QUORUM {
				return answer, nil
			}
		}

		if tried[rpc] {
			continue
		}
		tried[rpc] = true
		result, err := getUsing(client)
		if err == nil {
			answers[result]++
			util.Debugf("got answer: %#+v\n", result)
		} else {
			util.Debug(err)
		}
	}

	var nothing R
	return nothing, fmt.Errorf("No two nodes were successful and agreed")
}
