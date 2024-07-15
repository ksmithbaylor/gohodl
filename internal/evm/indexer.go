package evm

import (
	"fmt"

	"github.com/ksmithbaylor/gohodl/internal/core"
)

func NewIndexer(network Network) (core.Indexer, error) {
	if network.Etherscan.URL != "" {
		return NewEtherscanClient(network)
	}

	return nil, fmt.Errorf(
		"No etherscan config for %s, and no alternative implementation",
		network.Name,
	)
}
