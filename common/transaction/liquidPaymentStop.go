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
	"errors"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// LiquidPaymentStopTransaction is Transaction Type that implemented TypeAction
	LiquidPaymentStopTransaction struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
		EscrowQuery                   query.EscrowTransactionQueryInterface
		FeeScaleService               fee.FeeScaleServiceInterface
	}
)

func (tx *LiquidPaymentStopTransaction) ApplyConfirmed(blockTimestamp int64) error {
	var (
		row           *sql.Row
		err           error
		liquidPayment model.LiquidPayment
		transaction   model.Transaction
		txType        TypeAction
	)

	// update sender
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-tx.TransactionObject.Fee,
		model.EventType_EventLiquidPaymentStopTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	// processing the liquid payment transaction
	liquidPaymentQ, liquidPaymentArgs := tx.LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID(
		tx.Body.TransactionID,
		model.LiquidPaymentStatus_LiquidPaymentPending,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(liquidPaymentQ, true, liquidPaymentArgs...)
	if err != nil {
		return err
	}
	err = tx.LiquidPaymentTransactionQuery.Scan(&liquidPayment, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "LiquidPaymentNotExists")
		}
		return err
	}

	// handle multiple stop transaction
	if liquidPayment.Status == model.LiquidPaymentStatus_LiquidPaymentCompleted {
		return nil
	}

	// get what transaction type it is, and switch to specific approval
	transactionQ := tx.TransactionQuery.GetTransaction(tx.Body.GetTransactionID())
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

	txType, err = tx.TypeActionSwitcher.GetTransactionType(&transaction)
	if err != nil {
		return err
	}
	liquidPaymentTransaction, ok := txType.(LiquidPaymentTransactionInterface)
	if !ok {
		return blocker.NewBlocker(blocker.AppErr, "Wrong type of transaction")
	}
	err = liquidPaymentTransaction.CompletePayment(tx.TransactionObject.Height, blockTimestamp, liquidPayment.AppliedTime)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) ApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -(tx.TransactionObject.Fee))
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) UndoApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) Validate(dbTx bool) error {
	var (
		row           *sql.Row
		err           error
		liquidPayment model.LiquidPayment
		enough        bool
	)
	if tx.TransactionObject.SenderAccountAddress == nil {
		return errors.New("transaction must have a valid sender account id")
	}

	if tx.Body.TransactionID == 0 {
		return errors.New("transaction must have a valid transaction id")
	}

	liquidPaymentQ, liquidPaymentArgs := tx.LiquidPaymentTransactionQuery.GetPendingLiquidPaymentTransactionByID(
		tx.Body.TransactionID,
		model.LiquidPaymentStatus_LiquidPaymentPending,
	)
	row, err = tx.QueryExecutor.ExecuteSelectRow(liquidPaymentQ, dbTx, liquidPaymentArgs...)
	if err != nil {
		return err
	}
	err = tx.LiquidPaymentTransactionQuery.Scan(&liquidPayment, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "LiquidPaymentNotExists")
		}
		return err
	}

	if !bytes.Equal(liquidPayment.SenderAddress, tx.TransactionObject.SenderAccountAddress) &&
		!bytes.Equal(liquidPayment.RecipientAddress, tx.TransactionObject.SenderAccountAddress) {
		return blocker.NewBlocker(blocker.ValidationErr, "Only sender or recipient of the payment can stop the payment")
	}

	if liquidPayment.Status == model.LiquidPaymentStatus_LiquidPaymentCompleted {
		return blocker.NewBlocker(blocker.ValidationErr, "LiquidPaymentHasPreviouslyCompleted")
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

func (tx *LiquidPaymentStopTransaction) GetMinimumFee() (int64, error) {
	var lastFeeScale model.FeeScale
	err := tx.FeeScaleService.GetLatestFeeScale(&lastFeeScale)
	if err != nil {
		return 0, err
	}
	return fee.CalculateTxMinimumFee(tx.TransactionObject, lastFeeScale.FeeScale)
}

func (tx *LiquidPaymentStopTransaction) GetAmount() int64 {
	return tx.TransactionObject.Fee
}

func (tx *LiquidPaymentStopTransaction) GetSize() (uint32, error) {
	// only TransactionID
	return constant.TransactionID, nil
}

func (tx *LiquidPaymentStopTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// validate the body bytes is correct
	txSize, err := tx.GetSize()
	if err != nil {
		return nil, err
	}
	_, err = util.ReadTransactionBytes(bytes.NewBuffer(txBodyBytes), int(txSize))
	if err != nil {
		return nil, err
	}
	// read body bytes
	bufferBytes := bytes.NewBuffer(txBodyBytes)
	txID := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.Balance)))
	return &model.LiquidPaymentStopTransactionBody{
		TransactionID: int64(txID),
	}, nil
}

func (tx *LiquidPaymentStopTransaction) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.TransactionID)))
	return buffer.Bytes(), nil
}

func (tx *LiquidPaymentStopTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_LiquidPaymentStopTransactionBody{
		LiquidPaymentStopTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction filter out of the mempool tx under specific condition
func (tx *LiquidPaymentStopTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

func (tx *LiquidPaymentStopTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil &&
		tx.TransactionObject.Escrow.GetApproverAddress() != nil &&
		!bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *LiquidPaymentStopTransaction) EscrowApplyConfirmed(blockTimestamp int64) (err error) {
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

func (tx *LiquidPaymentStopTransaction) EscrowApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) EscrowUndoApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress,
		tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission())
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentStopTransaction) EscrowValidate(dbTx bool) (err error) {
	var enough bool

	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx,
		tx.TransactionObject.SenderAccountAddress,
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

func (tx *LiquidPaymentStopTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {

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
