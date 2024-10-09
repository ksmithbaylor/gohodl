package kevin

import (
	"errors"
	"strings"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

var NOT_HANDLED = errors.New("transaction not handled")

var Implementation = personalHandler(struct{}{})

type personalHandler struct{}

func (h personalHandler) HandleTransaction(
	info *evm.TxInfo,
	client *evm.Client,
	readTransactionBundle handlers.TransactionReader,
	export handlers.CTCWriter,
) (bool, error) {
	readAndThen := func(handle handlers.TransactionHandlerFunc) error {
		tx, receipt, block, err := readTransactionBundle(info.Network, info.Hash)
		if err != nil {
			return err
		}

		bundle := handlers.TransactionBundle{
			Info:    info,
			Tx:      tx,
			Receipt: receipt,
			Block:   block,
		}
		return handle(bundle, client, export)
	}

	var handle handlers.TransactionHandlerFunc

	switch {
	case !info.Success:
		handle = handleFailed
	case info.Method == "":
		handle = handleNoData
	case info.Method == abis.ERC20_TRANSFER || info.Method == abis.ERC20_TRANSFER_FROM:
		handle = handleErc20Transfer
	case info.Method == abis.ERC20_APPROVE:
		handle = handleErc20Approve
	case info.Method == abis.INSTADAPP_CAST:
		handle = handleInstadapp
	case info.Method == abis.AAVE_SUPPLY:
		handle = handleAaveSupply
	case info.Method == abis.AAVE_BORROW:
		handle = handleAaveBorrow
	case info.Method == abis.AAVE_REPAY:
		handle = handleAaveRepay
	case info.Method == abis.AAVE_REPAY_WITH_A_TOKENS:
		handle = handleAaveRepayWithATokens
	case info.Method == abis.AAVE_DEPOSIT:
		handle = handleAaveDeposit
	case info.Method == abis.AAVE_WITHDRAW:
		handle = handleAaveWithdraw
	case info.Method == abis.AAVE_SET_USER_E_MODE:
		handle = handleAaveSetUserEMode
	case info.Method == abis.MOONWELL_ENTER_MARKETS:
		handle = handleMoonwellEnterMarkets
	case
		info.Method == abis.MOONWELL_CLAIM_REWARD,
		info.Method == abis.MOONWELL_CLAIM_REWARD_0,
		info.Method == abis.MOONWELL_STAKING_CLAIM:
		handle = handleMoonwellClaimReward
	case info.Method == abis.MOONWELL_MINT && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellMint
	case info.Method == abis.MOONWELL_BORROW && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellBorrow
	case info.Method == abis.MOONWELL_REPAY_BORROW && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellRepayBorrow
	case info.Method == abis.MOONWELL_REDEEM && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellRedeem
	case info.Method == abis.MOONWELL_STAKING_STAKE:
		handle = handleMoonwellStake
	case info.Method == abis.MOONWELL_STAKING_COOLDOWN:
		handle = handleMoonwellStakingCooldown
	case info.Method == abis.MOONWELL_STAKING_REDEEM:
		handle = handleMoonwellStakingRedeem
	}

	if handle != nil {
		return true, readAndThen(handle)
	}

	return false, nil
}
