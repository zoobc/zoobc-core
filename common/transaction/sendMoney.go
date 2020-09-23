package transaction

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// SendMoney is Transaction Type that implemented TypeAction
	SendMoney struct {
		ID                  int64
		Fee                 int64
		SenderAddress       string
		RecipientAddress    string
		Height              uint32
		Body                *model.SendMoneyTransactionBody
		QueryExecutor       query.ExecutorInterface
		Escrow              *model.Escrow
		AccountBalanceQuery query.AccountBalanceQueryInterface
		AccountLedgerQuery  query.AccountLedgerQueryInterface
		EscrowQuery         query.EscrowTransactionQueryInterface
		BlockQuery          query.BlockQueryInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
		NormalFee           fee.FeeModelInterface
		EscrowFee           fee.FeeModelInterface
	}
)

// SkipMempoolTransaction this tx type has no mempool filter
func (tx *SendMoney) SkipMempoolTransaction(
	selectedTransactions []*model.Transaction,
	newBlockTimestamp int64,
	newBlockHeight uint32,
) (bool, error) {
	return false, nil
}

/*
ApplyConfirmed func that for applying Transaction SendMoney type.
If Genesis:
		- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
If Not Genesis:
		- perhaps sender and recipient is exists, so update `account_balance`, `recipient.balance` = current + amount and
		`sender.balance` = current - amount
*/
func (tx *SendMoney) ApplyConfirmed(blockTimestamp int64) error {
	var (
		queries [][]interface{}
		err     error
	)

	// insert / update recipient
	accountBalanceRecipientQ := tx.AccountBalanceQuery.AddAccountBalance(
		tx.Body.Amount,
		map[string]interface{}{
			"account_address": tx.RecipientAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceRecipientQ...)
	// recipient Ledger Log
	recipientAccountLedgerQ, recipientAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.RecipientAddress,
		BalanceChange:  tx.GetAmount(),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventSendMoneyTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	recipientAccountLedgerArgs = append([]interface{}{recipientAccountLedgerQ}, recipientAccountLedgerArgs...)
	queries = append(queries, recipientAccountLedgerArgs)

	// update sender
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.Amount + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// sender ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(tx.GetAmount() + tx.Fee),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventSendMoneyTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	err = tx.QueryExecutor.ExecuteTransactions(queries)

	if err != nil {
		return err
	}
	return nil
}

/*
ApplyUnconfirmed is func that for applying to unconfirmed Transaction `SendMoney` type:
	- perhaps recipient is not exists , so create new `account` and `account_balance`, balance and spendable = amount.
*/
func (tx *SendMoney) ApplyUnconfirmed() error {

	var (
		err error
	)

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Body.Amount + tx.Fee),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}

	return nil
}

/*
UndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *SendMoney) UndoApplyUnconfirmed() error {
	var (
		err error
	)

	// update sender
	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Body.Amount+tx.Fee,
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)
	err = tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}

	return nil
}

/*
Validate is func that for validating to Transaction SendMoney type
That specs:
	- If Genesis, sender and recipient allowed not exists,
	- If Not Genesis,  sender and recipient must be exists, `sender.spendable_balance` must bigger than amount
*/
func (tx *SendMoney) Validate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		accountDataset model.AccountDataset
		accDatasetArgs []interface{}
		accDatasetQ    string
		row            *sql.Row
		err            error
	)

	if tx.Body.GetAmount() <= 0 {
		return errors.New("transaction must have an amount more than 0")
	}
	if tx.RecipientAddress == "" {
		return errors.New("transaction must have a valid recipient account id")
	}
	// checking the recipient has an model.AccountDatasetProperty_AccountDatasetEscrowApproval
	// yes would be error
	// TODO: Move this part to `transactionCoreService` when all transaction types need this part
	accDatasetQ, accDatasetArgs = tx.AccountDatasetQuery.GetAccountDatasetEscrowApproval(tx.RecipientAddress)
	row, err = tx.QueryExecutor.ExecuteSelectRow(accDatasetQ, dbTx, accDatasetArgs...)
	if err != nil {
		return err
	}
	err = tx.AccountDatasetQuery.Scan(&accountDataset, row)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	// false if err in above is sql.ErrNoRows || nil
	if accountDataset.GetIsActive() {
		return fmt.Errorf("RecipientRequireEscrow")
	}
	// todo: this is temporary solution, later we should depend on coinbase, so no genesis transaction exclusion in
	// validation needed
	if tx.SenderAddress != constant.MainchainGenesisAccountAddress {
		if tx.SenderAddress == "" {
			return errors.New("transaction must have a valid sender account id")
		}

		qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
		row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
		if err != nil {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}

		err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
		if err != nil {
			return err
		}

		if accountBalance.SpendableBalance < (tx.Body.GetAmount() + tx.Fee) {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"UserBalanceNotEnough",
			)
		}
	}
	return nil
}

