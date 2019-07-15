package transaction

import (
	"database/sql"
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
ApplyConfirmed is func that for applying Transaction SendMoney type
update `AccountBalance.Balance` for affected `Account.ID`
if account not exists would be created new.
return error while query is failure
*/
func (tx *SendMoney) ApplyConfirmed() error {

	var (
		err            error
		rows           *sql.Rows
		account        model.Account
		accountBalance model.AccountBalance
	)

	accountQ, accountQArgs := tx.AccountQuery.GetAccountByID(util.CreateAccountIDFromAddress(
		tx.RecipientAccountType,
		tx.RecipientAddress,
	))

	rows, err = tx.QueryExecutor.ExecuteSelect(accountQ, accountQArgs)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {

		err = rows.Scan(
			&account.ID,
			&account.Address,
			&account.AccountType,
		)
		if err != nil {
			return err
		}

		accountBalanceQ, accountBalanceQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
			map[string]interface{}{
				"balance": fmt.Sprintf("balance - %d", tx.Body.GetAmount()),
			},
			map[string]interface{}{
				"account_id": account.ID,
			},
		)

		_, err = tx.QueryExecutor.ExecuteStatement(accountBalanceQ, accountBalanceQArgs)
		if err != nil {
			return err
		}
	} else {
		account = model.Account{
			ID:          util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
			AccountType: tx.RecipientAccountType,
			Address:     tx.RecipientAddress,
		}
		accountQ, accountQArgs = tx.AccountQuery.InsertAccount(&account)
		_, err = tx.QueryExecutor.ExecuteStatement(accountQ, accountQArgs)
		if err != nil {
			return err
		}
		accountBalance = model.AccountBalance{
			AccountID:        account.ID,
			BlockHeight:      tx.Height,
			SpendableBalance: tx.Body.GetAmount(),
			Balance:          tx.Body.GetAmount(),
			PopRevenue:       0,
		}
		accountBalanceQ, accountBalanceArgs := tx.AccountBalanceQuery.InsertAccountBalance(&accountBalance)
		_, err = tx.QueryExecutor.ExecuteStatement(accountBalanceQ, accountBalanceArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction SendMoney type
update `AccountBalance.SpendableBalance` for affected `Account.ID`
if account not exists would be created new.
return error while query is failure
*/
func (tx *SendMoney) ApplyUnconfirmed() error {

	var (
		err            error
		rows           *sql.Rows
		account        model.Account
		accountBalance model.AccountBalance
		accountQ       string
		accountQArgs   interface{}
	)

	accountQ, accountQArgs = tx.AccountQuery.GetAccountByID(util.CreateAccountIDFromAddress(
		tx.RecipientAccountType,
		tx.RecipientAddress,
	))
	rows, err = tx.QueryExecutor.ExecuteSelect(accountQ, accountQArgs)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&account.ID, &account.AccountType, &account.Address)
		if err != nil {
			return err
		}
		accountBalanceQ, accountBalanceQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
			map[string]interface{}{
				"spendable_balance": fmt.Sprintf("spendable_balance - %d", tx.Body.GetAmount()),
			},
			map[string]interface{}{
				"account_id": account.ID,
			},
		)

		_, err = tx.QueryExecutor.ExecuteStatement(accountBalanceQ, accountBalanceQArgs)
		if err != nil {
			return err
		}

	} else {
		account = model.Account{
			ID:          util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
			AccountType: tx.RecipientAccountType,
			Address:     tx.RecipientAddress,
		}
		accountQ, accountQArgs = tx.AccountQuery.InsertAccount(&account)
		_, err = tx.QueryExecutor.ExecuteStatement(accountQ, accountQArgs)
		if err != nil {
			return err
		}
		accountBalance = model.AccountBalance{
			AccountID:        account.ID,
			BlockHeight:      tx.Height,
			SpendableBalance: tx.Body.GetAmount(),
			Balance:          tx.Body.GetAmount(),
			PopRevenue:       0,
		}
		accountBalanceQ, accountBalanceArgs := tx.AccountBalanceQuery.InsertAccountBalance(&accountBalance)
		_, err = tx.QueryExecutor.ExecuteStatement(accountBalanceQ, accountBalanceArgs)
		if err != nil {
			return err
		}

	}

	return nil
}

// Validate is func that for validating to Transaction SendMoney type
func (tx *SendMoney) Validate() error {

	var (
		rows           *sql.Rows
		err            error
		accountBalance model.AccountBalance
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.Height != 0 {
		if (tx.RecipientAddress == "") || (tx.RecipientAccountType <= 0) {
			return errors.New("transaction must have a valid recipient account id")
		}
		if (tx.SenderAddress == "") || (tx.SenderAccountType <= 0) {
			return errors.New("transaction must have a valid sender account id")
		}

		rows, err = tx.QueryExecutor.ExecuteSelect(
			tx.AccountBalanceQuery.GetAccountBalanceByAccountID(),
			util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			_ = rows.Scan(
				&accountBalance.AccountID,
				&accountBalance.BlockHeight,
				&accountBalance.SpendableBalance,
				&accountBalance.Balance,
				&accountBalance.PopRevenue,
			)
		} else {
			return errors.New("account not exists")
		}

		if accountBalance.SpendableBalance < tx.Body.GetAmount() {
			return errors.New("transaction amount not enough")
		}
	}
	return nil
}
