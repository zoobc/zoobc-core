package transaction

import (
	"bytes"
	"database/sql"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// FeeVoteCommitTransaction is Transaction Type that implemented TypeAction
type FeeVoteCommitTransaction struct {
	ID                         int64
	Fee                        int64
	SenderAddress              []byte
	Height                     uint32
	Body                       *model.FeeVoteCommitTransactionBody
	FeeScaleService            fee.FeeScaleServiceInterface
	NodeRegistrationQuery      query.NodeRegistrationQueryInterface
	BlockQuery                 query.BlockQueryInterface
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
	qry, qryArgs := tx.FeeVoteCommitmentVoteQuery.InsertCommitVote(voteCommit)
	err = tx.QueryExecutor.ExecuteTransaction(qry, qryArgs...)
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
		return err
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
		row              *sql.Row
		err              error
		qry              string
		qryArgs          []interface{}
		accountBalance   model.AccountBalance
		feeVotePhase     model.FeeVotePhase
		nodeRegistration model.NodeRegistration
		lastBlock        *model.Block
	)

	// Checking length hash of fee vote
	if len(tx.Body.GetVoteHash()) != sha3.New256().Size() {
		return blocker.NewBlocker(blocker.ValidationErr, "FeeVoteHashRequired")
	}

	// check is period to submit commit vote or not
	lastBlock, err = util.GetLastBlock(tx.QueryExecutor, tx.BlockQuery)
	if err != nil {
		return err
	}
	feeVotePhase, _, err = tx.FeeScaleService.GetCurrentPhase(lastBlock.Timestamp, true)
	if err != nil {
		return err
	}
	if feeVotePhase != model.FeeVotePhase_FeeVotePhaseCommmit {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidFeeCommitVotePeriod")
	}
	// check duplicate vote
	err = tx.checkDuplicateVoteCommit(dbTx)
	if err != nil {
		return err
	}
	// check the sender account is owner of node registration
	qry, qryArgs = tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, qryArgs...)
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

	// check existing & balance account sender
	err = tx.AccountBalanceHelper.GetBalanceByAccountID(&accountBalance, tx.SenderAddress, dbTx)
	if err != nil {
		return err
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "BalanceNotEnough")
	}
	return nil
}

func (tx *FeeVoteCommitTransaction) checkDuplicateVoteCommit(dbTx bool) error {
	var (
		err          error
		row          *sql.Row
		qry          string
		qryArgs      []interface{}
		voteCommit   model.FeeVoteCommitmentVote
		lastFeeScale model.FeeScale
	)
	// get last fee scale height
	err = tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// get previous vote based on sender account address and lastest fee scale height
	qry, qryArgs = tx.FeeVoteCommitmentVoteQuery.GetVoteCommitByAccountAddressAndHeight(
		tx.SenderAddress,
		lastFeeScale.BlockHeight,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, qryArgs...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.FeeVoteCommitmentVoteQuery.Scan(&voteCommit, row)
	if err != nil {
		// it means don't have vote commit for current phase
		if err == sql.ErrNoRows {
			return nil
		}
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return blocker.NewBlocker(blocker.ValidationErr, "DuplicatedCommitVote")
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

/*
SkipMempoolTransaction filter out current fee commit vote tx when
	- Current time is already not commit vote phase based on new block timestamp
	- There are other tx fee commit vote with same sender in mempool
	- Fee commit vote tx for current phase already exist in previous block
*/
func (tx *FeeVoteCommitTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	// check tx is still valid for commit vote phase based on new block timestamp
	var feeVotePhase, _, err = tx.FeeScaleService.GetCurrentPhase(newBlockTimestamp, true)
	if err != nil {
		return true, err
	}
	if feeVotePhase != model.FeeVotePhase_FeeVotePhaseCommmit {
		return true, nil
	}
	// check duplicate vote on mempool
	for _, selectedTx := range selectedTransactions {
		// if we find another fee vote commit tx in currently selected transactions, filter current one out of selection
		sameTxType := model.TransactionType_FeeVoteCommitmentVoteTransaction == model.TransactionType(selectedTx.GetTransactionType())
		if sameTxType && bytes.Equal(tx.SenderAddress, selectedTx.SenderAccountAddress) {
			return true, nil
		}
	}
	// check duplicate on previous vote
	err = tx.checkDuplicateVoteCommit(false)
	if err != nil {
		if strings.Contains(err.Error(), string(blocker.ValidationErr)) {
			return true, nil
		}
		return true, err
	}
	return false, nil
}

// Escrowable will check the transaction is escrow or not. Curently doesn't have ecrow option
func (*FeeVoteCommitTransaction) Escrowable() (EscrowTypeAction, bool) {
	return nil, false
}
