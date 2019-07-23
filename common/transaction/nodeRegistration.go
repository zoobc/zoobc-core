package transaction

import (
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
)

// NodeRegistration Implement service layer for (new) node registration's transaction
type NodeRegistration struct {
	Body                  *model.NodeRegistrationTransactionBody
	Fee                   int64
	SenderAddress         string
	SenderAccountType     uint32
	Height                uint32
	AccountBalanceQuery   query.AccountBalanceQueryInterface
	AccountQuery          query.AccountQueryInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	QueryExecutor         query.ExecutorInterface
}

func (tx *NodeRegistration) ApplyConfirmed() error {
	return nil
}

/*
ApplyConfirmed is func that for applying Transaction NodeRegistration type,

__If Genesis__:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.

__If Not Genesis__:
	- perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
	`sender.balance` = current - amount
*/
// func (tx *NodeRegistration) ApplyConfirmed() error {
// 	// todo: undo apply unconfirmed for non-genesis transaction
// 	var (
// 		recipientAccountBalance model.AccountBalance
// 		recipientAccount        model.Account
// 		senderAccountBalance    model.AccountBalance
// 		senderAccount           model.Account
// 		err                     error
// 	)

// 	if err := tx.Validate(); err != nil {
// 		return err
// 	}

// 	recipientAccount = model.Account{
// 		ID:          util.CreateAccountIDFromAddress(tx.RecipientAccountType, tx.RecipientAddress),
// 		AccountType: tx.RecipientAccountType,
// 		Address:     tx.RecipientAddress,
// 	}
// 	senderAccount = model.Account{
// 		ID:          util.CreateAccountIDFromAddress(tx.SenderAccountType, tx.SenderAddress),
// 		AccountType: tx.SenderAccountType,
// 		Address:     tx.SenderAddress,
// 	}

// 	if tx.Height == 0 {
// 		senderAccountQ, senderAccountArgs := tx.AccountQuery.GetAccountByID(senderAccount.ID)
// 		senderAccountRows, _ := tx.QueryExecutor.ExecuteSelect(senderAccountQ, senderAccountArgs...)
// 		if !senderAccountRows.Next() { // genesis account not created yet
// 			senderAccountBalance = model.AccountBalance{
// 				AccountID:        senderAccount.ID,
// 				BlockHeight:      tx.Height,
// 				SpendableBalance: 0,
// 				Balance:          0,
// 				PopRevenue:       0,
// 				Latest:           true,
// 			}
// 			senderAccountInsertQ, senderAccountInsertArgs := tx.AccountQuery.InsertAccount(&senderAccount)
// 			senderAccountBalanceInsertQ, senderAccountBalanceInsertArgs := tx.AccountBalanceQuery.InsertAccountBalance(&senderAccountBalance)
// 			_, err = tx.QueryExecutor.ExecuteTransactionStatements([][]interface{}{
// 				append([]interface{}{senderAccountInsertQ}, senderAccountInsertArgs...),
// 				append([]interface{}{senderAccountBalanceInsertQ}, senderAccountBalanceInsertArgs...),
// 			})
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		_ = senderAccountRows.Close()
// 		recipientAccountBalance = model.AccountBalance{
// 			AccountID:        recipientAccount.ID,
// 			BlockHeight:      tx.Height,
// 			SpendableBalance: tx.Body.GetAmount(),
// 			Balance:          tx.Body.GetAmount(),
// 			PopRevenue:       0,
// 			Latest:           true,
// 		}
// 		accountQ, accountQArgs := tx.AccountQuery.InsertAccount(&recipientAccount)
// 		accountBalanceQ, accountBalanceArgs := tx.AccountBalanceQuery.InsertAccountBalance(&recipientAccountBalance)
// 		// update sender
// 		accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountBalance(
// 			-tx.Body.GetAmount(),
// 			map[string]interface{}{
// 				"account_id": senderAccount.ID,
// 			},
// 		)
// 		_, err = tx.QueryExecutor.ExecuteTransactionStatements([][]interface{}{
// 			append([]interface{}{accountQ}, accountQArgs...),
// 			append([]interface{}{accountBalanceQ}, accountBalanceArgs...),
// 			append([]interface{}{accountBalanceSenderQ}, accountBalanceSenderQArgs...),
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		// update recipient
// 		accountBalanceRecipientQ, accountBalanceRecipientQArgs := tx.AccountBalanceQuery.AddAccountBalance(
// 			tx.Body.Amount,
// 			map[string]interface{}{
// 				"account_id": recipientAccount.ID,
// 			},
// 		)
// 		// update sender
// 		accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountBalance(
// 			-tx.Body.Amount,
// 			map[string]interface{}{
// 				"account_id": senderAccount.ID,
// 			},
// 		)
// 		_, err = tx.QueryExecutor.ExecuteTransactionStatements([][]interface{}{
// 			append([]interface{}{accountBalanceSenderQ}, accountBalanceSenderQArgs...),
// 			append([]interface{}{accountBalanceRecipientQ}, accountBalanceRecipientQArgs...),
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `NodeRegistration` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *NodeRegistration) ApplyUnconfirmed() error {

	var (
		err error
	)

	if err := tx.Validate(); err != nil {
		return err
	}

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-tx.Fee,
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	_, err = tx.QueryExecutor.ExecuteTransactionStatements([][]interface{}{
		{append([]interface{}{accountBalanceSenderQ}, accountBalanceSenderQArgs...)},
	})
	if err != nil {
		return err
	}

	return nil
}

// Validate validate node registration transaction and tx body
func (tx *NodeRegistration) Validate() error {
	return nil
}

func (tx *NodeRegistration) GetAmount() int64 {
	return 0
}

func (*NodeRegistration) GetSize() uint32 {
	nodePublicKey := 32
	accountType := 2
	//TODO: this is valid for account type = 0
	accountAddress := 44
	registrationHeight := 4
	// Note: as address is a variable string, by convention, the client should pass a string long 100 bytes, then we parse it internally
	nodeAddress := 100
	lockedBalance := 8
	return uint32(nodePublicKey + accountType + accountAddress + registrationHeight + nodeAddress + lockedBalance)
}
