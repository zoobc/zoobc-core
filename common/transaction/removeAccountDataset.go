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

// RemoveAccountDataset has fields that's needed
type RemoveAccountDataset struct {
	ID                  int64
	Fee                 int64
	SenderAddress       string
	Height              uint32
	Body                *model.RemoveAccountDatasetTransactionBody
	Escrow              *model.Escrow
	AccountBalanceQuery query.AccountBalanceQueryInterface
	AccountDatasetQuery query.AccountDatasetQueryInterface
	QueryExecutor       query.ExecutorInterface
	AccountLedgerQuery  query.AccountLedgerQueryInterface
	EscrowQuery         query.EscrowTransactionQueryInterface
}

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *RemoveAccountDataset) SkipMempoolTransaction([]*model.Transaction) (bool, error) {
	return false, nil
}

/*
ApplyConfirmed is func that for applying Transaction RemoveAccountDataset type,
*/
func (tx *RemoveAccountDataset) ApplyConfirmed(blockTimestamp int64) error {
	var (
		err     error
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

	// Account dataset removed, need to set IsActive false
	datasetQ := tx.AccountDatasetQuery.RemoveAccountDataset(&model.AccountDataset{
		SetterAccountAddress:    tx.Body.GetSetterAccountAddress(),
		RecipientAccountAddress: tx.Body.GetRecipientAccountAddress(),
		Property:                tx.Body.GetProperty(),
		Value:                   tx.Body.GetValue(),
		Height:                  tx.Height,
		Latest:                  true,
		IsActive:                false,
	})
	queries = append(accountBalanceSenderQ, datasetQ...)
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -tx.Fee,
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventRemoveNodeRegistrationTransaction,
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
		row            *sql.Row
	)

	/*
		Check existing dataset
		Account Dataset can only delete when account dataset exist
	*/
	datasetQ, datasetArgs := tx.AccountDatasetQuery.GetLatestAccountDataset(
		tx.Body.GetSetterAccountAddress(),
		tx.Body.GetRecipientAccountAddress(),
		tx.Body.GetProperty(),
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(datasetQ, dbTx, datasetArgs...)
	if err != nil {
		return err
	}
	err = tx.AccountDatasetQuery.Scan(&accountDataset, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		return blocker.NewBlocker(blocker.ValidationErr, "AccountDatasetNotExists")
	}
	// !false if err in above is sql.ErrNoRows || nil
	if !accountDataset.GetIsActive() {
		return blocker.NewBlocker(blocker.ValidationErr, "AccountDatasetAlreadyRemoved")
	}

	// check account balance sender
	qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
	if err != nil {
		return err
	}
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

func (*RemoveAccountDataset) GetMinimumFee() (int64, error) {
	return 0, nil
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

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *RemoveAccountDataset) Escrowable() (EscrowTypeAction, bool) {

	return nil, false
}

/*
EscrowValidate is func that for validating to Transaction RemoveAccountDataset type
That specs:
	- Check existing Account Dataset
	- Check Spendable Balance sender
*/
func (tx *RemoveAccountDataset) EscrowValidate(bool) error {

	return nil
}

/*
EscrowApplyUnconfirmed is func that for applying to unconfirmed Transaction `RemoveAccountDataset` type
*/
func (tx *RemoveAccountDataset) EscrowApplyUnconfirmed() error {

	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *RemoveAccountDataset) EscrowUndoApplyUnconfirmed() error {

	return nil
}

/*
EscrowApplyConfirmed is func that for applying Transaction RemoveAccountDataset type,
*/
func (tx *RemoveAccountDataset) EscrowApplyConfirmed(int64) error {

	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *RemoveAccountDataset) EscrowApproval(int64, *model.ApprovalEscrowTransactionBody) error {

	return nil
}
