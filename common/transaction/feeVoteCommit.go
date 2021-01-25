// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package transaction

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

// FeeVoteCommitTransaction is Transaction Type that implemented TypeAction
type FeeVoteCommitTransaction struct {
	TransactionObject          *model.Transaction
	Body                       *model.FeeVoteCommitTransactionBody
	FeeScaleService            fee.FeeScaleServiceInterface
	NodeRegistrationQuery      query.NodeRegistrationQueryInterface
	BlockQuery                 query.BlockQueryInterface
	FeeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface
	AccountBalanceHelper       AccountBalanceHelperInterface
	QueryExecutor              query.ExecutorInterface
	EscrowQuery                query.EscrowTransactionQueryInterface
	EscrowFee                  fee.FeeModelInterface
	NormalFee                  fee.FeeModelInterface
}

// ApplyConfirmed to apply confirmed transaction FeeVoteCommitTransaction type
func (tx *FeeVoteCommitTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err        error
		voteCommit *model.FeeVoteCommitmentVote
	)
	// deduct fee from sender
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-tx.TransactionObject.Fee,
		model.EventType_EventFeeVoteCommitTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID, uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	// insert into fee vote
	voteCommit = &model.FeeVoteCommitmentVote{
		VoteHash:     tx.Body.VoteHash,
		VoterAddress: tx.TransactionObject.SenderAccountAddress,
		BlockHeight:  tx.TransactionObject.Height,
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
		err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -tx.TransactionObject.Fee)
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
		err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
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
		feeVotePhase     model.FeeVotePhase
		nodeRegistration model.NodeRegistration
		lastBlock        *model.Block
		enough           bool
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
	qry, qryArgs = tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
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
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
	}
	if !enough {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotEnough")
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
		tx.TransactionObject.SenderAccountAddress,
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
func (tx *FeeVoteCommitTransaction) GetMinimumFee() (int64, error) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
}

// GetSize is size of transaction body
func (tx *FeeVoteCommitTransaction) GetSize() (uint32, error) {
	txBodyBytes, err := tx.GetBodyBytes()
	if err != nil {
		return 0, err
	}
	return uint32(len(txBodyBytes)), nil
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
func (tx *FeeVoteCommitTransaction) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(tx.Body.VoteHash)
	return buffer.Bytes(), nil
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
		if sameTxType && bytes.Equal(tx.TransactionObject.SenderAccountAddress, selectedTx.SenderAccountAddress) {
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

// Escrowable will check the transaction is escrow or not. Currently doesn't have escrow option
func (tx *FeeVoteCommitTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *FeeVoteCommitTransaction) EscrowApplyConfirmed(blockTimestamp int64) (err error) {
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *FeeVoteCommitTransaction) EscrowApplyUnconfirmed() (err error) {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -(tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()))
}

func (tx *FeeVoteCommitTransaction) EscrowUndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
}

func (tx *FeeVoteCommitTransaction) EscrowValidate(dbTx bool) (err error) {
	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}

	var enough bool
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
	}
	if !enough {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotEnough")
	}
	return nil
}

func (tx *FeeVoteCommitTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {
	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			tx.TransactionObject.Fee,
			model.EventType_EventEscrowedTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
		err = tx.ApplyConfirmed(blockTimestamp)
		if err != nil {
			return err
		}
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.Escrow.GetApproverAddress(),
			tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	case model.EscrowApproval_Reject:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Rejected
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.Escrow.GetApproverAddress(),
			tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	default:
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			tx.TransactionObject.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.TransactionObject.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}
	return nil
}
