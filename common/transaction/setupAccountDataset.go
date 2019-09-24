package transaction

import (
	"bytes"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type SetupAccountDataset struct {
	Body                *model.SetupAccountDatasetTransactionBody
	Fee                 int64
	SenderAddress       string
	Height              uint32
	AccountBalanceQuery query.AccountBalanceQueryInterface
	AccountDatasetQuery query.AccountDatasetsQueryInterface
	QueryExecutor       query.ExecutorInterface
}

/*
ApplyConfirmed is func that for applying Transaction SetupAccountDataset type,
*/
func (tx *SetupAccountDataset) ApplyConfirmed() error {
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

	// This is Default mode, Dataset will be active as soon as block creation
	currentTime := uint64(time.Now().Unix())
	dataset = &model.AccountDataset{
		SetterAccountAddress:    tx.Body.GetSetterAccountAddress(),
		RecipientAccountAddress: tx.Body.GetRecipientAccountAddress(),
		Property:                tx.Body.GetProperty(),
		Value:                   tx.Body.GetValue(),
		TimestampStarts:         currentTime,
		TimestampExpires:        currentTime + tx.Body.GetMuchTime(),
		Height:                  tx.Height,
		Latest:                  true,
	}

	datasetQuery := tx.AccountDatasetQuery.AddDataset(dataset)
	queries := append(accountBalanceSenderQ, datasetQuery...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `SetupAccountDataset` type
*/
func (tx *SetupAccountDataset) ApplyUnconfirmed() error {

	var (
		err error
	)

	// update account sender spendable balance
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		// TODO: transaction fee + (expiration time fee)
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
func (tx *SetupAccountDataset) UndoApplyUnconfirmed() error {
	var (
		err error
	)

	// update account sender spendable balance
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		// TODO: transaction fee + (expiration time fee)
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
Validate is func that for validating to Transaction SetupAccountDataset type
That specs:
	- Checking the expiration time
	- Checking Spendable Balance sender
*/
func (tx *SetupAccountDataset) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
	)
	if tx.Body.GetMuchTime() == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "SetupAccountDataset, starts time is not allowed same with expiration time")
	}
	// check account balance sender
	senderQ, senderArg := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row := tx.QueryExecutor.ExecuteSelectRow(senderQ, senderArg)
	err := tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// TODO: transaction fee + (expiration time fee)
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "SetupAccountDataset, user balance not enough")
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SetupAccountDataset) GetAmount() int64 {
	// TODO: transaction fee + (expiration time fee)
	return tx.Fee
}

// GetSize is size of transaction body
func (tx *SetupAccountDataset) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *SetupAccountDataset) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
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
	muchTimeBytes, err := util.ReadTransactionBytes(buffer, int(constant.Timestamp))
	if err != nil {
		return nil, err
	}
	muchTime := util.ConvertBytesToUint64(muchTimeBytes)
	txBody := &model.SetupAccountDatasetTransactionBody{
		SetterAccountAddress:    string(setterAccountAddress),
		RecipientAccountAddress: string(recipientAccountAddress),
		Property:                string(property),
		Value:                   string(value),
		MuchTime:                muchTime,
	}
	return txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *SetupAccountDataset) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetSetterAccountAddress())))))
	buffer.Write([]byte(tx.Body.GetSetterAccountAddress()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetRecipientAccountAddress())))))
	buffer.Write([]byte(tx.Body.GetRecipientAccountAddress()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetProperty())))))
	buffer.Write([]byte(tx.Body.GetProperty()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetValue())))))
	buffer.Write([]byte(tx.Body.GetValue()))

	buffer.Write(util.ConvertUint64ToBytes(tx.Body.GetMuchTime()))

	return buffer.Bytes()
}
