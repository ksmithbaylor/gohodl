package evm_util

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

type TokenTransfers map[common.Address]*core.Amount
type NetTransfers map[core.Asset]TokenTransfers

func (nt NetTransfers) String() string {
	s := ""

	if len(nt) == 0 {
		return s
	}

	for asset, transfers := range nt {
		s += fmt.Sprintf("  %s:\n", asset)
		for addr, amount := range transfers {
			s += fmt.Sprintf("    %s: %s\n", addr.Hex(), amount)
		}
	}

	return s[:len(s)-1]
}

func (nt NetTransfers) Print() {
	fmt.Printf("net transfers:\n%s\n", nt.String())
}

type transfer struct {
	from   common.Address
	to     common.Address
	amount core.Amount
}

func NetTokenTransfers(client *evm.Client, info *evm.TxInfo, logs []*types.Log) (NetTransfers, error) {
	transfers := make([]transfer, 0)
	netTransfers := make(NetTransfers)

	nativeAsset, err := client.NativeAsset()
	if err != nil {
		return nil, err
	}

	if info.Value != "0" {
		amount, err := nativeAsset.WithAtomicStringValue(info.Value)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer{
			common.HexToAddress(info.From),
			common.HexToAddress(info.To),
			amount,
		})
	}

	erc20Logs, err := evm.ParseKnownEvents(info.Network, logs, abis.Erc20Abi)
	if err != nil {
		return nil, err
	}

	for _, log := range erc20Logs {
		if log.Name != "Transfer" {
			continue
		}
		from, ok := log.Data["from"]
		if !ok {
			continue
		}
		to, ok := log.Data["to"]
		if !ok {
			continue
		}
		value, ok := log.Data["value"]
		if !ok {
			continue
		}
		fromAddr, ok := from.(common.Address)
		if !ok {
			continue
		}
		toAddr, ok := to.(common.Address)
		if !ok {
			continue
		}
		valueBig, ok := value.(*big.Int)
		if !ok {
			continue
		}

		asset, err := client.TokenAsset(log.Contract)
		if err != nil {
			return nil, err
		}

		amount, err := asset.WithAtomicStringValue(valueBig.String())
		if err != nil {
			return nil, err
		}

		transfers = append(transfers, transfer{fromAddr, toAddr, amount})
	}

	wrappedNativeLogs, err := evm.ParseKnownEvents(info.Network, logs, abis.WrappedNativeAbi)
	if err != nil {
		return nil, err
	}

	for _, log := range wrappedNativeLogs {
		if log.Name != "Deposit" && log.Name != "Withdrawal" {
			continue
		}

		asset, err := client.TokenAsset(log.Contract)
		if err != nil {
			return nil, err
		}

		var addrRaw any
		var ok bool
		if log.Name == "Deposit" {
			addrRaw, ok = log.Data["dst"]
		} else {
			addrRaw, ok = log.Data["src"]
		}
		if !ok {
			continue
		}

		var addr common.Address
		addr, ok = addrRaw.(common.Address)
		if !ok {
			continue
		}

		wad, ok := log.Data["wad"]
		if !ok {
			continue
		}

		value, ok := wad.(*big.Int)
		if !ok {
			continue
		}

		amount, err := asset.WithAtomicStringValue(value.String())
		if err != nil {
			return nil, err
		}

		if log.Name == "Deposit" {
			transfers = append(transfers, transfer{log.Contract, addr, amount})
		} else {
			transfers = append(transfers, transfer{addr, log.Contract, amount})
		}
	}

	internalTxs, _, err := client.GetInternalTransactions(info.Hash)
	if err != nil {
		return nil, err
	}

	for _, tx := range internalTxs {
		amount, err := nativeAsset.WithAtomicStringValue(tx.Value.Int().String())
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer{
			common.HexToAddress(tx.From),
			common.HexToAddress(tx.To),
			amount,
		})
	}

	for _, transfer := range transfers {
		asset := transfer.amount.Asset

		if netTransfers[asset] == nil {
			netTransfers[asset] = make(TokenTransfers)
		}

		if netTransfers[asset][transfer.from] == nil {
			zero := asset.WithAtomicValue(0)
			netTransfers[asset][transfer.from] = &zero
		}

		if netTransfers[asset][transfer.to] == nil {
			zero := asset.WithAtomicValue(0)
			netTransfers[asset][transfer.to] = &zero
		}

		newFromAmount, err := netTransfers[asset][transfer.from].Sub(transfer.amount)
		if err != nil {
			return nil, err
		}
		netTransfers[asset][transfer.from] = &newFromAmount

		newToAmount, err := netTransfers[asset][transfer.to].Add(transfer.amount)
		if err != nil {
			return nil, err
		}
		netTransfers[asset][transfer.to] = &newToAmount
	}

	return netTransfers, nil
}

func NetTokenTransfersOnlyMine(client *evm.Client, info *evm.TxInfo, logs []*types.Log) (NetTransfers, error) {
	netTransfers, err := NetTokenTransfers(client, info, logs)
	if err != nil {
		return nil, err
	}

	for asset, transfers := range netTransfers {
		for addr, amount := range transfers {
			if !config.Config.IsMyEvmAddress(addr) || amount.Value.IsZero() {
				delete(transfers, addr)
			}
		}
		if len(transfers) == 0 {
			delete(netTransfers, asset)
		}
	}

	return netTransfers, nil
}
