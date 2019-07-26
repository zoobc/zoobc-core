package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	TransactionServiceInterface interface {
		GetTransaction(contract.ChainType, *model.GetTransactionRequest) (*model.Transaction, error)
		GetTransactions(contract.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error)
		PostTransaction(contract.ChainType, *model.PostTransactionRequest) (*model.Transaction, error)
	}

	// TransactionService represents struct of TransactionService
	TransactionService struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Log                *logrus.Logger
	}
)

var transactionServiceInstance *TransactionService

// NewTransactionService creates a singleton instance of TransactionService
func NewTransactionService(queryExecutor query.ExecutorInterface, signature crypto.SignatureInterface,
	txTypeSwitcher transaction.TypeActionSwitcher, mempoolService service.MempoolServiceInterface,
	log *logrus.Logger) *TransactionService {
	if transactionServiceInstance == nil {
		transactionServiceInstance = &TransactionService{
			Query:              queryExecutor,
			Signature:          signature,
			ActionTypeSwitcher: txTypeSwitcher,
			MempoolService:     mempoolService,
			Log:                log,
		}
	}
	return transactionServiceInstance
}

// GetTransaction fetches a single transaction from DB
func (ts *TransactionService) GetTransaction(chainType contract.ChainType,
	params *model.GetTransactionRequest) (*model.Transaction, error) {
	var (
		err    error
		rows   *sql.Rows
		txTemp []*model.Transaction
	)
	txQuery := query.NewTransactionQuery(chainType)
	rows, err = ts.Query.ExecuteSelect(txQuery.GetTransaction(params.ID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txTemp = txQuery.BuildModel(txTemp, rows)
	if len(txTemp) != 0 {
		return txTemp[0], nil
	}
	return nil, errors.New("TransactionNotFound")
}

// GetTransactions fetches a single transaction from DB
func (ts *TransactionService) GetTransactions(chainType contract.ChainType,
	params *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	var (
		err          error
		rows         *sql.Rows
		rows2        *sql.Rows
		txs          []*model.Transaction
		totalRecords uint64
	)
	txQuery := query.NewTransactionQuery(chainType)
	selectQuery := txQuery.GetTransactions(params.Limit, params.Offset)
	rows, err = ts.Query.ExecuteSelect(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	txs = txQuery.BuildModel(txs, rows)

	rows2, err = ts.Query.ExecuteSelect(query.GetTotalRecordOfSelect(selectQuery))
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	if rows2.Next() {
		err = rows2.Scan(
			&totalRecords,
		)

		if err != nil {
			return &model.GetTransactionsResponse{}, err
		}

	}

	return &model.GetTransactionsResponse{
		Total:        totalRecords,
		Count:        uint32(len(txs)),
		Transactions: txs,
	}, nil
}

func (ts *TransactionService) PostTransaction(chaintype contract.ChainType, req *model.PostTransactionRequest) (*model.Transaction, error) {
	txBytes := req.TransactionBytes
	// get unsigned bytes
	tx, err := util.ParseTransactionBytes(txBytes, true)
	if err != nil {
		return nil, err
	}
	// Validate Tx
	txType := ts.ActionTypeSwitcher.GetTransactionType(tx)

	// Save to mempool
	mpTx := &model.MempoolTransaction{
		FeePerByte:       0,
		ID:               tx.ID,
		TransactionBytes: txBytes,
		ArrivalTimestamp: time.Now().Unix(),
	}
	if err := ts.MempoolService.ValidateMempoolTransaction(mpTx); err != nil {
		ts.Log.Warnf("Invalid transaction submitted: %v", err)
		return nil, err
	}
	// Apply Unconfirmed
	err = ts.Query.BeginTx()
	if err != nil {
		ts.Log.Warnf("error opening db transaction %v", err)
		return nil, err
	}
	err = txType.ApplyUnconfirmed()
	if err != nil {
		ts.Log.Warnf("fail ApplyUnconfirmed tx: %v", err)
		errRollback := ts.Query.RollbackTx()
		if errRollback != nil {
			ts.Log.Warnf("error rolling back db transaction %v", errRollback)
			return nil, errRollback
		}
		return nil, err
	}
	err = ts.MempoolService.AddMempoolTransaction(mpTx)
	if err != nil {
		ts.Log.Warnf("error AddMempoolTransaction: %v", err)
		errRollback := ts.Query.RollbackTx()
		if errRollback != nil {
			ts.Log.Warnf("error rolling back db transaction %v", errRollback)
			return nil, err
		}
		return nil, err
	}
	err = ts.Query.CommitTx()
	if err != nil {
		ts.Log.Warnf("error committing db transaction: %v", err)
		return nil, err
	}

	// return parsed transaction
	return tx, nil
}
