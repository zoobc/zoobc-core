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
	"math"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// LiquidPaymentTransaction is Transaction Type that implemented TypeAction
	LiquidPaymentTransaction struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		NormalFee                     fee.FeeModelInterface
		EscrowFee                     fee.FeeModelInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
	}
	LiquidPaymentTransactionInterface interface {
		CompletePayment(blockHeight uint32, blockTimestamp, firstAppliedTimestamp int64) error
	}
)

func (tx *LiquidPaymentTransaction) ApplyConfirmed(blockTimestamp int64) (err error) {

	// update sender
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.Body.Amount + tx.TransactionObject.Fee),
		model.EventType_EventLiquidPaymentTransaction,
		tx.TransactionObject.Height,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)

	if err != nil {
		return err
	}

	// create the Liquid payment record
	liquidPaymentTransaction := &model.LiquidPayment{
		ID:               tx.TransactionObject.ID,
		SenderAddress:    tx.TransactionObject.SenderAccountAddress,
		RecipientAddress: tx.TransactionObject.RecipientAccountAddress,
		Amount:           tx.Body.GetAmount(),
		AppliedTime:      blockTimestamp,
		CompleteMinutes:  tx.Body.GetCompleteMinutes(),
		Status:           model.LiquidPaymentStatus_LiquidPaymentPending,
		BlockHeight:      tx.TransactionObject.Height,
	}
	liquidPaymentTransactionQ := tx.LiquidPaymentTransactionQuery.InsertLiquidPaymentTransaction(liquidPaymentTransaction)
	err = tx.QueryExecutor.ExecuteTransactions(liquidPaymentTransactionQ)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) ApplyUnconfirmed() (err error) {
	// update sender
	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, -(tx.Body.Amount + tx.TransactionObject.Fee))
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) UndoApplyUnconfirmed() (err error) {
	// update sender
	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.TransactionObject.SenderAccountAddress, tx.Body.Amount+tx.TransactionObject.Fee)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) Validate(dbTx bool) error {
	var (
		err    error
		enough bool
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.TransactionObject.SenderAccountAddress == nil {
		return errors.New("transaction must have a valid sender account id")
	}
	if tx.TransactionObject.RecipientAccountAddress == nil {
		return errors.New("transaction must have a valid recipient account id")
	}

	// check existing & balance account sender
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.TransactionObject.SenderAccountAddress, tx.Body.GetAmount()+tx.TransactionObject.Fee)
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

func (tx *LiquidPaymentTransaction) GetMinimumFee() (int64, error) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.TransactionObject)
}

func (tx *LiquidPaymentTransaction) GetAmount() int64 {
	return tx.Body.Amount
}

func (tx *LiquidPaymentTransaction) GetSize() (uint32, error) {
	// only amount
	return constant.Balance + constant.LiquidPaymentCompleteMinutesLength, nil
}

func (tx *LiquidPaymentTransaction) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
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
	amount := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.Balance)))
	completeMinutes := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.LiquidPaymentCompleteMinutesLength)))
	return &model.LiquidPaymentTransactionBody{
		Amount:          int64(amount),
		CompleteMinutes: completeMinutes,
	}, nil
}

func (tx *LiquidPaymentTransaction) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.Amount)))
	buffer.Write(util.ConvertUint64ToBytes(tx.Body.CompleteMinutes))
	return buffer.Bytes(), nil
}

func (tx *LiquidPaymentTransaction) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_LiquidPaymentTransactionBody{
		LiquidPaymentTransactionBody: tx.Body,
	}
}

// SkipMempoolTransaction filter out of the mempool tx under specific condition
func (tx *LiquidPaymentTransaction) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