// GetAmount return Amount from TransactionBody
func (tx *SendMoney) GetAmount() int64 {
	return tx.Body.Amount
}

func (tx *SendMoney) GetMinimumFee() (int64, error) {
	if tx.Escrow.ApproverAddress != "" {
		return tx.EscrowFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
	}
	return tx.NormalFee.CalculateTxMinimumFee(tx.Body, tx.Escrow)
}

// GetSize send money Amount should be 8
func (*SendMoney) GetSize() uint32 {
	// only amount
	return constant.Balance
}

// ParseBodyBytes read and translate body bytes to body implementation fields
func (tx *SendMoney) ParseBodyBytes(txBodyBytes []byte) (model.TransactionBodyInterface, error) {
	// validate the body bytes is correct
	_, err := util.ReadTransactionBytes(bytes.NewBuffer(txBodyBytes), int(tx.GetSize()))
	if err != nil {
		return nil, err
	}
	// read body bytes
	bufferBytes := bytes.NewBuffer(txBodyBytes)
	amount := util.ConvertBytesToUint64(bufferBytes.Next(int(constant.Balance)))
	return &model.SendMoneyTransactionBody{
		Amount: int64(amount),
	}, nil
}

// GetBodyBytes translate tx body to bytes representation
func (tx *SendMoney) GetBodyBytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(tx.Body.Amount)))
	return buffer.Bytes()
}

// GetTransactionBody append isTransaction_TransactionBody oneOf
func (tx *SendMoney) GetTransactionBody(transaction *model.Transaction) {
	transaction.TransactionBody = &model.Transaction_SendMoneyTransactionBody{
		SendMoneyTransactionBody: tx.Body,
	}
}

/*
Escrowable will check the transaction is escrow or not.
Rebuild escrow if not nil, and can use for whole sibling methods (escrow)
*/
func (tx *SendMoney) Escrowable() (EscrowTypeAction, bool) {
	if tx.Escrow.GetApproverAddress() != "" {
		tx.Escrow = &model.Escrow{
			ID:               tx.ID,
			SenderAddress:    tx.SenderAddress,
			RecipientAddress: tx.RecipientAddress,
			ApproverAddress:  tx.Escrow.GetApproverAddress(),
			Amount:           tx.Body.GetAmount(),
			Commission:       tx.Escrow.GetCommission(),
			Timeout:          tx.Escrow.GetTimeout(),
			Status:           tx.Escrow.GetStatus(),
			BlockHeight:      tx.Height,
			Latest:           true,
			Instruction:      tx.Escrow.GetInstruction(),
		}

		return EscrowTypeAction(tx), true
	}
	return nil, false
}

