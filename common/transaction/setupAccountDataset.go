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
	SenderAddress       []byte
	RecipientAddress    []byte
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
func (tx *SetupAccountDataset) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
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
		SetterAccountAddress:    tx.SenderAddress,
		RecipientAccountAddress: tx.RecipientAddress,
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
		qry            string
		qryArgs        []interface{}
	)

	// Recipient required while property set as AccountDatasetEscrowApproval
	_, ok := model.AccountDatasetProperty_value[tx.Body.GetProperty()]
	if ok && tx.RecipientAddress == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "RecipientRequired")
	}

	// check existing account_dataset
	qry, qryArgs = tx.AccountDatasetQuery.GetLatestAccountDataset(
		tx.SenderAddress,
		tx.RecipientAddress,
		tx.Body.GetProperty(),
	)
	// NOTE: currently dbTx became true only when calling on push block,
	// this is will make allow to execute all of same tx in mempool if all of them selected
	// TODO: should be using skip mempool to check double same tx in mempool
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, false, qryArgs...)
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
	qry, qryArgs = tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, qryArgs...)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if accountBalance.GetSpendableBalance() < tx.Fee {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"UserBalanceNotEnough",
		)
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SetupAccountDataset) GetAmount() int64 {
	return tx.Fee
}

// GetMinimumFee return minimum fee of transaction
// TODO: need to calculate the minimum fee
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
		err          error
		chunkedBytes []byte
		dataLength   uint32
		txBody       model.SetupAccountDatasetTransactionBody
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
func (tx *SetupAccountDataset) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetProperty())))))
	buffer.Write([]byte(tx.Body.GetProperty()))

	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(tx.Body.GetValue())))))
	buffer.Write([]byte(tx.Body.GetValue()))

	return buffer.Bytes()
}

// GetTransactionBody return transaction body of SetupAccountDataset transactions
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
