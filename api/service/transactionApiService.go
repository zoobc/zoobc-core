package service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
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
func NewTransactionService(
	queryExecutor query.ExecutorInterface,
	signature crypto.SignatureInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	mempoolService service.MempoolServiceInterface,
	log *logrus.Logger,
) *TransactionService {
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
func (ts *TransactionService) GetTransaction(
	chainType contract.ChainType,
	params *model.GetTransactionRequest,
) (*model.Transaction, error) {
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
// included filters
func (ts *TransactionService) GetTransactions(
	chainType contract.ChainType,
	params *model.GetTransactionsRequest,
) (*model.GetTransactionsResponse, error) {
	var (
		err          error
		rows         *sql.Rows
		txs          []*model.Transaction
		selectQuery  string
		args         []interface{}
		totalRecords uint64
	)

	txQuery := query.NewTransactionQuery(chainType)
	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}
	caseQuery.Select(txQuery.TableName, txQuery.Fields...)

	accountAddress := params.GetAccountAddress()
	if accountAddress != "" {
		caseQuery.Where(caseQuery.Equal("sender_account_address", accountAddress)).
			Or(caseQuery.Equal("recipient_account_address", accountAddress))
	}

	selectQuery, args = caseQuery.Build()
	// count first
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	rows, err = ts.Query.ExecuteSelect(countQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&totalRecords,
		)
		if err != nil {
			return &model.GetTransactionsResponse{}, err
		}
	}

	// Get Transactions
	caseQuery.Paginate(params.GetLimit(), params.GetPage())
	selectQuery, args = caseQuery.Build()
	fmt.Println(selectQuery, args)
	rows, err = ts.Query.ExecuteSelect(selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs = txQuery.BuildModel(txs, rows)

	return &model.GetTransactionsResponse{
		Total:        totalRecords,
		Transactions: txs,
	}, nil
}

func (ts *TransactionService) PostTransaction(
	chaintype contract.ChainType,
	req *model.PostTransactionRequest,
) (*model.Transaction, error) {
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
		FeePerByte:              0,
		ID:                      tx.ID,
		TransactionBytes:        txBytes,
		ArrivalTimestamp:        time.Now().Unix(),
		SenderAccountAddress:    tx.SenderAccountAddress,
		RecipientAccountAddress: tx.RecipientAccountAddress,
	}
	if err := ts.MempoolService.ValidateMempoolTransaction(mpTx); err != nil {
		return nil, err
	}
	// Apply Unconfirmed
	err = ts.Query.BeginTx()
	if err != nil {
		return nil, err
	}
	err = txType.ApplyUnconfirmed()
	if err != nil {
		errRollback := ts.Query.RollbackTx()
		if errRollback != nil {
			return nil, errRollback
		}
		return nil, err
	}
	err = ts.MempoolService.AddMempoolTransaction(mpTx)
	if err != nil {
		errRollback := ts.Query.RollbackTx()
		if errRollback != nil {
			return nil, err
		}
		return nil, err
	}
	err = ts.Query.CommitTx()
	if err != nil {
		return nil, err
	}

	// return parsed transaction
	return tx, nil
}