func (tx *LiquidPaymentTransaction) CompletePayment(blockHeight uint32, blockTimestamp, firstAppliedTimestamp int64) error {
	var (
		err                                               error
		recipientBalanceIncrement, senderBalanceIncrement int64
		blockTimestampTime                                = time.Unix(blockTimestamp, 0)
		firstAppliedTimestampTime                         = time.Unix(firstAppliedTimestamp, 0)
		durationPassed                                    = blockTimestampTime.Sub(firstAppliedTimestampTime).Minutes()
	)

	if durationPassed < 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "blockTimestamp is less than firstAppliedTimestamp")
	}

	durationRate := durationPassed / float64(tx.Body.GetCompleteMinutes())
	if durationRate > 1 {
		recipientBalanceIncrement = tx.Body.GetAmount()
	} else {
		recipientBalanceIncrement = int64(math.Ceil(durationRate * float64(tx.Body.GetAmount())))
		senderBalanceIncrement = tx.Body.GetAmount() - recipientBalanceIncrement
	}

	// transfer the money to the recipient pro-rate wise
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.RecipientAccountAddress,
		recipientBalanceIncrement,
		model.EventType_EventLiquidPaymentPaidTransaction,
		blockHeight,
		tx.TransactionObject.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	if senderBalanceIncrement > 0 {
		// returning the remaining payment to the sender
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			senderBalanceIncrement,
			model.EventType_EventLiquidPaymentPaidTransaction,
			blockHeight,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}

	// update the status of the liquid payment
	liquidPaymentStatusUpdateQ := tx.LiquidPaymentTransactionQuery.CompleteLiquidPaymentTransaction(
		tx.TransactionObject.ID,
		map[string]interface{}{"block_height": blockHeight},
	)

	err = tx.QueryExecutor.ExecuteTransactions(liquidPaymentStatusUpdateQ)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) Escrowable() (EscrowTypeAction, bool) {
	if tx.TransactionObject.Escrow != nil && tx.TransactionObject.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.TransactionObject.Escrow.GetApproverAddress(), []byte{}) {
		tx.TransactionObject.Escrow = util.PrepareEscrowObjectForAction(tx.TransactionObject)
		return EscrowTypeAction(tx), true
	}
	return nil, false
}

func (tx *LiquidPaymentTransaction) EscrowApplyConfirmed(blockTimestamp int64) (err error) {
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.Body.Amount + tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
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

func (tx *LiquidPaymentTransaction) EscrowApplyUnconfirmed() (err error) {
	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(
		tx.TransactionObject.SenderAccountAddress,
		-(tx.Body.Amount + tx.TransactionObject.Fee + tx.TransactionObject.Escrow.GetCommission()),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) EscrowUndoApplyUnconfirmed() (err error) {
	err = tx.AccountBalanceHelper.AddAccountSpendableBalance(
		tx.TransactionObject.SenderAccountAddress,
		tx.Body.Amount+tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (tx *LiquidPaymentTransaction) EscrowValidate(dbTx bool) (err error) {
	var enough bool
	err = util.ValidateBasicEscrow(tx.TransactionObject)
	if err != nil {
		return err
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(
		dbTx,
		tx.TransactionObject.SenderAccountAddress,
		tx.Body.GetAmount()+tx.TransactionObject.Fee+tx.TransactionObject.Escrow.GetCommission(),
	)
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

func (tx *LiquidPaymentTransaction) EscrowApproval(blockTimestamp int64, txBody *model.ApprovalEscrowTransactionBody) (err error) {

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Approved
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			tx.Body.Amount+tx.TransactionObject.Fee,
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
			tx.TransactionObject.SenderAccountAddress,
			tx.Body.Amount,
			model.EventType_EventApprovalEscrowTransaction,
			tx.TransactionObject.Height,
			tx.TransactionObject.ID,
			uint64(blockTimestamp),
		)
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
	default:
		tx.TransactionObject.Escrow.Status = model.EscrowStatus_Expired
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.TransactionObject.SenderAccountAddress,
			tx.Body.GetAmount()+tx.TransactionObject.Escrow.GetCommission(),
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
