package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// FeeScaleCommitVote is Transaction Type that implemented TypeAction
type FeeScaleCommitVote struct {
	ID                      int64
	Fee                     int64
	SenderAddress           string
	Height                  uint32
	Body                    *model.FeeScaleCommitVoteTransactionsBody
	AccountBalanceQuery     query.AccountBalanceQueryInterface
	BlockQuery              query.BlockQueryInterface
	AccountLedgerQuery      query.AccountLedgerQueryInterface
	FeeScaleVoteCommitQuery query.FeeScaleVoteCommitQueryInterface
	QueryExecutor           query.ExecutorInterface
}

//ApplyConfirmed to apply confirmed transaction FeeScaleCommitVote type
func (tx *FeeScaleCommitVote) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err        error
		voteCommit *model.FeeScaleVoteCommit
		queries    [][]interface{}
	)
	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)
	// build query to insert commit vote
	voteCommit = &model.FeeScaleVoteCommit{
		VoteHash:     tx.Body.VoteHash,
		VoterAddress: tx.SenderAddress,
		BlockHeight:  tx.Height,
	}
	voteCommitQ, voteCommitArgs := tx.FeeScaleVoteCommitQuery.InsertCommitVote(voteCommit)
	queries = append(queries, append([]interface{}{voteCommitQ}, voteCommitArgs...))

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventFeeScaleCommitVoteTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	queries = append(queries, append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...))

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

// ApplyUnconfirmed to apply unconfirmed transaction FeeScaleCommitVote type
func (tx *FeeScaleCommitVote) ApplyUnconfirmed() error {
	var (
		// update account sender spendable balance
		accountBalanceSenderQ, accountBalanceSenderQArgs = tx.AccountBalanceQuery.AddAccountSpendableBalance(
			-(tx.Fee),
			map[string]interface{}{
				"account_address": tx.SenderAddress,
			},
		)
		err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *FeeScaleCommitVote) UndoApplyUnconfirmed() error {
	var (
		// update account sender spendable balance
		accountBalanceSenderQ, accountBalanceSenderQArgs = tx.AccountBalanceQuery.AddAccountSpendableBalance(
			tx.Fee,
			map[string]interface{}{
				"account_address": tx.SenderAddress,
			},
		)
		err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil
}

/*
Validate to validating Transaction FeeScaleCommitVote type
*/
func (tx *FeeScaleCommitVote) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		lastBlock      model.Block
		row            *sql.Row
		err            error
	)

	// Checking existing fee vote hash
	if tx.Body.GetVoteHash() == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "fee vote hash required")
	}

	// check is period to submit commit vote or not
	row, err = tx.QueryExecutor.ExecuteSelectRow(tx.BlockQuery.GetLastBlock(), dbTx)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var day = util.GetDayOfMonthUTC(lastBlock.GetTimestamp())
	if day > constant.FeeScaleDayPhaseBounds {
		return blocker.NewBlocker(blocker.ValidationErr, "curently it's not commitment phase")
	}

	// check account balance sender
	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "balance not enough")
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *FeeScaleCommitVote) GetAmount() int64 {
	return 0
}

// GetMinimumFee return minimum fee of transaction
func (*FeeScaleCommitVote) GetMinimumFee() (int64, error) {
	return 0, nil
}

// GetSize is size of transaction body
func (tx *FeeScaleCommitVote) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *FeeScaleCommitVote) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		buffer = bytes.NewBuffer(txBodyBytes)
	)
	voteHashLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.FeeScaleVoteHashLength))
	if err != nil {
		return nil, err
	}
	voteHashLength := util.ConvertBytesToUint32(voteHashLengthBytes)
	voteHash, err := util.ReadTransactionBytes(buffer, int(voteHashLength))
	if err != nil {
		return nil, err
	}
	return &model.FeeScaleCommitVoteTransactionsBody{
		VoteHash: voteHash,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *FeeScaleCommitVote) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.VoteHash))))
	buffer.Write(tx.Body.VoteHash)
	return buffer.Bytes()
}

// GetTransactionBody return transaction body of FeeScaleCommitVote transactions
func (tx *FeeScaleCommitVote) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_FeeScaleCommitVoteTransactionBody{
		FeeScaleCommitVoteTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *FeeScaleCommitVote) SkipMempoolTransaction([]*model.Transaction) (bool, error) {
	return false, nil
}

// Escrowable will check the transaction is escrow or not. Curently doesn't have ecrow option
func (*FeeScaleCommitVote) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}
