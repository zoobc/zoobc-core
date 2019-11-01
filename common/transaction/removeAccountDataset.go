package transaction

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type RemoveAccountDataset struct {
	Body                *model.RemoveAccountDatasetTransactionBody
	Fee                 int64
	SenderAddress       string
	Height              uint32
	AccountBalanceQuery query.AccountBalanceQueryInterface
	AccountDatasetQuery query.AccountDatasetsQueryInterface
	QueryExecutor       query.ExecutorInterface
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *RemoveAccountDataset) SkipMempoolTransaction(selectedTransactions []*model.Transaction) (bool, error) {
	return false, nil
}

/*
ApplyConfirmed is func that for applying Transaction RemoveAccountDataset type,
*/
func (tx *RemoveAccountDataset) ApplyConfirmed() error {
	var (
		err     error
		dataset *model.AccountDataset
	)

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)

	// Account dataset removed when TimestampStarts same with TimestampExpires
	currentTime := uint64(time.Now().Unix())
	dataset = &model.AccountDataset{
		SetterAccountAddress:    tx.Body.GetSetterAccountAddress(),
		RecipientAccountAddress: tx.Body.GetRecipientAccountAddress(),
		Property:                tx.Body.GetProperty(),
		Value:                   tx.Body.GetValue(),
		TimestampStarts:         currentTime,
		TimestampExpires:        currentTime,
		Height:                  tx.Height,
		Latest:                  true,
	}

	datasetQuery := tx.AccountDatasetQuery.RemoveDataset(dataset)
	queries := append(accountBalanceSenderQ, datasetQuery...)
	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveAccountDataset` type
*/
func (tx *RemoveAccountDataset) ApplyUnconfirmed() error {
	var (
		err error
	)

	// update account sender spendable balance
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *RemoveAccountDataset) UndoApplyUnconfirmed() error {
	var (
		err error
	)

	// update account sender spendable balance
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return nil
}

/*
Validate is func that for validating to Transaction RemoveAccountDataset type
That specs:
	- Check existing Account Dataset
	- Check Spendable Balance sender
*/
func (tx *RemoveAccountDataset) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		accountDataset model.AccountDataset
		err            error
	)

	/*
		Chack existing dataset
		Account Dataset can only delete when account dataset exist
	*/
	datasetQ, datasetArg := tx.AccountDatasetQuery.GetLastDataset(
		tx.Body.GetSetterAccountAddress(),
		tx.Body.GetRecipientAccountAddress(),
		tx.Body.GetProperty(),
	)
	row := tx.QueryExecutor.ExecuteSelectRow(datasetQ, datasetArg...)
	err = tx.AccountDatasetQuery.Scan(&accountDataset, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return blocker.NewBlocker(blocker.ValidationErr, "Remove Account Dataset, Dataset does not exist ")
		}
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	// check account balance sender
	row = tx.QueryExecutor.ExecuteSelectRow(tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress))
	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "RemoveAccountDataset, user balance not enough")
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *RemoveAccountDataset) GetAmount() int64 {
	return tx.Fee
}

// GetSize is size of transaction body
func (tx *RemoveAccountDataset) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *RemoveAccountDataset) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// read body bytes
	buffer := bytes.NewBuffer(txBodyBytes)
	setterAccountAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	setterAccountAddressLength := util.ConvertBytesToUint32(setterAccountAddressLengthBytes)
	setterAccountAddress, err := util.ReadTransactionBytes(buffer, int(setterAccountAddressLength))
	if err != nil {
		return nil, err
	}
	recipientAccountAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	recipientAccountAddressLength := util.ConvertBytesToUint32(recipientAccountAddressLengthBytes)
	recipientAccountAddress, err := util.ReadTransactionBytes(buffer, int(recipientAccountAddressLength))
	if err != nil {
		return nil, err
	}
	propertyLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.DatasetPropertyLength))
	if err != nil {
		return nil, err
	}
	propertyLength := util.ConvertBytesToUint32(propertyLengthBytes)
	property, err := util.ReadTransactionBytes(buffer, int(propertyLength))
	if err != nil {
		return nil, err
	}
	valueLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.DatasetValueLength))
	if err != nil {
		return nil, err
	}
	valueLength := util.ConvertBytesToUint32(valueLengthBytes)
	value, err := util.ReadTransactionBytes(buffer, int(valueLength))
	if err != nil {
		return nil, err
	}
	txBody := &model.RemoveAccountDatasetTransactionBody{
		SetterAccountAddress:    string(setterAccountAddress),
		RecipientAccountAddress: string(recipientAccountAddress),
		Property:                string(property),
		Value:                   string(value),
	}
	return txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *RemoveAccountDataset) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetSetterAccountAddress())))))
	buffer.Write([]byte(tx.Body.GetSetterAccountAddress()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetRecipientAccountAddress())))))
	buffer.Write([]byte(tx.Body.GetRecipientAccountAddress()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetProperty())))))
	buffer.Write([]byte(tx.Body.GetProperty()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetValue())))))
	buffer.Write([]byte(tx.Body.GetValue()))

	return buffer.Bytes()
}

func (tx *RemoveAccountDataset) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_RemoveAccountDatasetTransactionBody{
		RemoveAccountDatasetTransactionBody: tx.Body,
	}
}
