package generic

import (
	"fmt"
	"log"
	"sync"

	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

type AllNodeClients map[string]core.NodeClient

func NewAllNodeClients(networks []core.Network) AllNodeClients {
	clients := make(map[string]core.NodeClient)
	var clientsMu sync.Mutex

	errs := make([]error, 0)
	var errsMu sync.Mutex

	var wg sync.WaitGroup

	fmt.Printf("Getting node clients for %d networks...\n", len(networks))

	for _, network := range networks {
		wg.Add(1)
		go func(network core.Network) {
			defer wg.Done()
			client, err := GetNodeClientForNetwork(network)
			if err != nil {
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("Failed to make node client for %s: %w", network.GetName(), err))
				errsMu.Unlock()
				return
			}

			clientsMu.Lock()
			clients[network.GetName()] = client
			clientsMu.Unlock()
		}(network)
	}

	wg.Wait()

	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		log.Fatal("Error(s) connecting to nodes, see logs above")
	}

	fmt.Printf("Got %d node clients\n", len(networks))

	return AllNodeClients(clients)
}

func GetNodeClientForNetwork(network core.Network) (core.NodeClient, error) {
	switch network.GetKind() {
	case core.EvmNetworkKind:
		return evm.NewClient(network.(evm.Network))
	default:
		return nil, fmt.Errorf("No node client implemented for %s", network.GetKind())
	}
}

func (anc AllNodeClients) LatestBlocks() map[string]uint64 {
	blocks := make(map[string]uint64)
	var blocksMu sync.Mutex

	errs := make([]error, 0)
	var errsMu sync.Mutex

	var wg sync.WaitGroup

	fmt.Printf("Getting latest blocks for %d networks...\n", len(anc))

	for networkName, client := range anc {
		wg.Add(1)
		go func(networkName string, client core.NodeClient) {
			defer wg.Done()

			block, err := client.LatestBlock()
			if err != nil {
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("Failed to get latest block for %s: %w", networkName, err))
				errsMu.Unlock()
				return
			}

			blocksMu.Lock()
			fmt.Printf("Network %s is on block %d\n", networkName, block)
			blocks[networkName] = block
			blocksMu.Unlock()
		}(networkName, client)
	}

	wg.Wait()

	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		log.Fatal("Error(s) getting latest blocks, see logs above")
	}

	return blocks
}

func (anc AllNodeClients) ForEach(action func(networkName string, client core.NodeClient)) {
	var wg sync.WaitGroup

	for networkName, client := range anc {
		wg.Add(1)
		go func(networkName string, client core.NodeClient) {
			defer wg.Done()
			action(networkName, client)
		}(networkName, client)
	}

	wg.Wait()
}
