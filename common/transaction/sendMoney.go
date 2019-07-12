package transaction

import (
	"database/sql"
	"errors"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
)

type SendMoney struct {
	Body                *model.SendMoneyTransactionBody
	SenderAccountID     []byte
	RecipientAccountID  []byte
	Height              uint32
	AccountBalanceQuery query.AccountBalanceInt
	AccountQuery        query.AccountQueryInterface
	QueryExecutor       query.ExecutorInterface
}

func (tx *SendMoney) Apply() error {
	return nil
}

func (tx *SendMoney) Unconfirmed() error {
	var (
		rows    *sql.Rows
		err     error
		account *model.Account
	)
	accountQ, accountQArgs := tx.AccountQuery.GetAccountByID(tx.RecipientAccountID)
	rows, err = tx.QueryExecutor.ExecuteSelect(accountQ, accountQArgs)
	if err != nil {
		return err
	}

	if rows.Next() {
		err = rows.Scan(&account.ID, &account.AccountType, &account.Address)
		if err != nil {
			return err
		}
	} else {
		accountQ = tx.AccountQuery.InsertAccount()
		_, err = tx.QueryExecutor.ExecuteStatement(accountQ, tx.AccountQuery.ExtractModel())
	}

	//accountBalanceQ, accountBalanceQArgs := tx.AccountBalanceQuery.UpdateAccountBalance(
	//	map[string]interface{}{
	//		"spendable_balance": tx.Body.GetAmount(),
	//	},
	//	map[string]interface{}{
	//		"account_id": tx.RecipientAccountID,
	//	},
	//)

	return nil
}
func (tx *SendMoney) Validate() error {

	var (
		rows           *sql.Rows
		err            error
		accountBalance *model.AccountBalance
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.Height != 0 {
		if tx.RecipientAccountID == nil {
			return errors.New("transaction must have a valid recipient account id")
		}
		if tx.SenderAccountID == nil {
			return errors.New("transaction must hav a valid sender account id")
		}

		rows, err = tx.QueryExecutor.ExecuteSelect(tx.AccountBalanceQuery.GetAccountBalanceByAccountID())
		if err != nil {
			return err
		}
		err = rows.Scan(
			&accountBalance.AccountID,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
		)
		if err != nil {
			return err
		}

		if accountBalance.SpendableBalance < tx.Body.GetAmount() {
			return errors.New("transaction amount not enough")
		}
	}
	return nil
}
