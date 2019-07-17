package transaction

import (
	"errors"
	"fmt"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// SendMoney is Transaction Type that implemented TypeAction
type SendMoney struct {
	Body                 *model.SendMoneyTransactionBody
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
		accountBalance model.AccountBalance
		account        model.Account
		err            error
	)

	if err := tx.Validate(); err != nil {
		return err
	}

	if tx.Height == 0 {
		account = model.Account{
			ID:          util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
			AccountType: tx.RecipientAccountType,
			Address:     tx.RecipientAddress,
		}
		accountBalance = model.AccountBalance{
			AccountID:        account.ID,
			BlockHeight:      tx.Height,
			SpendableBalance: tx.Body.GetAmount(),
			Balance:          tx.Body.GetAmount(),
			PopRevenue:       0,
		}
		accountQ, accountQArgs := tx.AccountQuery.InsertAccount(&account)
		accountBalanceQ, accountBalanceArgs := tx.AccountBalanceQuery.InsertAccountBalance(&accountBalance)
		_, err = tx.QueryExecutor.ExecuteTransactionStatements(map[*string][]interface{}{
			&accountQ:        accountQArgs,
			&accountBalanceQ: accountBalanceArgs,
		})
		if err != nil {
			return err
		}
	} else {
		// update recipient
		accountBalanceRecipientQ, accountBalanceRecipientQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
			map[string]interface{}{
				"balance": fmt.Sprintf("balance + %d", tx.Body.GetAmount()),
			},
			map[string]interface{}{
				"account_id": util.CreateAccountIDFromAddress(
					tx.RecipientAccountType,
					tx.RecipientAddress,
				),
			},
		)
		// update sender
		accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
			map[string]interface{}{
				"balance": fmt.Sprintf("balance - %d", tx.Body.GetAmount()),
			},
			map[string]interface{}{
				"account_id": util.CreateAccountIDFromAddress(
					tx.SenderAccountType,
					tx.SenderAddress,
				),
			},
		)
		_, err = tx.QueryExecutor.ExecuteTransactionStatements(map[*string][]interface{}{
			&accountBalanceSenderQ:    accountBalanceSenderQArgs,
			&accountBalanceRecipientQ: accountBalanceRecipientQArgs,
		})
		if err != nil {
			return err
		}
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

	if err := tx.Validate(); err != nil {
		return err
	}

	// update recipient
	accountBalanceRecipientQ, accountBalanceRecipientQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
		map[string]interface{}{
			"account_balance": fmt.Sprintf("account_balance + %d", tx.Body.GetAmount()),
		},
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.RecipientAccountType,
				tx.RecipientAddress,
			),
		},
	)
	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
		map[string]interface{}{
			"account_balance": fmt.Sprintf("account_balance - %d", tx.Body.GetAmount()),
		},
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	_, err = tx.QueryExecutor.ExecuteTransactionStatements(map[*string][]interface{}{
		&accountBalanceSenderQ:    accountBalanceSenderQArgs,
		&accountBalanceRecipientQ: accountBalanceRecipientQArgs,
	})
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

		err = tx.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(accounts), accountArgs).Scan(&count)
		if err != nil {
			return err
		}

		if count <= 1 {
			return fmt.Errorf("count recipient and sender got: %d", count)
		}

		if rows, err := tx.QueryExecutor.ExecuteSelect(
			tx.AccountBalanceQuery.GetAccountBalanceByAccountID(),
			util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress),
		); err != nil {
			return err
		} else if rows.Next() {
			_ = rows.Scan(
				&accountBalance.AccountID,
				&accountBalance.BlockHeight,
				&accountBalance.SpendableBalance,
				&accountBalance.Balance,
				&accountBalance.PopRevenue,
			)
		}

		if accountBalance.SpendableBalance < tx.Body.GetAmount() {
			return errors.New("transaction amount not enough")
		}
	}
	return nil
}

func (tx *SendMoney) GetAmount() int64 {
	return tx.Body.Amount
}
func (tx *SendMoney) GetSize() uint32 {
	// only amount (int64)
	return uint32(8)
}
