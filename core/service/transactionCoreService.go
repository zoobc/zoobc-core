package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	TransactionCoreServiceInterface interface {
		GetTransactionsByIds(transactionIds []int64) ([]*model.Transaction, error)
		GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error)
		ValidateTransaction(txAction transaction.TypeAction, useTX bool) error
		ApplyUnconfirmedTransaction(txAction transaction.TypeAction) error
		UndoApplyUnconfirmedTransaction(txAction transaction.TypeAction) error
		ApplyConfirmedTransaction(txAction transaction.TypeAction, blockTimestamp int64) error
		ExpiringEscrowTransactions(blockHeight uint32, blockTimestamp int64, useTX bool) error
		CompletePassedLiquidPayment(block *model.Block) error
	}

	TransactionCoreService struct {
		Log                           *logrus.Logger
		QueryExecutor                 query.ExecutorInterface
		TypeActionSwitcher            transaction.TypeActionSwitcher
		TransactionUtil               transaction.UtilInterface
		TransactionQuery              query.TransactionQueryInterface
		EscrowTransactionQuery        query.EscrowTransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
	}
)

func NewTransactionCoreService(
	log *logrus.Logger,
	queryExecutor query.ExecutorInterface,
	typeActionSwitcher transaction.TypeActionSwitcher,
	transactionUtil transaction.UtilInterface,
	transactionQuery query.TransactionQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
	liquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface,
) TransactionCoreServiceInterface {
	return &TransactionCoreService{
		Log:                           log,
		QueryExecutor:                 queryExecutor,
		TypeActionSwitcher:            typeActionSwitcher,
		TransactionUtil:               transactionUtil,
		TransactionQuery:              transactionQuery,
		EscrowTransactionQuery:        escrowTransactionQuery,
		LiquidPaymentTransactionQuery: liquidPaymentTransactionQuery,
	}
}

