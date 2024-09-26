package kevin

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

const AVAX_OLD_BRIDGE_L1 = "0xE78388b4CE79068e89Bf8aA7f218eF6b9AB0e9d0"

func handleErc20Transfer(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.Erc20Abi)
	if err != nil {
		return err
	}

	if len(events) != 1 {
		panic("Found more than one event in ERC20 transfer call")
	}

	transfer := events[0]
	if transfer.Name != "Transfer" {
		panic("Non-transfer event found in ERC20 transfer call")
	}

	from := transfer.Data["from"].(common.Address).Hex()
	to := transfer.Data["to"].(common.Address).Hex()
	value := transfer.Data["value"].(*big.Int).String()
	tokenAsset, err := client.TokenAsset(transfer.Contract)
	if err != nil {
		return err
	}

	amount, err := tokenAsset.WithAtomicStringValue(value)
	if err != nil {
		return err
	}

	ctcType := ctc_util.CTCSend
	if !config.Config.IsMyEvmAddressString(from) {
		if config.Config.IsMyEvmAddressString(to) {
			ctcType = ctc_util.CTCReceive
		} else {
			panic("Found irrelevant transaction, not from/to any of my addresses")
		}
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctcType,
		BaseCurrency: tokenAsset.Symbol,
		BaseAmount:   amount.Value,
		From:         from,
		To:           to,
		Description: fmt.Sprintf("transfer %s from %s to %s on %s",
			amount.String(),
			from,
			to,
			bundle.Info.Network,
		),
	}

	if bundle.Info.To == AVAX_OLD_BRIDGE_L1 {
		ctcTx.Type = ctc_util.CTCBridgeOut
		ctcTx.To = ctcTx.From
		ctcTx.Description = fmt.Sprintf("bridge %s to %s on avalanche",
			amount.String(),
			to,
		)

		otherSideTx := ctc_util.CTCTransaction{
			Timestamp:    ctcTx.Timestamp,
			Blockchain:   "avalanche",
			ID:           bundle.Info.Hash + "-from-ethereum",
			Type:         ctc_util.CTCBridgeIn,
			BaseCurrency: tokenAsset.Symbol,
			BaseAmount:   amount.Value,
			From:         ctcTx.From,
			To:           ctcTx.To,
			Description: fmt.Sprintf("bridge %s to %s on avalanche (not a real tx)",
				amount.String(),
				to,
			),
		}
		err = export(otherSideTx.ToCSV())
		if err != nil {
			return fmt.Errorf("Could not export synthetic avax bridge tx: %w", err)
		}
	}

	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleErc20Approve(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.Erc20Abi)
	if err != nil {
		return err
	}

	if len(events) != 1 {
		panic("Found more than one event in ERC20 transfer call")
	}

	approval := events[0]
	if approval.Name != "Approval" {
		panic("Non-approval event found in ERC20 approve call")
	}

	owner := approval.Data["owner"].(common.Address).Hex()
	spender := approval.Data["spender"].(common.Address).Hex()
	value := approval.Data["value"].(*big.Int).String()

	tokenAsset, err := client.TokenAsset(approval.Contract)
	if err != nil {
		return err
	}

	amount, err := tokenAsset.WithAtomicStringValue(value)
	if err != nil {
		return err
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCApproval,
		BaseCurrency: tokenAsset.Symbol,
		// BaseAmount:   amount.Value, // Don't include, messes up CTC
		From:        owner,
		To:          spender,
		Description: fmt.Sprintf("approve %s for spending by %s from %s", amount.String(), spender, owner),
	}

	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
