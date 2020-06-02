package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

// FeeVoteCommitTransaction is Transaction Type that implemented TypeAction
type FeeVoteCommitTransaction struct {
	ID                         int64
	Fee                        int64
	SenderAddress              string
	Height                     uint32
	Timestamp                  int64
	Body                       *model.FeeVoteCommitTransactionBody
	AccountBalanceQuery        query.AccountBalanceQueryInterface
	NodeRegistrationQuery      query.NodeRegistrationQueryInterface
	FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
	AccountBalanceHelper       AccountBalanceHelperInterface
	AccountLedgerHelper        AccountLedgerHelperInterface
	QueryExecutor              query.ExecutorInterface
}

// ApplyConfirmed to apply confirmed transaction FeeVoteCommitTransaction type
func (tx *FeeVoteCommitTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err        error
		voteCommit *model.FeeVoteCommitmentVote
	)
	// deduct fee from sender
	err = tx.AccountBalanceHelper.AddAccountBalance(tx.SenderAddress, -tx.Fee, tx.Height)
	if err != nil {
		return err
	}
	// sender ledger
	err = tx.AccountLedgerHelper.InsertLedgerEntry(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventMultiSignatureTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	if err != nil {
		return err
	}
	// insert into fee vote
	voteCommit = &model.FeeVoteCommitmentVote{
		VoteHash:     tx.Body.VoteHash,
		VoterAddress: tx.SenderAddress,
		BlockHeight:  tx.Height,
	}
	qry, args := tx.FeeVoteCommitmentVoteQuery.InsertCommitVote(voteCommit)
	err = tx.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		return err
	}
	return nil
}

// ApplyUnconfirmed to apply unconfirmed transaction FeeVoteCommitTransaction type
func (tx *FeeVoteCommitTransaction) ApplyUnconfirmed() error {
	var (
		// update account sender spendable balance
		err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -tx.Fee)
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
		err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
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
		row            *sql.Row
		err            error
	)

	// Checking existing fee vote hash
	if len(tx.Body.GetVoteHash()) != sha3.New256().Size() {
		return blocker.NewBlocker(blocker.ValidationErr, "fee vote hash required")
	}

	// TODO: check is period to submit commit vote or not
	// TODO: check duplicated vote
	var (
		qry              string
		args             []interface{}
		nodeRegistration model.NodeRegistration
	)
	// check the sender account is owner of node registration
	qry, args = tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
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
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *FeeVoteCommitTransaction) GetAmount() int64 {
	return 0
}

// GetMinimumFee return minimum fee of transaction
// TODO: need to calculate the minimum fee
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
	voteHash, err := util.ReadTransactionBytes(buffer, sha3.New256().Size())
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
