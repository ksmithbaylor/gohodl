package generic

import (
	"fmt"
	"log"

	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

type AllIndexers map[string]core.Indexer

func NewAllIndexers(networks []core.Network) AllIndexers {
	indexers := make(map[string]core.Indexer)

	fmt.Printf("Getting indexers for %d networks...\n", len(networks))

	for _, network := range networks {
		indexer, err := GetIndexerForNetwork(network)
		if err != nil {
			log.Fatal(err)
		}
		indexers[network.GetName()] = indexer
	}

	fmt.Printf("Got %d indexers\n", len(networks))

	return AllIndexers(indexers)
}

func GetIndexerForNetwork(network core.Network) (core.Indexer, error) {
	switch network.GetKind() {
	case core.EvmNetworkKind:
		return evm.NewIndexer(network.(evm.Network))
	default:
		return nil, fmt.Errorf("No indexer implemented for %s", network.GetKind())
	}
}
