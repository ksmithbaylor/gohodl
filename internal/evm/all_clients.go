package evm

import (
	"fmt"
	"sync"
)

func AllClients(allNetworks []Network) (map[NetworkName]*Client, []error) {
	clients := make(map[NetworkName]*Client)
	var clientsMu sync.Mutex

	errs := make([]error, 0)
	var errsMu sync.Mutex

	var wg sync.WaitGroup

	for _, network := range allNetworks {
		network := network
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := NewClient(network)

			if err != nil {
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("failed to make client for %s: %w", network.Name, err))
				errsMu.Unlock()
				return
			}

			clientsMu.Lock()
			clients[network.Name] = client
			clientsMu.Unlock()
		}()
	}

	wg.Wait()

	return clients, errs
}
