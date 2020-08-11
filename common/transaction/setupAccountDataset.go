package transaction

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// SetupAccountDataset fields that's needed
type SetupAccountDataset struct {
	ID                  int64
	Fee                 int64
	SenderAddress       string
	Height              uint32
	Body                *model.SetupAccountDatasetTransactionBody
	Escrow              *model.Escrow
	AccountBalanceQuery query.AccountBalanceQueryInterface
	AccountDatasetQuery query.AccountDatasetQueryInterface
	QueryExecutor       query.ExecutorInterface
	AccountLedgerQuery  query.AccountLedgerQueryInterface
	EscrowQuery         query.EscrowTransactionQueryInterface
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *SetupAccountDataset) SkipMempoolTransaction([]*model.Transaction) (bool, error) {
	return false, nil
}

/*
ApplyConfirmed is func that for applying Transaction SetupAccountDataset type,
*/
func (tx *SetupAccountDataset) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err     error
		dataset *model.AccountDataset
		queries [][]interface{}
	)

	// update sender balance by reducing his spendable balance of the tx fee
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// This is Default mode, Dataset will be active as soon as block creation
	dataset = &model.AccountDataset{
		SetterAccountAddress:    tx.Body.GetSetterAccountAddress(),
		RecipientAccountAddress: tx.Body.GetRecipientAccountAddress(),
		Property:                tx.Body.GetProperty(),
		Value:                   tx.Body.GetValue(),
		Height:                  tx.Height,
		IsActive:                true,
		Latest:                  true,
	}

	accDatasetQ := tx.AccountDatasetQuery.InsertAccountDataset(dataset)
	queries = append(queries, accDatasetQ...)

	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventSetupAccountDatasetTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	queries = append(queries, append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...))

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
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
		accountDataset model.AccountDataset
		row            *sql.Row
		err            error
	)

	// Recipient required while property set as AccountDatasetEscrowApproval
	_, ok := model.AccountDatasetProperty_value[tx.Body.GetProperty()]
	if ok && tx.Body.GetRecipientAccountAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "RecipientRequired")
	}

	// check existing account_dataset
	accDatasetQ, accDatasetArgs := tx.AccountDatasetQuery.GetLatestAccountDataset(
		tx.Body.GetSetterAccountAddress(),
		tx.Body.GetRecipientAccountAddress(),
		tx.Body.GetProperty(),
	)
	// NOTE: currently dbTx became true only when calling on push block,
	// this is will make allow to execute all of same tx in mempool if all of them selected
	// TODO: should be using skip mempool to check double same tx in mempool
	row, err = tx.QueryExecutor.ExecuteSelectRow(accDatasetQ, false, accDatasetArgs...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.AccountDatasetQuery.Scan(&accountDataset, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
	}
	// false if err in above is sql.ErrNoRows || nil
	if accountDataset.GetIsActive() {
		return blocker.NewBlocker(blocker.ValidationErr, "DatasetAlreadyExists")
	}
	// check account balance sender
	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(blocker.ValidationErr, "SetupAccountDataset, user balance not enough")
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SetupAccountDataset) GetAmount() int64 {
	return tx.Fee
}

func (*SetupAccountDataset) GetMinimumFee() (int64, error) {
	return 0, nil
}

// GetSize is size of transaction body
func (tx *SetupAccountDataset) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *SetupAccountDataset) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		buffer = bytes.NewBuffer(txBodyBytes)
		txBody model.SetupAccountDatasetTransactionBody
	)

	setterAccountAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	setterAccountAddressLength := util.ConvertBytesToUint32(setterAccountAddressLengthBytes)
	setterAccountAddress, err := util.ReadTransactionBytes(buffer, int(setterAccountAddressLength))
	if err != nil {
		return nil, err
	}
	txBody.SetterAccountAddress = string(setterAccountAddress)

	recipientAccountAddressLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	recipientAccountAddressLength := util.ConvertBytesToUint32(recipientAccountAddressLengthBytes)
	recipientAccountAddress, err := util.ReadTransactionBytes(buffer, int(recipientAccountAddressLength))
	if err != nil {
		return nil, err
	}
	txBody.RecipientAccountAddress = string(recipientAccountAddress)

	propertyLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.DatasetPropertyLength))
	if err != nil {
		return nil, err
	}
	propertyLength := util.ConvertBytesToUint32(propertyLengthBytes)
	property, err := util.ReadTransactionBytes(buffer, int(propertyLength))
	if err != nil {
		return nil, err
	}
	txBody.Property = string(property)

	valueLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.DatasetValueLength))
	if err != nil {
		return nil, err
	}
	valueLength := util.ConvertBytesToUint32(valueLengthBytes)
	value, err := util.ReadTransactionBytes(buffer, int(valueLength))
	if err != nil {
		return nil, err
	}
	txBody.Value = string(value)

	return &txBody, nil
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

	return buffer.Bytes()
}

func (tx *SetupAccountDataset) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_SetupAccountDatasetTransactionBody{
		SetupAccountDatasetTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *SetupAccountDataset) Escrowable() (EscrowTypeAction, bool) {

	return nil, false
}

/*
EscrowValidate is func that for validating to Transaction SetupAccountDataset type
That specs:
	- Checking the expiration time
	- Checking Spendable Balance sender
*/
func (tx *SetupAccountDataset) EscrowValidate() error {

	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `SetupAccountDataset` type
*/
func (tx *SetupAccountDataset) EscrowApplyUnconfirmed() error {

	// update account sender spendable balance
	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *SetupAccountDataset) EscrowUndoApplyUnconfirmed() error {

	return nil
}

/*
EscrowApplyConfirmed is func that for applying Transaction SetupAccountDataset type,
*/
func (tx *SetupAccountDataset) EscrowApplyConfirmed() error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *SetupAccountDataset) EscrowApproval(int64) error {

	return nil
}