// EscrowValidate special validation for escrow's transaction
func (tx *SendMoney) EscrowValidate(dbTx bool) error {
	var (
		accountBalance model.AccountBalance
		err            error
		row            *sql.Row
	)

	if tx.Body.GetAmount() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "AmountNotEnough")
	}
	if tx.Escrow.GetCommission() <= 0 {
		return blocker.NewBlocker(blocker.ValidationErr, "CommissionNotEnough")
	}
	if tx.Escrow.GetApproverAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetRecipientAddress() == "" {
		return blocker.NewBlocker(blocker.ValidationErr, "RecipientAddressRequired")
	}
	if tx.Escrow.GetTimeout() > uint64(constant.MinRollbackBlocks) {
		return blocker.NewBlocker(blocker.ValidationErr, "TimeoutLimitExceeded")
	}
	// todo: this is temporary solution, later we should depend on coinbase, so no genesis transaction exclusion in
	// validation needed
	if tx.SenderAddress != constant.MainchainGenesisAccountAddress {
		if tx.SenderAddress == "" {
			return blocker.NewBlocker(blocker.ValidationErr, "SenderAddressRequired")
		}

		qry, args := tx.AccountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAddress)
		row, err = tx.QueryExecutor.ExecuteSelectRow(qry, dbTx, args...)
		if err != nil {
			return err
		}

		err = tx.AccountBalanceQuery.Scan(&accountBalance, row)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			return blocker.NewBlocker(blocker.ValidationErr, "AccountBalanceNotFound")
		}

		if accountBalance.SpendableBalance < (tx.Body.GetAmount() + tx.Fee + tx.Escrow.GetCommission()) {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"UserBalanceNotEnough",
			)
		}
	}
	return nil

}

/*
EscrowApplyUnconfirmed is applyUnconfirmed specific for Escrow's transaction
similar with ApplyUnconfirmed and Escrow.Commission
*/
func (tx *SendMoney) EscrowApplyUnconfirmed() error {

	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		-(tx.Body.GetAmount() + tx.Fee + tx.Escrow.GetCommission()),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)

	err := tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowUndoApplyUnconfirmed is used to undo the previous applied unconfirmed tx action
this will be called on apply confirmed or when rollback occurred
*/
func (tx *SendMoney) EscrowUndoApplyUnconfirmed() error {

	accountBalanceSenderQ, accountBalanceSenderQArgs := tx.AccountBalanceQuery.AddAccountSpendableBalance(
		tx.Body.Amount+tx.Fee+tx.Escrow.GetCommission(),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
		},
	)

	err := tx.QueryExecutor.ExecuteTransaction(accountBalanceSenderQ, accountBalanceSenderQArgs...)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApplyConfirmed func that for applying Transaction SendMoney type, insert and update balance,
account ledger, and escrow
*/
func (tx *SendMoney) EscrowApplyConfirmed(blockTimestamp int64) error {
	var (
		queries [][]interface{}
		err     error
	)

	// update sender balance
	accountBalanceSenderQ := tx.AccountBalanceQuery.AddAccountBalance(
		-(tx.Body.Amount + tx.Fee + tx.Escrow.GetCommission()),
		map[string]interface{}{
			"account_address": tx.SenderAddress,
			"block_height":    tx.Height,
		},
	)
	queries = append(queries, accountBalanceSenderQ...)

	// sender ledger
	senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
		AccountAddress: tx.SenderAddress,
		BalanceChange:  -(tx.Body.GetAmount() + tx.Fee + tx.Escrow.GetCommission()),
		TransactionID:  tx.ID,
		BlockHeight:    tx.Height,
		EventType:      model.EventType_EventSendMoneyTransaction,
		Timestamp:      uint64(blockTimestamp),
	})
	senderAccountLedgerArgs = append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...)
	queries = append(queries, senderAccountLedgerArgs)

	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)

	err = tx.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

