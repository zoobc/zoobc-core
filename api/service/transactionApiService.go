package service

import (
	"database/sql"
	"fmt"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	TransactionServiceInterface interface {
		GetTransaction(contract.ChainType, *model.GetTransactionRequest) (*model.Transaction, error)
		GetTransactions(contract.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error)
	}

	// TransactionService represents struct of TransactionService
	TransactionService struct {
		Query *query.Executor
	}
)

var transactionServiceInstance *TransactionService

// NewTransactionService creates a singleton instance of TransactionService
func NewTransactionService(queryExecutor *query.Executor) *TransactionService {
	if transactionServiceInstance == nil {
		transactionServiceInstance = &TransactionService{Query: queryExecutor}
	}
	return transactionServiceInstance
}

// GetTransaction fetches a single transaction from DB
func (ts *TransactionService) GetTransaction(chainType contract.ChainType, params *model.GetTransactionRequest) (*model.Transaction, error) {
	var (
		err    error
		rows   *sql.Rows
		txTemp model.Transaction
	)
	rows, err = ts.Query.ExecuteSelect(query.NewTransactionQuery(chainType).GetTransaction(params.ID))
	if err != nil {
		fmt.Printf("GetTransaction fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&txTemp.ID,
			&txTemp.BlockID,
			&txTemp.Height,
			&txTemp.SenderAccountType,
			&txTemp.SenderAccountAddress,
			&txTemp.RecipientAccountType,
			&txTemp.RecipientAccountAddress,
			&txTemp.TransactionType,
			&txTemp.Fee,
			&txTemp.Timestamp,
			&txTemp.TransactionHash,
			&txTemp.TransactionBodyLength,
			&txTemp.TransactionBodyBytes,
			&txTemp.Signature,
		)
	}
	return &txTemp, nil
}

// GetTransactions fetches a single transaction from DB
func (ts *TransactionService) GetTransactions(chainType contract.ChainType, params *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error) {
	var (
		err          error
		rows         *sql.Rows
		rows2        *sql.Rows
		results      []*model.Transaction
		totalRecords uint64
	)
	selectQuery := query.NewTransactionQuery(chainType).GetTransactions(params.Limit, params.Offset)
	rows, err = ts.Query.ExecuteSelect(selectQuery)
	if err != nil {
		fmt.Printf("GetTransactions fails %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var txTemp model.Transaction
		err = rows.Scan(
			&txTemp.ID,
			&txTemp.BlockID,
			&txTemp.Height,
			&txTemp.SenderAccountType,
			&txTemp.SenderAccountAddress,
			&txTemp.RecipientAccountType,
			&txTemp.RecipientAccountAddress,
			&txTemp.TransactionType,
			&txTemp.Fee,
			&txTemp.Timestamp,
			&txTemp.TransactionHash,
			&txTemp.TransactionBodyLength,
			&txTemp.TransactionBodyBytes,
			&txTemp.Signature,
		)
		results = append(results, &txTemp)
	}

	rows2, err = ts.Query.ExecuteSelect(query.GetTotalRecordOfSelect(selectQuery))
	if err != nil {
		fmt.Printf("GetTransactions total records fails %v\n", err)
		return nil, err
	}
	defer rows2.Close()

	if rows2.Next() {
		err = rows2.Scan(
			&totalRecords,
		)

	}

	return &model.GetTransactionsResponse{
		Total:        totalRecords,
		Count:        uint32(len(results)),
		Transactions: results,
	}, nil
}
