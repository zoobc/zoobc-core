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
	RecipientAddress    string
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
func (tx *RemoveAccountDataset) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	blockTimestamp int64,
) (bool, error) {
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
		SetterAccountAddress:    tx.SenderAddress,
		RecipientAccountAddress: tx.RecipientAddress,
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
		qry            string
		qryArgs        []interface{}
	)

	/*
		Check existing dataset
		Account Dataset can only delete when account dataset exist
	*/
	qry, qryArgs = tx.AccountDatasetQuery.GetLatestAccountDataset(
		tx.SenderAddress,
		tx.RecipientAddress,
		tx.Body.GetProperty(),
	)

	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, qryArgs...)
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
	qry, qryArgs = tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, qryArgs...)
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

// GetMinimumFee return minimum fee of transaction
// TODO: need to calculate the minimum fee
func (*RemoveAccountDataset) GetMinimumFee() (int64, error) {
	return 0, nil
}

// GetSize is size of transaction body
func (tx *RemoveAccountDataset) GetSize() uint32 {
	return uint32(len(tx.GetBodyBytes()))
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *RemoveAccountDataset) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	var (
		err          error
		chunkedBytes []byte
		dataLength   uint32
		txBody       model.RemoveAccountDatasetTransactionBody
		buffer       = bytes.NewBuffer(txBodyBytes)
	)
	// get length of property dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.DatasetPropertyLength))
	if err != nil {
		return nil, err
	}
	dataLength = util.ConvertBytesToUint32(chunkedBytes)
	// get property of dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
	if err != nil {
		return nil, err
	}
	txBody.Property = string(chunkedBytes)
	// get length of value property dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.DatasetValueLength))
	if err != nil {
		return nil, err
	}
	dataLength = util.ConvertBytesToUint32(chunkedBytes)
	// get value property of dataset
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(dataLength))
	if err != nil {
		return nil, err
	}
	txBody.Value = string(chunkedBytes)
	return &txBody, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *RemoveAccountDataset) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetProperty())))))
	buffer.Write([]byte(tx.Body.GetProperty()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetValue())))))
	buffer.Write([]byte(tx.Body.GetValue()))

	return buffer.Bytes()
}

// GetTransactionBody return transaction body of RemoveAccountDataset transactions
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