func (tg *TransactionCoreService) GetTransactionsByIds(transactionIds []int64) ([]*model.Transaction, error) {
	var (
		transactions = make([]*model.Transaction, 0)
		escrows      []*model.Escrow
		txMap        = make(map[int64]*model.Transaction)
		rows         *sql.Rows
		err          error
	)

	transactions, err = func() ([]*model.Transaction, error) {
		txQuery, args := tg.TransactionQuery.GetTransactionsByIds(transactionIds)
		rows, err = tg.QueryExecutor.ExecuteSelect(txQuery, false, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		return tg.TransactionQuery.BuildModel(transactions, rows)
	}()
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, tx := range transactions {
		txMap[tx.GetID()] = tx
		ids = append(ids, strconv.FormatInt(tx.GetID(), 10))
	}
	if len(ids) > 0 {
		escrows, err = func() ([]*model.Escrow, error) {
			escrowQ := tg.EscrowTransactionQuery.GetEscrowTransactionsByTransactionIdsAndStatus(ids, model.EscrowStatus_Pending)
			rows, err = tg.QueryExecutor.ExecuteSelect(escrowQ, false)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			return tg.EscrowTransactionQuery.BuildModels(rows)

		}()
		if err != nil {
			return nil, err
		}

		for _, escrow := range escrows {
			if _, ok := txMap[escrow.GetID()]; ok {
				txMap[escrow.GetID()].Escrow = escrow
			} else {
				return nil, fmt.Errorf("escrow ID and Transaction ID Did not match")
			}
		}
	}
	return transactions, nil
}

// GetTransactionsByBlockID get transactions of the block
func (tg *TransactionCoreService) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	var (
		transactionsMap = make(map[int64]*model.Transaction)
		transactions    []*model.Transaction
		escrows         []*model.Escrow
		txIdsStr        []string
		err             error
	)

	// get transaction of the block
	transactions, err = func() ([]*model.Transaction, error) {
		transactionQ, transactionArg := tg.TransactionQuery.GetTransactionsByBlockID(blockID)
		rows, err := tg.QueryExecutor.ExecuteSelect(transactionQ, false, transactionArg...)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		defer rows.Close()
		return tg.TransactionQuery.BuildModel(transactions, rows)
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// fetch escrow if exist
	for _, tx := range transactions {
		txIdsStr = append(txIdsStr, "'"+strconv.FormatInt(tx.ID, 10)+"'")
		transactionsMap[tx.ID] = tx
	}
	if len(txIdsStr) > 0 {
		escrows, err = func() ([]*model.Escrow, error) {
			escrowQ := tg.EscrowTransactionQuery.GetEscrowTransactionsByTransactionIdsAndStatus(
				txIdsStr, model.EscrowStatus_Pending,
			)
			rows, err := tg.QueryExecutor.ExecuteSelect(escrowQ, false)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			return tg.EscrowTransactionQuery.BuildModels(rows)
		}()
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		for _, escrow := range escrows {
			transactionsMap[escrow.ID].Escrow = escrow
		}
	}
	return transactions, nil
}

// ExpiringEscrowTransactions push an observer event that is ExpiringEscrowTransactions,
// will set status to be expired caused by current block height
// query lock from outside (PushBlock)
func (tg *TransactionCoreService) ExpiringEscrowTransactions(blockHeight uint32, blockTimestamp int64, useTX bool) error {
	var (
		escrows []*model.Escrow
		rows    *sql.Rows
		err     error
	)

	err = func() error {
		escrowQ := tg.EscrowTransactionQuery.GetExpiredEscrowTransactionsAtCurrentBlock(blockHeight)
		rows, err = tg.QueryExecutor.ExecuteSelect(escrowQ, useTX)
		if err != nil {
			return err
		}
		defer rows.Close()

		escrows, err = tg.EscrowTransactionQuery.BuildModels(rows)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}

	if len(escrows) > 0 {
		if !useTX {
			err = tg.QueryExecutor.BeginTx()
			if err != nil {
				return err
			}
		}
		for _, escrow := range escrows {
			var (
				refTransaction model.Transaction
				typeAction     transaction.TypeAction
				row            *sql.Row
			)
			/**
			SET Escrow
			2. status = expired
			*/
			row, err = tg.QueryExecutor.ExecuteSelectRow(tg.TransactionQuery.GetTransaction(escrow.GetID()), useTX)
			if err != nil {
				break
			}
			err = tg.TransactionQuery.Scan(&refTransaction, row)
			if err != nil {
				break
			}

			refTransaction.Height = blockHeight
			refTransaction.Escrow = escrow
			typeAction, err = tg.TypeActionSwitcher.GetTransactionType(&refTransaction)
			if err != nil {
				break
			}
			if escrowTypAction, ok := typeAction.Escrowable(); ok {
				err = escrowTypAction.EscrowApproval(blockTimestamp, &model.ApprovalEscrowTransactionBody{
					Approval:      model.EscrowApproval_Expire,
					TransactionID: escrow.GetID(),
				})
				if err != nil {
					break
				}
			}
		}

		if !useTX {
			/*
				Check the latest error is not nil, otherwise need to aborting the whole query transactions safety with rollBack.
				And automatically unlock mutex
			*/
			if err != nil {
				if rollbackErr := tg.QueryExecutor.RollbackTx(); rollbackErr != nil {
					tg.Log.Errorf("Rollback fail: %s", rollbackErr.Error())
				}
				return err
			}

			err = tg.QueryExecutor.CommitTx()
			if err != nil {
				if rollbackErr := tg.QueryExecutor.RollbackTx(); rollbackErr != nil {
					tg.Log.Errorf("Rollback fail: %s", rollbackErr.Error())
				}
				return err
			}
		}
	}
	return nil
}

func (tg *TransactionCoreService) CompletePassedLiquidPayment(block *model.Block) error {
	var (
		rows           *sql.Rows
		row            *sql.Row
		err            error
		liquidPayments []*model.LiquidPayment
		tx             model.Transaction
		txType         transaction.TypeAction
	)
	liquidPayments, err = func() ([]*model.LiquidPayment, error) {
		liquidPaymentQ, liquidPaymentArgs := tg.LiquidPaymentTransactionQuery.GetPassedTimePendingLiquidPaymentTransactions(block.GetTimestamp())
		rows, err = tg.QueryExecutor.ExecuteSelect(liquidPaymentQ, true, liquidPaymentArgs...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return tg.LiquidPaymentTransactionQuery.BuildModels(rows)
	}()
	if err != nil {
		return err
	}

	for _, payment := range liquidPayments {
		// get what transaction type it is, and switch to specific approval
		transactionQ := tg.TransactionQuery.GetTransaction(payment.ID)
		row, err = tg.QueryExecutor.ExecuteSelectRow(transactionQ, false)
		if err != nil {
			return err
		}
		err = tg.TransactionQuery.Scan(&tx, row)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			return blocker.NewBlocker(blocker.AppErr, "TransactionNotFound")

		}

		txType, err = tg.TypeActionSwitcher.GetTransactionType(&tx)
		if err != nil {
			return err
		}
		liquidPaymentTransaction, ok := txType.(transaction.LiquidPaymentTransactionInterface)
		if !ok {
			return blocker.NewBlocker(blocker.AppErr, "Wrong type of transaction")
		}
		err = liquidPaymentTransaction.CompletePayment(block.GetHeight(), block.GetTimestamp(), payment.AppliedTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tg *TransactionCoreService) ValidateTransaction(txAction transaction.TypeAction, useTX bool) error {
	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		return escrowAction.EscrowValidate(useTX)
	default:
		return txAction.Validate(useTX)
	}
}

func (tg *TransactionCoreService) ApplyUnconfirmedTransaction(txAction transaction.TypeAction) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		err := escrowAction.EscrowApplyUnconfirmed()
		return err
	default:
		err := txAction.ApplyUnconfirmed()
		return err
	}
}

func (tg *TransactionCoreService) UndoApplyUnconfirmedTransaction(txAction transaction.TypeAction) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		err := escrowAction.EscrowUndoApplyUnconfirmed()
		return err
	default:
		err := txAction.UndoApplyUnconfirmed()
		return err
	}
}

func (tg *TransactionCoreService) ApplyConfirmedTransaction(
	txAction transaction.TypeAction,
	blockTimestamp int64,
) error {

	escrowAction, ok := txAction.Escrowable()
	switch ok {
	case true:
		err := escrowAction.EscrowApplyConfirmed(blockTimestamp)
		return err
	default:
		err := txAction.ApplyConfirmed(blockTimestamp)
		return err
	}
}
