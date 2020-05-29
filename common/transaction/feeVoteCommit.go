package transaction

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// FeeVoteCommitTransaction is Transaction Type that implemented TypeAction
type FeeVoteCommitTransaction struct {
	ID                         int64
	Fee                        int64
	SenderAddress              string
	Height                     uint32
	TimeStamp                  int64
	Body                       *model.FeeVoteCommitTransactionBody
	AccountBalanceQuery        query.AccountBalanceQueryInterface
	BlockQuery                 query.BlockQueryInterface
	AccountLedgerQuery         query.AccountLedgerQueryInterface
	NodeRegistrationQuery      query.NodeRegistrationQueryInterface
	FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
	QueryExecutor              query.ExecutorInterface
}

//ApplyConfirmed to apply confirmed transaction FeeVoteCommitTransaction type
func (tx *FeeVoteCommitTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err        error
		voteCommit *model.FeeVoteCommitmentVote
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
	voteCommit = &model.FeeVoteCommitmentVote{
		VoteHash:     tx.Body.VoteHash,
		VoterAddress: tx.SenderAddress,
		BlockHeight:  tx.Height,
	}
	voteCommitQ, voteCommitArgs := tx.FeeVoteCommitmentVoteQuery.InsertCommitVote(voteCommit)
	queries = append(queries, append([]interface{}{voteCommitQ}, voteCommitArgs...))

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventFeeVoteCommitTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	queries = append(queries, append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...))

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

// ApplyUnconfirmed to apply unconfirmed transaction FeeVoteCommitTransaction type
func (tx *FeeVoteCommitTransaction) ApplyUnconfirmed() error {
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
func (tx *FeeVoteCommitTransaction) UndoApplyUnconfirmed() error {
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
Validate to validating Transaction FeeVoteCommitTransaction type
*/
func (tx *FeeVoteCommitTransaction) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		block          model.Block
		row            *sql.Row
		err            error
	)

	// Checking existing fee vote hash
	if tx.Body.GetVoteHash() == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "fee vote hash required")
	}

	// TODO: check is period to submit commit vote or not

	var (
		qry            string
		args           []interface{}
		commitmentVote model.FeeVoteCommitmentVote
	)
	// check the sender account is awner of node registration
	qry, args = tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	// check account balance sender
	qry, args = tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
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

	// checking prevoius commit vote
	qry, args = tx.FeeVoteCommitmentVoteQuery.GetVoteCommitByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.FeeVoteCommitmentVoteQuery.Scan(&commitmentVote, row)
	if err != nil && err != sql.ErrNoRows {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	// no need to check existing vote commit if height is 0
	if commitmentVote.BlockHeight == 0 {
		return nil
	}
	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.BlockQuery.GetBlockByHeight(commitmentVote.BlockHeight),
		dbTx,
	)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.BlockQuery.Scan(&block, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// check alredy have vote commit or not
	var (
		yearBlock, mountBlock, _ = time.Unix(block.Timestamp, 0).UTC().Date()
		yearTx, mountTx, _       = time.Unix(block.Timestamp, 0).UTC().Date()
	)
	if yearBlock == yearTx && mountBlock == mountTx {
		return blocker.NewBlocker(blocker.ValidationErr, "already have vote commit for this phase")
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *FeeVoteCommitTransaction) GetAmount() int64 {
	return 0
}

// GetMinimumFee return minimum fee of transaction
func (*FeeVoteCommitTransaction) GetMinimumFee() (int64, error) {
	return 0, nil
}

// GetSize is size of transaction body
func (tx *FeeVoteCommitTransaction) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *FeeVoteCommitTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		buffer = bytes.NewBuffer(txBodyBytes)
	)
	voteHashLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.FeeVoteObjectHashLength))
	if err != nil {
		return nil, err
	}
	voteHashLength := util.ConvertBytesToUint32(voteHashLengthBytes)
	voteHash, err := util.ReadTransactionBytes(buffer, int(voteHashLength))
	if err != nil {
		return nil, err
	}
	return &model.FeeVoteCommitTransactionBody{
		VoteHash: voteHash,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *FeeVoteCommitTransaction) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.VoteHash))))
	buffer.Write(tx.Body.VoteHash)
	return buffer.Bytes()
}

// GetTransactionBody return transaction body of FeeVoteCommitTransaction transactions
func (tx *FeeVoteCommitTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_FeeVoteCommitTransactionBody{
		FeeVoteCommitTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *FeeVoteCommitTransaction) SkipMempoolTransaction([]*model.Transaction) (bool, error) {
	return false, nil
}

// Escrowable will check the transaction is escrow or not. Curently doesn't have ecrow option
func (*FeeVoteCommitTransaction) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}
