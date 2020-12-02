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
	// SendMoney is Transaction Type that implemented TypeAction
	SendMoney struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		RecipientAddress     []byte
		Height               uint32
		Body                 *model.SendMoneyTransactionBody
		QueryExecutor        query.ExecutorInterface
		Escrow               *model.Escrow
		EscrowQuery          query.EscrowTransactionQueryInterface
		BlockQuery           query.BlockQueryInterface
		NormalFee            fee.FeeModelInterface
		EscrowFee            fee.FeeModelInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
)

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *SendMoney) SkipMempoolTransaction([]*model.Transaction, int64, uint32) (bool, error) {
	return false, nil
}

/*
ApplyConfirmed func that for applying Transaction SendMoney type.
If Genesis:
		- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
If Not Genesis:
		- perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
		`sender.balance` = current - amount
*/
func (tx *SendMoney) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err error
	)

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.RecipientAddress,
		tx.Body.GetAmount(),
		model.EventType_EventSendMoneyTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}
	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Body.GetAmount() + tx.Fee),
		model.EventType_EventSendMoneyTransaction,
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
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `SendMoney` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *SendMoney) ApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Body.GetAmount() + tx.Fee))
	if err != nil {
		return err
	}
	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *SendMoney) UndoApplyUnconfirmed() error {
	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Body.GetAmount()+tx.Fee)
	if err != nil {
		return err
	}
	return nil
}

/*
Validate is func that for validating to Transaction SendMoney type
That specs:
	- If Genesis, sender and recipient allowed not exists,
	- If Not Genesis,  sender and recipient must be exists, `sender.spendable_balance` must bigger than amount
*/
func (tx *SendMoney) Validate(dbTx bool) error {
	var (
		err error
	)
	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.RecipientAddress == nil {
		return errors.New("transaction must have a valid recipient account id")
	}

	// todo: this is temporary solution, later we should depend on coinbase, so no genesis transaction exclusion in
	// validation needed
	if !bytes.Equal(tx.SenderAddress, constant.MainchainGenesisAccountAddress) {
		if tx.SenderAddress == nil {
			return errors.New("transaction must have a valid sender account id")
		}

		enough, e := tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Body.GetAmount()+tx.Fee)
		if e != nil {
			if e != sql.ErrNoRows {
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
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SendMoney) GetAmount() int64 {
	return tx.Body.Amount
}

func (tx *SendMoney) GetMinimumFee() (int64, error) {
	if tx.Escrow != nil && tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

// GetSize send money Amount should be 8
func (*SendMoney) GetSize() (uint32, error) {
	// only amount
	return constant.Balance, nil
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *SendMoney) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
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
	return &model.SendMoneyTransactionBody{
		Amount: int64(amount),
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *SendMoney) GetBodyBytes() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.Amount)))
	return buffer.Bytes(), nil
}

// GetTransactionBody append isTransaction_TransactionBody oneOf
func (tx *SendMoney) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_SendMoneyTransactionBody{
		SendMoneyTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *SendMoney) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != nil && !bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		tx.Escrow = &model.Escrow{
			ID:               tx.ID,
			SenderAddress:    tx.SenderAddress,
			RecipientAddress: tx.RecipientAddress,
			ApproverAddress:  tx.Escrow.GetApproverAddress(),
			Amount:           tx.Body.GetAmount(),
			Commission:       tx.Escrow.GetCommission(),
			Timeout:          tx.Escrow.GetTimeout(),
			Status:           tx.Escrow.GetStatus(),
			BlockHeight:      tx.Height,
			Latest:           true,
			Instruction:      tx.Escrow.GetInstruction(),
		}

		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate special validation for escrow's transaction
func (tx *SendMoney) EscrowValidate(dbTx bool) error {
	var (
		err    error
		enough bool
	)

	if tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetRecipientAddress() == nil || bytes.Equal(tx.Escrow.GetRecipientAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "RecipientAddressRequired")
	}

	err = tx.Validate(dbTx)
	if err != nil {
		return err
	}
	enough, err = tx.AccountBalanceHelper.HasEnoughSpendableBalance(dbTx, tx.SenderAddress, tx.Body.GetAmount()+tx.Fee+tx.Escrow.GetCommission())
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
EscrowApplyUnconfirmed is applyUnconfirmed specific for Escrow's transaction
similar with ApplyUnconfirmed and Escrow.Commission
*/
func (tx *SendMoney) EscrowApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, -(tx.Body.GetAmount() + tx.Fee + tx.Escrow.GetCommission()))
	if err != nil {
		return err
	}
	return nil

}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *SendMoney) EscrowUndoApplyUnconfirmed() error {

	var err = tx.AccountBalanceHelper.AddAccountSpendableBalance(tx.SenderAddress, tx.Body.GetAmount()+tx.Escrow.GetCommission()+tx.Fee)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApplyConfirmed func that for applying Transaction SendMoney type, insert and update balance,
account ledger, and escrow
*/
func (tx *SendMoney) EscrowApplyConfirmed(blockTimestamp int64) error {
	var (
		err error
	)

	err = tx.AccountBalanceHelper.AddAccountBalance(
		tx.SenderAddress,
		-(tx.Body.GetAmount() + tx.Fee + tx.Escrow.GetCommission()),
		model.EventType_EventEscrowedTransaction,
		tx.Height,
		tx.ID,
		uint64(blockTimestamp),
	)
	if err != nil {
		return err
	}

	addEscrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(addEscrowQ)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *SendMoney) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {
	var (
		err error
	)

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.Escrow.Status = model.EscrowStatus_Approved
		// Bring back the fee that was decreased on EscrowApplyConfirmed before do ApplyConfirmed
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Body.GetAmount()+tx.Fee,
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
			tx.SenderAddress,
			tx.Body.Amount,
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
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
	default:
		tx.Escrow.Status = model.EscrowStatus_Expired
		err = tx.AccountBalanceHelper.AddAccountBalance(
			tx.SenderAddress,
			tx.Escrow.GetCommission()+tx.Body.Amount,
			model.EventType_EventApprovalEscrowTransaction,
			tx.Height,
			tx.ID,
			uint64(blockTimestamp),
		)
		if err != nil {
			return err
		}
	}

	addEscrowQ := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	err = tx.QueryExecutor.ExecuteTransactions(addEscrowQ)
	if err != nil {
		return err
	}
	return nil
}
