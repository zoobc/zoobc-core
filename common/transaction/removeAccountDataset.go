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

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// RemoveAccountDataset has fields that's needed
type RemoveAccountDataset struct {
	ID                   int64
	Fee                  int64
	SenderAddress        []byte
	RecipientAddress     []byte
	Height               uint32
	Body                 *model.RemoveAccountDatasetTransactionBody
	Escrow               *model.Escrow
	AccountDatasetQuery  query.AccountDatasetQueryInterface
	QueryExecutor        query.ExecutorInterface
	EscrowQuery          query.EscrowTransactionQueryInterface
	AccountBalanceHelper AccountBalanceHelperInterface
	EscrowFee            fee.FeeModelInterface
	NormalFee            fee.FeeModelInterface
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *RemoveAccountDataset) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

/*
ApplyConfirmed is func that for applying Transaction RemoveAccountDataset type,
*/
func (tx *RemoveAccountDataset) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err error
	)

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Fee),
		model.EventType_EventRemoveAccountDatasetTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	// Account dataset removed, need to set IsActive false
	datasetQ := tx.AccountDatasetQuery.InsertAccountDataset(&model.AccountDataset{
		SetterAccountAddress:    tx.SenderAddress,
		RecipientAccountAddress: tx.RecipientAddress,
		Property:                tx.Body.GetProperty(),
		Value:                   tx.Body.GetValue(),
		Height:                  tx.Height,
		Latest:                  true,
		IsActive:                false,
	})

	err = tx.QueryExecutor.ExecuteTransactions(datasetQ)
	if err != nil {
		return err
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveAccountDataset` type
*/
func (tx *RemoveAccountDataset) ApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Fee))
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *RemoveAccountDataset) UndoApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return nil
}

/*
Validate is func that for validating to Transaction RemoveAccountDataset type
That specs:
	- Check existing Account Dataset
	- Check Spendable Balance sender
*/
func (tx *RemoveAccountDataset) Validate(dbTx, checkOnSpendableBalance bool) error {
	var (
		accountDataset model.AccountDataset
		err            error
		row            *sql.Row
		qry            string
		qryArgs        []interface{}
		enough         bool
	)

	/*
		Check existing dataset
		Account Dataset can only delete when account dataset exist
	*/
	qry, qryArgs = tx.AccountDatasetQuery.GetLatestAccountDataset(
		tx.SenderAddress,
		tx.RecipientAddress,
		tx.Body.GetProperty(),
	)

	// NOTE: currently dbTx became true only when calling on push block,
	// this is will make allow to execute all of same tx in mempool if all of them selected
	// TODO: should be using skip mempool to check double same tx in mempool
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, qryArgs...)
	if err != nil {
		return err
	}
	err = tx.AccountDatasetQuery.Scan(&accountDataset, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		return blocker.NewBlocker(blocker.ValidationErr, "AccountDatasetNotExists")
	}
	// !false if err in above is sql.ErrNoRows || nil
	if !accountDataset.GetIsActive() {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountDatasetAlreadyRemoved")
	}
	// checkOnSpendableBalance will check to the spendable balance of the sender otherwise will check the actual balance
	if checkOnSpendableBalance {
		enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Fee)
	} else {
		enough, err = tx.AccountBalanceHelper.HasEnoughBalance(dbTx, tx.SenderAddress, tx.Fee)
	}
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
	}
	if !enough {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"UserBalanceNotEnough",
		)
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *RemoveAccountDataset) GetAmount() int64 {
	return 0
}

// GetMinimumFee return minimum fee of transaction
// TODO: need to calculate the minimum fee
func (tx *RemoveAccountDataset) GetMinimumFee() (int64, error) {
	if tx.Escrow != nil && tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

// GetSize is size of transaction body
func (tx *RemoveAccountDataset) GetSize() (uint32, error) {
	txBodyBytes, err := tx.GetBodyBytes()
	if err != nil {
		return 0, err
	}
	return uint32(len(txBodyBytes)), nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *RemoveAccountDataset) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		err          error
		chunkedBytes []byte
		dataLength   uint32
		txBody       model.RemoveAccountDatasetTransactionBody
		buffer       = bytes.NewBuffer(txBodyBytes)
	)
	// get length of property dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.DatasetPropertyLength))
	if err != nil {
		return nil, err
	}
	dataLength = util.ConvertBytesToUint32(chunkedBytes)
	// get property of dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
	if err != nil {
		return nil, err
	}
	txBody.Property = string(chunkedBytes)
	// get length of value property dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.DatasetValueLength))
	if err != nil {
		return nil, err
	}
	dataLength = util.ConvertBytesToUint32(chunkedBytes)
	// get value property of dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
	if err != nil {
		return nil, err
	}
	txBody.Value = string(chunkedBytes)
	return &txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *RemoveAccountDataset) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetProperty())))))
	buffer.Write([]byte(tx.Body.GetProperty()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetValue())))))
	buffer.Write([]byte(tx.Body.GetValue()))

	return buffer.Bytes(), nil
}

// GetTransactionBody return transaction body of RemoveAccountDataset transactions
func (tx *RemoveAccountDataset) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_RemoveAccountDatasetTransactionBody{
		RemoveAccountDatasetTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *RemoveAccountDataset) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		tx.Escrow = &model.Escrow{
			ID:              tx.ID,
			SenderAddress:   tx.SenderAddress,
			ApproverAddress: tx.Escrow.GetApproverAddress(),
			Commission:      tx.Escrow.GetCommission(),
			Timeout:         tx.Escrow.GetTimeout(),
			Status:          tx.Escrow.GetStatus(),
			BlockHeight:     tx.Height,
			Latest:          true,
			Instruction:     tx.Escrow.GetInstruction(),
		}

		return EscrowTypeAction(tx), true
	}
	return nil, false
}

/*
EscrowValidate is func that for validating to Transaction RemoveAccountDataset type
That specs:
	- Check existing Account Dataset
	- Check Spendable Balance sender
*/
func (tx *RemoveAccountDataset) EscrowValidate(dbTx, checkOnSpendableBalance bool) error {
	var (
		err    error
		enough bool
	)

	if tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetTimeout() > uint64(constant.MinRollbackBlocks) {
		return blocker.NewBlocker(blocker.ValidationErr, "TimeoutLimitExceeded")
	}

	err = tx.Validate(dbTx, checkOnSpendableBalance)
	if err != nil {
		return err
	}
	// checkOnSpendableBalance will check to the spendable balance of the sender otherwise will check the actual balance
	if checkOnSpendableBalance {
		enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
	} else {
		enough, err = tx.AccountBalanceHelper.HasEnoughBalance(dbTx, tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
	}

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

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveAccountDataset` type
*/
func (tx *RemoveAccountDataset) EscrowApplyUnconfirmed() error {

	err := tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Fee + tx.Escrow.GetCommission()))
	if err != nil {
		return err
	}

	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *RemoveAccountDataset) EscrowUndoApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Fee+tx.Escrow.GetCommission())
	if err != nil {
		return err
	}

	return nil
}

/*
EscrowApplyConfirmed is func that for applying Transaction RemoveAccountDataset type,
*/
func (tx *RemoveAccountDataset) EscrowApplyConfirmed(blockTimestamp int64) error {

	var err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Fee + tx.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *RemoveAccountDataset) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {

	var (
		err error
	)

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Fee,
			model.EventType_EventEscrowedTransaction,
			tx.Height,
			tx.ID,
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
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	case model.EscrowApproval_Reject:
		tx.Escrow.Status = model.EscrowStatus_Rejected
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.Escrow.GetApproverAddress(),
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	default:
		tx.Escrow.Status = model.EscrowStatus_Expired
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Escrow.GetCommission(),
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}

	// Insert Escrow
	escrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(escrowQ)
	if err != nil {
		return err
	}

	return nil
}
