package indexing

import (
	"fmt"

	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

func GetIndexerForNetwork(network core.Network) (core.Indexer, error) {
	switch network.GetKind() {
	case core.EvmNetworkKind:
		return evm.NewIndexer(network.(evm.Network))
	default:
		return nil, fmt.Errorf("No indexer implemented for %s", network.GetKind())
	}
}
