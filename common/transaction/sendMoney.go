package transaction

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// SendMoney is Transaction Type that implemented TypeAction
type SendMoney struct {
	Body                 *model.SendMoneyTransactionBody
	Fee                  int64
	SenderAddress        string
	SenderAccountType    uint32
	RecipientAddress     string
	RecipientAccountType uint32
	Height               uint32
	AccountBalanceQuery  query.AccountBalanceQueryInterface
	AccountQuery         query.AccountQueryInterface
	QueryExecutor        query.ExecutorInterface
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
		recipientAccount model.Account
		senderAccount    model.Account
		err              error
	)

	if err := tx.Validate(); err != nil {
		return err
	}

	if tx.Height > 0 {
		err = tx.UndoApplyUnconfirmed()
		if err != nil {
			return err
		}
	}

	recipientAccount = model.Account{
		ID:          util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
		AccountType: tx.RecipientAccountType,
		Address:     tx.RecipientAddress,
	}
	senderAccount = model.Account{
		ID:          util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress),
		AccountType: tx.SenderAccountType,
		Address:     tx.SenderAddress,
	}

	// insert / update recipient
	recipientAccountInsertQ, recipientAccountInsertArgs := tx.AccountQuery.InsertAccount(&recipientAccount)
	accountBalanceRecipientQ := tx.AccountBalanceQuery.AddAccountBalance(
		tx.Body.Amount,
		map[string]interface{}{
			"account_id":   recipientAccount.ID,
			"block_height": tx.Height,
		},
	)
	// update sender
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.Amount + tx.Fee),
		map[string]interface{}{
			"account_id":   senderAccount.ID,
			"block_height": tx.Height,
		},
	)
	queries := append(append(accountBalanceRecipientQ, accountBalanceSenderQ...),
		append([]interface{}{recipientAccountInsertQ}, recipientAccountInsertArgs...),
	)
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
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
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
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
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
func (tx *SendMoney) Validate() error {

	var (
		err            error
		accountBalance model.AccountBalance
		count          int
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.RecipientAddress == "" {
		return errors.New("transaction must have a valid recipient account id")
	}

	if tx.Height != 0 {
		if tx.SenderAddress == "" {
			return errors.New("transaction must have a valid sender account id")
		}

		accounts, accountArgs := tx.AccountQuery.GetAccountByIDs([][]byte{
			util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress),
			util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
		})

		err = tx.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(accounts), accountArgs...).Scan(&count)
		if err != nil {
			return err
		}

		if count <= 1 {
			return fmt.Errorf("count recipient and sender got: %d", count)
		}
		senderID := util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress)
		senderQ, senderArg := tx.AccountBalanceQuery.GetAccountBalanceByAccountID(senderID)
		rows, err := tx.QueryExecutor.ExecuteSelect(senderQ, senderArg)
		if err != nil {
			return err
		} else if rows.Next() {
			_ = rows.Scan(
				&accountBalance.AccountID,
				&accountBalance.BlockHeight,
				&accountBalance.SpendableBalance,
				&accountBalance.Balance,
				&accountBalance.PopRevenue,
				&accountBalance.Latest,
			)
		}
		defer rows.Close()

		if accountBalance.SpendableBalance < tx.Body.GetAmount() {
			return errors.New("transaction amount not enough")
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
	// only amount (int64)
	return 8
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*SendMoney) ParseBodyBytes(txBodyBytes []byte) *model.SendMoneyTransactionBody {
	amount := util.ConvertBytesToUint64(txBodyBytes)
	return &model.SendMoneyTransactionBody{
		Amount: int64(amount),
	}
}

// GetBodyBytes translate tx body to bytes representation
func (*SendMoney) GetBodyBytes(txBody *model.SendMoneyTransactionBody) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(txBody.Amount)))
	return buffer.Bytes()
}
