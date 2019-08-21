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
	if tx.Height > 0 {
		err = tx.UndoApplyUnconfirmed()
		if err != nil {
			return err
		}
	}

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
func (tx *SetupAccountDataset) Validate() error {

	var (
		accountBalance model.AccountBalance
	)

	if tx.Body.GetMuchTime() == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "SetupAccountDataset, starts time is not allowed same with expiration time")
	}

	// check balance
	senderQ, senderArg := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	rows, err := tx.QueryExecutor.ExecuteSelect(senderQ, senderArg)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	} else if rows.Next() {
		_ = rows.Scan(
			&accountBalance.AccountAddress,
			&accountBalance.BlockHeight,
			&accountBalance.SpendableBalance,
			&accountBalance.Balance,
			&accountBalance.PopRevenue,
			&accountBalance.Latest,
		)
	}
	defer rows.Close()
	// TODO: transaction fee + (expiration time fee)
	if accountBalance.SpendableBalance < tx.Fee {
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
func (*SetupAccountDataset) ParseBodyBytes(txBodyBytes []byte) model.TransactionBodyInterface {
	buffer := bytes.NewBuffer(txBodyBytes)
	setterAccountAddressLength := util.ConvertBytesToUint32(buffer.Next(int(constant.AccountAddressLength)))
	setterAccountAddress := buffer.Next(int(setterAccountAddressLength))

	recipientAccountAddressLength := util.ConvertBytesToUint32(buffer.Next(int(constant.AccountAddressLength)))
	recipientAccountAddress := buffer.Next(int(recipientAccountAddressLength))

	propertyLength := util.ConvertBytesToUint32(buffer.Next(int(constant.DatasetPropertyLength)))
	property := buffer.Next(int(propertyLength))

	valueLength := util.ConvertBytesToUint32(buffer.Next(int(constant.DatasetValueLength)))
	value := buffer.Next(int(valueLength))

	muchTime := util.ConvertBytesToUint64(buffer.Next(int(constant.Timestamp)))

	txBody := &model.SetupAccountDatasetTransactionBody{
		SetterAccountAddress:    string(setterAccountAddress),
		RecipientAccountAddress: string(recipientAccountAddress),
		Property:                string(property),
		Value:                   string(value),
		MuchTime:                muchTime,
	}
	return txBody
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
