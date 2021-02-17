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

type (
	// ApprovalEscrowTransaction field
	ApprovalEscrowTransaction struct {
		TransactionObject    *model.Transaction
		Body                 *model.ApprovalEscrowTransactionBody
		BlockQuery           query.BlockQueryInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		QueryExecutor        query.ExecutorInterface
		TransactionQuery     query.TransactionQueryInterface
		TypeActionSwitcher   TypeActionSwitcher
		AccountBalanceHelper AccountBalanceHelperInterface
		FeeScaleService      fee.FeeScaleServiceInterface
	}
	// EscrowTypeAction is escrow transaction type methods collection
	EscrowTypeAction interface {
		// EscrowApplyConfirmed perhaps this method called with QueryExecutor.BeginTX() because inside this process has separated QueryExecutor.Execute
		EscrowApplyConfirmed(blockTimestamp int64) error
		EscrowApplyUnconfirmed() error
		EscrowUndoApplyUnconfirmed() error
		EscrowValidate(dbTx bool) error
		// EscrowApproval handle approval an escrow transaction, execute tasks that was skipped on EscrowApplyConfirmed.
		EscrowApproval(
			blockTimestamp int64,
			txBody *model.ApprovalEscrowTransactionBody,
		) error
	}
)

// SkipMempoolTransaction to filter out current Approval escrow transaction when
// this tx already expired based on new block height
func (tx *ApprovalEscrowTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	var (
		err = tx.checkEscrowValidity(false, newBlockHeight)
	)
	if err != nil {
		return true, err
	}
	return false, nil
}

// GetSize of approval transaction body bytes
func (*ApprovalEscrowTransaction) GetSize() (uint32, error) {
	return constant.EscrowApprovalBytesLength, nil
}

func (tx *ApprovalEscrowTransaction) GetMinimumFee() (int64, error) {
	var lastFeeScale model.FeeScale
	err := tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return 0, err
	}
	return fee.CalculateTxMinimumFee(tx.TransactionObject, lastFeeScale.FeeScale)
}

// GetAmount return Amount from TransactionBody
func (tx *ApprovalEscrowTransaction) GetAmount() int64 {
	return 0
}

// GetBodyBytes translate tx body to bytes representation
func (tx *ApprovalEscrowTransaction) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(tx.Body.GetApproval())))
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.GetTransactionID())))
	return buffer.Bytes(), nil
}

// GetTransactionBody append isTransaction_TransactionBody oneOf
func (tx *ApprovalEscrowTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_ApprovalEscrowTransactionBody{
		ApprovalEscrowTransactionBody: tx.Body,
	}
}

// ParseBodyBytes validate and parse body bytes to TransactionBody interface
func (tx *ApprovalEscrowTransaction) ParseBodyBytes(
	bodyBytes []byte,
) (model.TransactionBodyInterface, error) {
	var (
		buffer  = bytes.NewBuffer(bodyBytes)
		chunked []byte
		err     error
	)

	chunked, err = util.ReadTransactionBytes(buffer, int(constant.EscrowApproval))
	if err != nil {
		return nil, err
	}
	approvalInt := util.ConvertBytesToUint32(chunked)

	chunked, err = util.ReadTransactionBytes(buffer, int(constant.EscrowID))
	if err != nil {
		return nil, err
	}
	escrowID := util.ConvertBytesToUint64(chunked)

	return &model.ApprovalEscrowTransactionBody{
		Approval:      model.EscrowApproval(approvalInt),
		TransactionID: int64(escrowID),
	}, nil
}

