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

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-tx.Fee,
		map[string]interface{}{
			"account_id": util.CreateAccountIDFromAddress(
				tx.SenderAccountType,
				tx.SenderAddress,
			),
		},
	)
	_, err = tx.QueryExecutor.ExecuteStatement(accountBalanceSenderQ, accountBalanceSenderQArgs...)
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
