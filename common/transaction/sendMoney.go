package transaction

import (
	"bytes"
	"errors"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// SendMoney is Transaction Type that implemented TypeAction
type SendMoney struct {
	Body                *model.SendMoneyTransactionBody
	Fee                 int64
	SenderAddress       string
	RecipientAddress    string
	Height              uint32
	AccountBalanceQuery query.AccountBalanceQueryInterface
	QueryExecutor       query.ExecutorInterface
}

// FilterMempoolTransaction this tx type has no mempool filter
func (tx *SendMoney) FilterMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
	return true, nil
}

/*
ApplyConfirmed is func that for applying Transaction SendMoney type,

__If Genesis__:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.

__If Not Genesis__:
	- perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
	`sender.balance` = current - amount
*/
func (tx *SendMoney) ApplyConfirmed() error {
	var (
		err error
	)

	// insert / update recipient
	accountBalanceRecipientQ := tx.AccountBalanceQuery.AddAccountBalance(
		tx.Body.Amount,
		map[string]interface{}{
			"account_address": tx.RecipientAddress,
			"block_height":    tx.Height,
		},
	)
	// update sender
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.Amount + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries := append(accountBalanceRecipientQ, accountBalanceSenderQ...)
	err = tx.QueryExecutor.ExecuteTransactions(queries)

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

	var (
		err error
	)

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Body.Amount + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
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
	var (
		err error
	)

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Body.Amount+tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
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
		accountBalance model.AccountBalance
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.RecipientAddress == "" {
		return errors.New("transaction must have a valid recipient account id")
	}
	// todo: this is temporary solution, later we should depend on coinbase, so no genesis transaction exclusion in
	// validation needed
	if tx.SenderAddress != constant.MainchainGenesisAccountAddress {
		if tx.SenderAddress == "" {
			return errors.New("transaction must have a valid sender account id")
		}

		qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
		rows, err := tx.QueryExecutor.ExecuteSelect(qry, dbTx, args...)
		if err != nil {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		} else if rows.Next() {
			_ = rows.Scan(
				&accountBalance.AccountAddress,
				&accountBalance.BlockHeight,
				&accountBalance.SpendableBalance,
				&accountBalance.Balance,
				&accountBalance.PopRevenue,
				&accountBalance.Latest,
			)
		}
		defer rows.Close()

		if accountBalance.SpendableBalance < (tx.Body.GetAmount() + tx.Fee) {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"balance not enough",
			)
		}
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SendMoney) GetAmount() int64 {
	return tx.Body.Amount
}

// GetSize send money Amount should be 8
func (*SendMoney) GetSize() uint32 {
	// only amount
	return constant.Balance
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *SendMoney) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// validate the body bytes is correct
	_, err := util.ReadTransactionBytes(bytes.NewBuffer(txBodyBytes), int(tx.GetSize()))
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
func (tx *SendMoney) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.Amount)))
	return buffer.Bytes()
}

func (tx *SendMoney) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_SendMoneyTransactionBody{
		SendMoneyTransactionBody: tx.Body,
	}
}
