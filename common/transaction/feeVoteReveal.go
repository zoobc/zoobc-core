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
	"crypto/sha256"
	"database/sql"
	"strings"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	FeeVoteRevealTransaction struct {
		TransactionObject      *model.Transaction
		Body                   *model.FeeVoteRevealTransactionBody
		FeeScaleService        fee.FeeScaleServiceInterface
		SignatureInterface     crypto.SignatureInterface
		BlockQuery             query.BlockQueryInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		FeeVoteCommitVoteQuery query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface
		AccountBalanceHelper   AccountBalanceHelperInterface
		QueryExecutor          query.ExecutorInterface
		EscrowQuery            query.EscrowTransactionQueryInterface
	}
)

// Validate for validating the transaction concerned
func (tx *FeeVoteRevealTransaction) Validate(dbTx bool) error {
	var (
		feeVotePhase model.FeeVotePhase
		recentBlock  model.Block
		commitVote   model.FeeVoteCommitmentVote
		nodeReg      model.NodeRegistration
		lastFeeScale model.FeeScale
		args         []interface{}
		row          *sql.Row
		qry          string
		err          error
		enough       bool
	)

	// check the transaction submitted on reveal-phase
	feeVotePhase, _, err = tx.FeeScaleService.GetCurrentPhase(tx.TransactionObject.Timestamp, true)
	if err != nil {
		return err
	}
	if feeVotePhase != model.FeeVotePhase_FeeVotePhaseReveal {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidPhasePeriod")
	}

	// get last fee scale height
	err = tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// must match the previously submitted in CommitmentVote
	qry, args = tx.FeeVoteCommitVoteQuery.GetVoteCommitByAccountAddressAndHeight(
		tx.TransactionObject.SenderAccountAddress,
		lastFeeScale.BlockHeight,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.FeeVoteCommitVoteQuery.Scan(&commitVote, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "CommitVoteNotFound")
		}
		return err
	}

	digest := sha3.New256()
	_, err = digest.Write(tx.GetFeeVoteInfoBytes())
	if err != nil {
		return err
	}

	if res := bytes.Compare(commitVote.GetVoteHash(), digest.Sum([]byte{})); res != 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "NotMatchVoteHashed")
	}

	// VoteObject.Signature must be a valid signature from node-owner on bytes(VoteInfo)
	err = tx.SignatureInterface.VerifySignature(
		tx.GetFeeVoteInfoBytes(),
		tx.Body.GetVoterSignature(),
		tx.TransactionObject.SenderAccountAddress,
	)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}
	row, err = tx.QueryExecutor.ExecuteSelectRow(
		tx.BlockQuery.GetBlockByHeight(tx.Body.GetFeeVoteInfo().GetRecentBlockHeight()),
		dbTx,
	)
	if err != nil {
		return err
	}
	err = tx.BlockQuery.Scan(&recentBlock, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "BlockNotFound")
	}
	if res := bytes.Compare(tx.Body.GetFeeVoteInfo().GetRecentBlockHash(), recentBlock.GetBlockHash()); res != 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidRecentBlock")
	}

	// sender must be as node owner
	qry, args = tx.NodeRegistrationQuery.GetNodeRegistrationByAccountAddress(tx.TransactionObject.SenderAccountAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
	err = tx.NodeRegistrationQuery.Scan(&nodeReg, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "SenderAccountNotNodeOwner")
	}

	// check duplicated reveal to database, once per node owner per period
	err = tx.checkDuplicateVoteReveal(dbTx)
	if err != nil {
		return err
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

func (tx *FeeVoteRevealTransaction) checkDuplicateVoteReveal(dbTx bool) error {
	var (
		revealVote model.FeeVoteRevealVote
		qry, args  = tx.FeeVoteRevealVoteQuery.GetFeeVoteRevealByAccountAddressAndRecentBlockHeight(
			tx.TransactionObject.SenderAccountAddress,
			tx.Body.GetFeeVoteInfo().GetRecentBlockHeight(),
		)
		row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	)
	if err != nil {
		return err
	}
	err = tx.FeeVoteRevealVoteQuery.Scan(&revealVote, row)
	if err != nil {
		// it means don't have previous vote
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return blocker.NewBlocker(blocker.ValidationErr, "DuplicatedFeeVoteReveal")
}

// ApplyUnconfirmed to apply unconfirmed transaction
func (tx *FeeVoteRevealTransaction) ApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -tx.TransactionObject.Fee)
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *FeeVoteRevealTransaction) UndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
}

