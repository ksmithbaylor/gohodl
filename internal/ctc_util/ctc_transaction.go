package ctc_util

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/shopspring/decimal"
)

type CTCTransaction struct {
	Timestamp              time.Time
	Type                   CTCTransactionType
	BaseCurrency           string
	BaseAmount             decimal.Decimal
	QuoteCurrency          string
	QuoteAmount            decimal.Decimal
	FeeCurrency            string
	FeeAmount              decimal.Decimal
	From                   string
	To                     string
	Blockchain             string
	ID                     string
	Description            string
	ReferencePricePerUnit  decimal.Decimal
	ReferencePriceCurrency string
}

var CTC_HEADERS = []string{
	"Timestamp (UTC)",
	"Type",
	"Base Currency",
	"Base Amount",
	"Quote Currency (Optional)",
	"Quote Amount (Optional)",
	"Fee Currency (Optional)",
	"Fee Amount (Optional)",
	"From (Optional)",
	"To (Optional)",
	"Blockchain (Optional)",
	"ID (Optional)",
	"Description (Optional)",
	"Reference Price Per Unit (Optional)",
	"Reference Price Currency (Optional)",
}

func (t *CTCTransaction) ToCSV() []string {
	if t.Type == "" || t.Timestamp.IsZero() || t.BaseCurrency == "" {
		panic("Invalid transaction, missing type or timestamp or base currency")
	}

	return []string{
		t.Timestamp.Format("2006-01-02 15:04:05"),
		string(t.Type),
		t.BaseCurrency,
		t.BaseAmount.String(),
		t.QuoteCurrency,
		emptyIfZero(t.QuoteAmount.String()),
		t.FeeCurrency,
		emptyIfZero(t.FeeAmount.String()),
		t.From,
		t.To,
		t.Blockchain,
		t.ID,
		t.Description,
		emptyIfZero(t.ReferencePricePerUnit.String()),
		t.ReferencePriceCurrency,
	}
}

func (t *CTCTransaction) ToPrintable() map[string]string {
	printable := make(map[string]string)
	values := t.ToCSV()

	for i, header := range CTC_HEADERS {
		printable[header] = values[i]
	}

	return printable
}

func (t *CTCTransaction) AddTransactionFeeIfMine(from, network string, receipt *types.Receipt) {
	if config.Config.IsMyEvmAddressString(from) {
		evmNetwork := config.Config.EvmNetworkByName(network)

		gasPrice, err := decimal.NewFromString(receipt.EffectiveGasPrice.String())
		if err != nil {
			panic("Could not parse gas price")
		}
		gasUsed := decimal.NewFromInt(int64(receipt.GasUsed))
		networkFee := gasPrice.Mul(gasUsed)

		l1Fee, err := decimal.NewFromString(receipt.L1Fee.String())
		if err != nil {
			networkFee = networkFee.Add(l1Fee)
		}

		transactionFee := evmNetwork.NativeAsset().WithAtomicDecimalValue(networkFee)

		t.FeeCurrency = transactionFee.Asset.Symbol
		t.FeeAmount = transactionFee.Value
	}
}

func emptyIfZero(s string) string {
	if s == "0" {
		return ""
	}
	return s
}

type CTCTransactionType string

const (
	CTCBuy                  CTCTransactionType = "buy"
	CTCSell                 CTCTransactionType = "sell"
	CTCFiatDeposit          CTCTransactionType = "fiat-deposit"
	CTCFiatWithdrawal       CTCTransactionType = "fiat-withdrawal"
	CTCFee                  CTCTransactionType = "fee"
	CTCApproval             CTCTransactionType = "approval"
	CTCReceive              CTCTransactionType = "receive"
	CTCSend                 CTCTransactionType = "send"
	CTCChainSplit           CTCTransactionType = "chain-split"
	CTCExpense              CTCTransactionType = "expense"
	CTCStolen               CTCTransactionType = "stolen"
	CTCLost                 CTCTransactionType = "lost"
	CTCBurn                 CTCTransactionType = "burn"
	CTCIncome               CTCTransactionType = "income"
	CTCInterest             CTCTransactionType = "interest"
	CTCMining               CTCTransactionType = "mining"
	CTCAirdrop              CTCTransactionType = "airdrop"
	CTCStaking              CTCTransactionType = "staking"
	CTCStakingDeposit       CTCTransactionType = "staking-deposit"
	CTCStakingWithdrawal    CTCTransactionType = "staking-withdrawal"
	CTCRebate               CTCTransactionType = "rebate"
	CTCRoyalty              CTCTransactionType = "royalty"
	CTCPersonalUse          CTCTransactionType = "personal-use"
	CTCIncomingGift         CTCTransactionType = "incoming-gift"
	CTCOutgoingGift         CTCTransactionType = "outgoing-gift"
	CTCBorrow               CTCTransactionType = "borrow"
	CTCLoanRepayment        CTCTransactionType = "loan-repayment"
	CTCLiquidate            CTCTransactionType = "liquidate"
	CTCBridgeIn             CTCTransactionType = "bridge-in"
	CTCBridgeOut            CTCTransactionType = "bridge-out"
	CTCMint                 CTCTransactionType = "mint"
	CTCCollateralWithdrawal CTCTransactionType = "collateral-withdrawal"
	CTCAddLiquidity         CTCTransactionType = "add-liquidity"
	CTCReceiveLPToken       CTCTransactionType = "receive-lp-token"
	CTCRemoveLiquidity      CTCTransactionType = "remove-liquidity"
	CTCReturnLPToken        CTCTransactionType = "return-lp-token"
	CTCFailedIn             CTCTransactionType = "failed-in"
	CTCFailedOut            CTCTransactionType = "failed-out"
	CTCSpam                 CTCTransactionType = "spam"
	CTCSwapIn               CTCTransactionType = "swap-in"
	CTCSwapOut              CTCTransactionType = "swap-out"
	CTCBridgeTradeIn        CTCTransactionType = "bridge-trade-in"
	CTCBridgeTradeOut       CTCTransactionType = "bridge-trade-out"
	CTCRealizedProfit       CTCTransactionType = "realized-profit"
	CTCRealizedLoss         CTCTransactionType = "realized-loss"
	CTCMarginFee            CTCTransactionType = "margin-fee"
	CTCOpenPosition         CTCTransactionType = "open-position"
	CTCClosePosition        CTCTransactionType = "close-position"
	CTCReceivePQ            CTCTransactionType = "receive-pq"
	CTCSendPQ               CTCTransactionType = "send-pq"
)
