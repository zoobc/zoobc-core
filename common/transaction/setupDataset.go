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

type SetupDataset struct {
	Body                *model.SetupDatasetTransactionBody
	Fee                 int64
	SenderAddress       string
	Height              uint32
	AccountBalanceQuery query.AccountBalanceQueryInterface
	DatasetQuery        query.DatasetsQueryInterface
	QueryExecutor       query.ExecutorInterface
}

/*
ApplyConfirmed is func that for applying Transaction SetupDataset type,
*/
func (tx *SetupDataset) ApplyConfirmed() error {
	var (
		err     error
		dataset *model.Dataset
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
	dataset = &model.Dataset{
		AccountSetter:    tx.Body.GetAccountSetter(),
		AccountRecipient: tx.Body.GetAccountRecipient(),
		Property:         tx.Body.GetProperty(),
		Value:            tx.Body.GetValue(),
		TimestampStarts:  currentTime,
		TimestampExpires: currentTime + tx.Body.GetMuchTime(),
		Height:           tx.Height,
		Latest:           true,
	}

	datasetQuery := tx.DatasetQuery.AddDataset(dataset)
	queries := append(accountBalanceSenderQ, datasetQuery...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `SetupDataset` type
*/
func (tx *SetupDataset) ApplyUnconfirmed() error {

	var (
		err error
	)

	if err := tx.Validate(); err != nil {
		return err
	}

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
func (tx *SetupDataset) UndoApplyUnconfirmed() error {
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
Validate is func that for validating to Transaction SetupDataset type
That specs:
	- Checking the expiration time
	- Checking Spendable Balance sender
*/
func (tx *SetupDataset) Validate() error {

	var (
		accountBalance model.AccountBalance
	)

	if tx.Body.GetMuchTime() == 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "SetupDataset, starts time is not allowed same with expiration time")
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
		return blocker.NewBlocker(blocker.ValidationErr, "SetupDataset, user balance not enough")
	}

	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SetupDataset) GetAmount() int64 {
	// TODO: transaction fee + (expiration time fee)
	return tx.Fee
}

// GetSize is size of transaction body
func (tx *SetupDataset) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (*SetupDataset) ParseBodyBytes(txBodyBytes []byte) model.TransactionBodyInterface {
	buffer := bytes.NewBuffer(txBodyBytes)
	accountSetterLength := util.ConvertBytesToUint32(buffer.Next(int(constant.AccountAddressLength)))
	accountSetter := buffer.Next(int(accountSetterLength))

	accountRecipientLength := util.ConvertBytesToUint32(buffer.Next(int(constant.AccountAddressLength)))
	accountRecipient := buffer.Next(int(accountRecipientLength))

	propertyLength := util.ConvertBytesToUint32(buffer.Next(int(constant.DatasetPropertyLength)))
	property := buffer.Next(int(propertyLength))

	valueLength := util.ConvertBytesToUint32(buffer.Next(int(constant.DatasetValueLength)))
	value := buffer.Next(int(valueLength))

	muchTime := util.ConvertBytesToUint64(buffer.Next(int(constant.Timestamp)))

	txBody := &model.SetupDatasetTransactionBody{
		AccountSetter:    string(accountSetter),
		AccountRecipient: string(accountRecipient),
		Property:         string(property),
		Value:            string(value),
		MuchTime:         muchTime,
	}
	return txBody
}

// GetBodyBytes translate tx body to bytes representation
func (tx *SetupDataset) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetAccountSetter())))))
	buffer.Write([]byte(tx.Body.GetAccountSetter()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetAccountRecipient())))))
	buffer.Write([]byte(tx.Body.GetAccountRecipient()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetProperty())))))
	buffer.Write([]byte(tx.Body.GetProperty()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetValue())))))
	buffer.Write([]byte(tx.Body.GetValue()))

	buffer.Write(util.ConvertUint64ToBytes(tx.Body.GetMuchTime()))

	return buffer.Bytes()
}
