package abis

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func DecodeAdhoc(rawSelector string, data []uint8) (string, []any, error) {
	selector, err := abi.ParseSelector(rawSelector)
	if err != nil {
		return "", nil, fmt.Errorf("Could not parse selector: %w", err)
	}

	selector.Type = "event"

	selectorJSON, err := json.Marshal(selector)
	if err != nil {
		return "", nil, fmt.Errorf("Coult not marshal json: %w", err)
	}

	abi, err := abi.JSON(strings.NewReader("[" + string(selectorJSON) + "]"))
	if err != nil {
		return "", nil, fmt.Errorf("Could not read abi: %w", err)
	}

	args, err := abi.Unpack(selector.Name, data)
	if err != nil {
		return "", nil, fmt.Errorf("Could not unpack abi: %w", err)
	}

	return selector.Name, args, nil
}