/*
EscrowApproval handle approval an escrow transaction, execute tasks that was skipped when escrow pending.
like: spreading commission and fee, and also more pending tasks
*/
func (tx *SendMoney) EscrowApproval(
	blockTimestamp int64,
	txBody *model.ApprovalEscrowTransactionBody,
) error {
	var (
		queries [][]interface{}
		err     error
	)

	switch txBody.GetApproval() {
	case model.EscrowApproval_Approve:
		tx.Escrow.Status = model.EscrowStatus_Approved
		// insert / update recipient balance
		accountBalanceRecipientQ := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Body.Amount,
			map[string]interface{}{
				"account_address": tx.Escrow.GetRecipientAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, accountBalanceRecipientQ...)

		// recipient Account Ledger Log
		recipientAccountLedgerQ, recipientAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.Escrow.GetRecipientAddress(),
			BalanceChange:  tx.Body.GetAmount(),
			TransactionID:  tx.ID,
			BlockHeight:    tx.Height,
			EventType:      model.EventType_EventSendMoneyTransaction,
			Timestamp:      uint64(blockTimestamp),
		})
		recipientAccountLedgerArgs = append([]interface{}{recipientAccountLedgerQ}, recipientAccountLedgerArgs...)
		queries = append(queries, recipientAccountLedgerArgs)

		// approver balance
		approverBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Escrow.GetCommission(),
			map[string]interface{}{
				"account_address": tx.Escrow.GetApproverAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, approverBalanceQ...)
		// approver ledger
		approverAccountLedgerQ, approverAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.Escrow.GetApproverAddress(),
			BalanceChange:  tx.Escrow.GetCommission(),
			BlockHeight:    tx.Height,
			TransactionID:  tx.ID,
			Timestamp:      uint64(blockTimestamp),
			EventType:      model.EventType_EventApprovalEscrowTransaction,
		})
		approverAccountLedgerArgs = append([]interface{}{approverAccountLedgerQ}, approverAccountLedgerArgs...)
		queries = append(queries, approverAccountLedgerArgs)
	case model.EscrowApproval_Reject:
		tx.Escrow.Status = model.EscrowStatus_Rejected
		// Give back sender balance
		senderBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Escrow.GetAmount(),
			map[string]interface{}{
				"account_address": tx.Escrow.GetSenderAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, senderBalanceQ...)

		// sender ledger
		senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.Escrow.GetSenderAddress(),
			BalanceChange:  tx.Escrow.GetAmount(),
			BlockHeight:    tx.Height,
			TransactionID:  tx.ID,
			Timestamp:      uint64(blockTimestamp),
			EventType:      model.EventType_EventApprovalEscrowTransaction,
		})
		queries = append(queries, append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...))

		// approver balance
		approverBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Escrow.GetCommission(),
			map[string]interface{}{
				"account_address": tx.Escrow.GetApproverAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, approverBalanceQ...)
		// approver ledger
		approverAccountLedgerQ, approverAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.Escrow.GetApproverAddress(),
			BalanceChange:  tx.Escrow.GetCommission(),
			BlockHeight:    tx.Height,
			TransactionID:  tx.ID,
			Timestamp:      uint64(blockTimestamp),
			EventType:      model.EventType_EventApprovalEscrowTransaction,
		})
		queries = append(queries, append([]interface{}{approverAccountLedgerQ}, approverAccountLedgerArgs...))

	default:
		tx.Escrow.Status = model.EscrowStatus_Expired
		// sender balance
		senderBalanceQ := tx.AccountBalanceQuery.AddAccountBalance(
			tx.Escrow.GetCommission()+tx.Escrow.GetAmount(),
			map[string]interface{}{
				"account_address": tx.Escrow.GetSenderAddress(),
				"block_height":    tx.Height,
			},
		)
		queries = append(queries, senderBalanceQ...)

		// sender ledger
		senderAccountLedgerQ, senderAccountLedgerArgs := tx.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: tx.Escrow.GetSenderAddress(),
			BalanceChange:  tx.Escrow.GetCommission() + tx.Escrow.GetAmount(),
			BlockHeight:    tx.Height,
			TransactionID:  tx.ID,
			Timestamp:      uint64(blockTimestamp),
			EventType:      model.EventType_EventApprovalEscrowTransaction,
		})
		queries = append(queries, append([]interface{}{senderAccountLedgerQ}, senderAccountLedgerArgs...))
	}

	// Insert Escrow
	escrowArgs := tx.EscrowQuery.InsertEscrowTransaction(tx.Escrow)
	queries = append(queries, escrowArgs...)
	err = tx.QueryExecutor.ExecuteTransactions(queries)

	if err != nil {
		return err
	}
	return nil
}