/*
Validate is func that for validating to Transaction type.
Check transaction fields, spendable balance and more
*/
func (tx *ApprovalEscrowTransaction) Validate(dbTx bool) error {
	var (
		err    error
		enough bool
	)
	err = tx.checkEscrowValidity(dbTx, tx.TransactionObject.Height)
	if err != nil {
		return err
	}
	// check existing account & balance

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

func (tx *ApprovalEscrowTransaction) checkEscrowValidity(dbTx bool, blockHeight uint32) error {
	var (
		latestEscrow        model.Escrow
		row                 *sql.Row
		err                 error
		escrowQ, escrowArgs = tx.EscrowQuery.GetLatestEscrowTransactionByID(tx.Body.GetTransactionID())
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(escrowQ, dbTx, escrowArgs...)
	if err != nil {
		return err
	}

	// Check escrow exists
	err = tx.EscrowQuery.Scan(&latestEscrow, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.ValidationErr, "EscrowNotExists")
	}
	if blockHeight >= latestEscrow.GetBlockHeight()+uint32(latestEscrow.Timeout) {
		return blocker.NewBlocker(blocker.ValidationErr, "EscrowTimeout")
	}

	// Check escrow status still pending before allow to apply
	if latestEscrow.GetStatus() != model.EscrowStatus_Pending {
		return blocker.NewBlocker(blocker.ValidationErr, "EscrowTargetNotValidByStatus")
	}

	// Check sender, should be approver address
	if !bytes.Equal(latestEscrow.GetApproverAddress(), tx.TransactionObject.SenderAccountAddress) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSenderAddress")
	}

	// check transaction id is valid
	if latestEscrow.GetID() != tx.Body.GetTransactionID() {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidTransactionID")
	}
	return nil
}

/*
ApplyUnconfirmed exec before Confirmed
*/
func (tx *ApprovalEscrowTransaction) ApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -tx.TransactionObject.Fee)
}

/*
UndoApplyUnconfirmed func exec before confirmed
*/
func (tx *ApprovalEscrowTransaction) UndoApplyUnconfirmed() error {
	return tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
}

/*
ApplyConfirmed func that for applying Transaction SendZBC type.
If Genesis perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
If Not Genesis, perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
`sender.balance` = current - amount
*/
func (tx *ApprovalEscrowTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		latestEscrow model.Escrow
		transaction  model.Transaction
		txType       TypeAction
		row          *sql.Row
		err          error
	)

	// Get escrow by reference transaction ID
	escrowQ, escrowArgs := tx.EscrowQuery.GetLatestEscrowTransactionByID(tx.Body.GetTransactionID())
	row, err = tx.QueryExecutor.ExecuteSelectRow(escrowQ, false, escrowArgs...)
	if err != nil {
		return err
	}

	err = tx.EscrowQuery.Scan(&latestEscrow, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "EscrowNotFound")
	}

	// get what transaction type it is, and switch to specific approval
	transactionQ := tx.TransactionQuery.GetTransaction(latestEscrow.GetID())
	row, err = tx.QueryExecutor.ExecuteSelectRow(transactionQ, false)
	if err != nil {
		return err
	}
	err = tx.TransactionQuery.Scan(&transaction, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		return blocker.NewBlocker(blocker.AppErr, "TransactionNotFound")

	}
	transaction.Height = tx.TransactionObject.Height
	transaction.Escrow = &latestEscrow

	txType, err = tx.TypeActionSwitcher.GetTransactionType(&transaction)
	if err != nil {
		return err
	}

	// now only send money has EscrowApproval method
	escrowable, ok := txType.Escrowable()
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "ExpectEscrowableTransaction")
	}
	err = escrowable.EscrowApproval(blockTimestamp, tx.Body)
	if err != nil {
		return blocker.NewBlocker(blocker.AppErr, "EscrowApprovalFailed")
	}

	// Update sender
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-tx.TransactionObject.Fee,
		model.EventType_EventApprovalEscrowTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	return nil
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *ApprovalEscrowTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil &&
		!bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate special validation for escrow's transaction
func (tx *ApprovalEscrowTransaction) EscrowValidate(dbTx bool) error {

	return nil
}

/*
EscrowApplyUnconfirmed is applyUnconfirmed specific for Escrow's transaction
similar with ApplyUnconfirmed and Escrow.Commission
*/
func (tx *ApprovalEscrowTransaction) EscrowApplyUnconfirmed() error {

	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *ApprovalEscrowTransaction) EscrowUndoApplyUnconfirmed() error {

	return nil
}

/*
EscrowApplyConfirmed func that for applying Transaction SendZBC type.
*/
func (tx *ApprovalEscrowTransaction) EscrowApplyConfirmed(int64) error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *ApprovalEscrowTransaction) EscrowApproval(int64, *model.ApprovalEscrowTransactionBody) error {
	return nil
}