// ApplyConfirmed applying transaction, will store ledger, account balance update, and also the transaction it self
func (tx *FeeVoteRevealTransaction) ApplyConfirmed(blockTimestamp int64) (err error) {

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-tx.TransactionObject.Fee,
		model.EventType_EventFeeVoteRevealTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	qry, args := tx.FeeVoteRevealVoteQuery.InsertRevealVote(&model.FeeVoteRevealVote{
		VoteInfo:       tx.Body.GetFeeVoteInfo(),
		VoterSignature: tx.Body.GetVoterSignature(),
		VoterAddress:   tx.TransactionObject.SenderAccountAddress,
		BlockHeight:    tx.TransactionObject.Height,
	})
	err = tx.QueryExecutor.ExecuteTransaction(qry, args...)
	if err != nil {
		return err
	}
	return nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*FeeVoteRevealTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		buff    = bytes.NewBuffer(txBodyBytes)
		chunked []byte
		err     error
	)

	recentBlockHash, err := util.ReadTransactionBytes(buff, sha256.Size)
	if err != nil {
		return nil, err
	}

	chunked, err = util.ReadTransactionBytes(buff, int(constant.RecentBlockHeight))
	if err != nil {
		return nil, err
	}
	recentBlockHeight := util.ConvertBytesToUint32(chunked)

	chunked, err = util.ReadTransactionBytes(buff, int(constant.FeeVote))
	if err != nil {
		return nil, err
	}
	feeVote := util.ConvertBytesToUint64(chunked)

	chunked, err = util.ReadTransactionBytes(buff, int(constant.VoterSignatureLength))
	if err != nil {
		return nil, err
	}
	voterSignature, err := util.ReadTransactionBytes(buff, int(util.ConvertBytesToUint32(chunked)))
	if err != nil {
		return nil, err
	}
	return &model.FeeVoteRevealTransactionBody{
		FeeVoteInfo: &model.FeeVoteInfo{
			RecentBlockHash:   recentBlockHash,
			RecentBlockHeight: recentBlockHeight,
			FeeVote:           int64(feeVote),
		},
		VoterSignature: voterSignature,
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *FeeVoteRevealTransaction) GetBodyBytes() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	buff.Write(tx.Body.FeeVoteInfo.RecentBlockHash)
	buff.Write(util.ConvertUint32ToBytes(tx.Body.FeeVoteInfo.RecentBlockHeight))
	buff.Write(util.ConvertUint64ToBytes(uint64(tx.Body.FeeVoteInfo.FeeVote)))
	buff.Write(util.ConvertUint32ToBytes(uint32(len(tx.Body.VoterSignature))))
	buff.Write(tx.Body.VoterSignature)
	return buff.Bytes(), nil
}

// GetTransactionBody append isTransaction_TransactionBody oneOf
func (tx *FeeVoteRevealTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_FeeVoteRevealTransactionBody{
		FeeVoteRevealTransactionBody: tx.Body,
	}
}

// GetFeeVoteInfoBytes will build bytes from model.FeeVoteInfo
func (tx *FeeVoteRevealTransaction) GetFeeVoteInfoBytes() []byte {
	buff := bytes.NewBuffer([]byte{})
	buff.Write(tx.Body.FeeVoteInfo.RecentBlockHash)
	buff.Write(util.ConvertUint32ToBytes(tx.Body.FeeVoteInfo.RecentBlockHeight))
	buff.Write(util.ConvertUint64ToBytes(uint64(tx.Body.FeeVoteInfo.FeeVote)))
	return buff.Bytes()
}

// GetAmount return Amount from TransactionBody
func (tx *FeeVoteRevealTransaction) GetAmount() int64 {
	return 0
}

// GetMinimumFee calculate fee
func (tx *FeeVoteRevealTransaction) GetMinimumFee() (int64, error) {
	var lastFeeScale model.FeeScale
	err := tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return 0, err
	}
	return fee.CalculateTxMinimumFee(tx.TransactionObject, lastFeeScale.FeeScale)
}

/*
SkipMempoolTransaction filter out current fee reveal vote tx when
	- Current time is already not reveal vote phase based on new block timestamp
	- There are other tx fee reveal vote with same sender in mempool
	- Fee reveal vote tx for current phase already exist in previous block
*/
func (tx *FeeVoteRevealTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	// check tx is still valid for reveal vote phase based on new block timestamp
	var feeVotePhase, _, err = tx.FeeScaleService.GetCurrentPhase(newBlockTimestamp, true)
	if err != nil {
		return true, err
	}
	if feeVotePhase != model.FeeVotePhase_FeeVotePhaseReveal {
		return true, nil
	}
	// check duplicate vote on mempool
	for _, selectedTx := range selectedTransactions {
		// if we find another fee reveal tx in currently selected transactions, filter current one out of selection
		sameTxType := model.TransactionType_FeeVoteRevealVoteTransaction == model.TransactionType(selectedTx.GetTransactionType())
		if sameTxType && bytes.Equal(tx.TransactionObject.SenderAccountAddress, selectedTx.SenderAccountAddress) {
			return true, nil
		}
	}
	// check previous vote
	err = tx.checkDuplicateVoteReveal(false)
	if err != nil {
		if strings.Contains(err.Error(), string(blocker.ValidationErr)) {
			return true, nil
		}
		return true, err
	}
	return false, nil
}

// GetSize send money Amount should be 8
func (tx *FeeVoteRevealTransaction) GetSize() (uint32, error) {
	// only amount
	txBodyBytes, err := tx.GetBodyBytes()
	if err != nil {
		return 0, err
	}
	return uint32(len(txBodyBytes)), nil
}

// Escrowable will check the transaction is escrow or not. Currently doesn't have escrow option
func (tx *FeeVoteRevealTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil &&
		!bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *FeeVoteRevealTransaction) EscrowApplyConfirmed(blockTimestamp int64) (err error) {
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

func (tx *FeeVoteRevealTransaction) EscrowApplyUnconfirmed() (err error) {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -tx.TransactionObject.Fee)
}

func (tx *FeeVoteRevealTransaction) EscrowUndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
}

func (tx *FeeVoteRevealTransaction) EscrowValidate(dbTx bool) (err error) {
	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}

	var enough bool
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress,
		tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
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

func (tx *FeeVoteRevealTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {
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
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Expired
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
