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
	Heigh               uint32
	AccountBalanceQuery query.AccountBalanceQuery
	QueryExecutor       query.ExecutorInterface
}

func (tx *SendMoney) Apply() error {
	return nil
}

func (tx *SendMoney) Unconfirmed() error {
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
	if tx.Heigh != 0 {
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
			&accountBalance.ID,
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
