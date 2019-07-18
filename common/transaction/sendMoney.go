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
	// todo: undo apply unconfirmed for non-genesis transaction
	var (
		recipientAccountBalance model.AccountBalance
		recipientAccount        model.Account
		senderAccountBalance    model.AccountBalance
		senderAccount           model.Account
		err                     error
	)

	if err := tx.Validate(); err != nil {
		return err
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

	if tx.Height == 0 {
		senderAccountQ, senderAccountArgs := tx.AccountQuery.GetAccountByID(senderAccount.ID)
		senderAccountRows, _ := tx.QueryExecutor.ExecuteSelect(senderAccountQ, senderAccountArgs...)
		if !senderAccountRows.Next() { // genesis account not created yet
			senderAccountBalance = model.AccountBalance{
				AccountID:        senderAccount.ID,
				BlockHeight:      tx.Height,
				SpendableBalance: 0,
				Balance:          0,
				PopRevenue:       0,
				Latest:           true,
			}
			senderAccountInsertQ, senderAccountInsertArgs := tx.AccountQuery.InsertAccount(&senderAccount)
			senderAccountBalanceInsertQ, senderAccountBalanceInsertArgs := tx.AccountBalanceQuery.InsertAccountBalance(&senderAccountBalance)
			_, err = tx.QueryExecutor.ExecuteTransactionStatements(map[*string][]interface{}{
				&senderAccountInsertQ:        senderAccountInsertArgs,
				&senderAccountBalanceInsertQ: senderAccountBalanceInsertArgs,
			})
			if err != nil {
				return err
			}
		}
		_ = senderAccountRows.Close()
		recipientAccountBalance = model.AccountBalance{
			AccountID:        recipientAccount.ID,
			BlockHeight:      tx.Height,
			SpendableBalance: tx.Body.GetAmount(),
			Balance:          tx.Body.GetAmount(),
			PopRevenue:       0,
			Latest:           true,
		}
		accountQ, accountQArgs := tx.AccountQuery.InsertAccount(&recipientAccount)
		accountBalanceQ, accountBalanceArgs := tx.AccountBalanceQuery.InsertAccountBalance(&recipientAccountBalance)
		// update sender
		accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountBalance(
			-tx.Body.GetAmount(),
			map[string]interface{}{
				"account_id": senderAccount.ID,
			},
		)
		_, err = tx.QueryExecutor.ExecuteTransactionStatements(map[*string][]interface{}{
			&accountQ:              accountQArgs,
			&accountBalanceQ:       accountBalanceArgs,
			&accountBalanceSenderQ: accountBalanceSenderQArgs,
		})
		if err != nil {
			return err
		}
	} else {
		// update recipient
		accountBalanceRecipientQ, accountBalanceRecipientQArgs := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Body.Amount,
			map[string]interface{}{
				"account_id": recipientAccount.ID,
			},
		)
		// update sender
		accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountBalance(
			-tx.Body.Amount,
			map[string]interface{}{
				"account_id": senderAccount.ID,
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

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-tx.Body.Amount,
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	_, err = tx.QueryExecutor.ExecuteTransactionStatements(map[*string][]interface{}{
		&accountBalanceSenderQ: accountBalanceSenderQArgs,
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
				&accountBalance.Latest,
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

func (*SendMoney) GetSize() uint32 {
	// only amount (int64)
	return 8
}
