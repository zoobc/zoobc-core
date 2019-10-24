package service

import (
	"database/sql"
	"math"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// TransactionServiceInterface represents interface for TransactionService
	TransactionServiceInterface interface {
		GetTransaction(chaintype.ChainType, *model.GetTransactionRequest) (*model.Transaction, error)
		GetTransactions(chaintype.ChainType, *model.GetTransactionsRequest) (*model.GetTransactionsResponse, error)
		PostTransaction(chaintype.ChainType, *model.PostTransactionRequest) (*model.Transaction, error)
	}

	// TransactionService represents struct of TransactionService
	TransactionService struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Observer           *observer.Observer
	}
)

var transactionServiceInstance *TransactionService

// NewTransactionService creates a singleton instance of TransactionService
func NewTransactionService(
	queryExecutor query.ExecutorInterface,
	signature crypto.SignatureInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	mempoolService service.MempoolServiceInterface,
	observer *observer.Observer,
) *TransactionService {
	if transactionServiceInstance == nil {
		transactionServiceInstance = &TransactionService{
			Query:              queryExecutor,
			Signature:          signature,
			ActionTypeSwitcher: txTypeSwitcher,
			MempoolService:     mempoolService,
			Observer:           observer,
		}
	}
	return transactionServiceInstance
}

// GetTransaction fetches a single transaction from DB
func (ts *TransactionService) GetTransaction(
	chainType chaintype.ChainType,
	params *model.GetTransactionRequest,
) (*model.Transaction, error) {
	var (
		err error
		row *sql.Row
		tx  model.Transaction
	)

	txQuery := query.NewTransactionQuery(chainType)
	row = ts.Query.ExecuteSelectRow(txQuery.GetTransaction(params.GetID()))
	err = txQuery.Scan(&tx, row)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	txType, err := ts.ActionTypeSwitcher.GetTransactionType(&tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	txType.GetTransactionBody(&tx)
	return &tx, nil
}

// GetTransactions fetches a single transaction from DB
// included filters
func (ts *TransactionService) GetTransactions(
	chainType chaintype.ChainType,
	params *model.GetTransactionsRequest,
) (*model.GetTransactionsResponse, error) {
	var (
		err          error
		rows         *sql.Rows
		rows2        *sql.Rows
		txs          []*model.Transaction
		selectQuery  string
		args         []interface{}
		totalRecords uint64
		txQuery      = query.NewTransactionQuery(chainType)
		caseQuery    = query.NewCaseQuery()
		// Represent transaction fields
		txFields = map[string]string{
			"Height":  "block_height",
			"BlockID": "block_id",
		}
	)
	caseQuery.Select(txQuery.TableName, txQuery.Fields...)

	page := params.GetPagination()
	height := params.GetHeight()
	if height != 0 {
		caseQuery.Where(caseQuery.Equal("block_height", height))
		if page != nil && page.GetLimit() == 0 {
			page.Limit = math.MaxUint32
		}
	}

	timestampStart := params.GetTimestampStart()
	timestampEnd := params.GetTimestampEnd()
	if timestampStart > 0 {
		caseQuery.And(caseQuery.Between("timestamp", timestampStart, timestampEnd))
	}

	transactionType := params.GetTransactionType()
	if transactionType > 0 {
		caseQuery.And(caseQuery.Equal("transaction_type", transactionType))
	}

	accountAddress := params.GetAccountAddress()
	if accountAddress != "" {
		caseQuery.AndOr(
			caseQuery.Equal("sender_account_address", accountAddress),
			caseQuery.Equal("recipient_account_address", accountAddress),
		)
	}
	selectQuery, args = caseQuery.Build()

	// count first
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	rows, err = ts.Query.ExecuteSelect(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&totalRecords,
		)
		if err != nil {
			return &model.GetTransactionsResponse{}, status.Error(codes.Internal, err.Error())
		}
	}

	// Get Transactions with Pagination
	if page.GetOrderField() == "" || txFields[page.GetOrderField()] == "" {
		caseQuery.OrderBy("timestamp", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(txFields[page.GetOrderField()], page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())
	selectQuery, args = caseQuery.Build()

	rows2, err = ts.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows2.Close()

	for rows2.Next() {
		var tx model.Transaction
		err = rows2.Scan(
			&tx.ID,
			&tx.BlockID,
			&tx.Height,
			&tx.SenderAccountAddress,
			&tx.RecipientAccountAddress,
			&tx.TransactionType,
			&tx.Fee,
			&tx.Timestamp,
			&tx.TransactionHash,
			&tx.TransactionBodyLength,
			&tx.TransactionBodyBytes,
			&tx.Signature,
			&tx.Version,
			&tx.TransactionIndex,
		)
		if err != nil {
			return nil, err
		}
		txType, err := ts.ActionTypeSwitcher.GetTransactionType(&tx)
		if err != nil {
			return nil, err
		}
		txType.GetTransactionBody(&tx)
		txs = append(txs, &tx)
	}

	return &model.GetTransactionsResponse{
		Total:        totalRecords,
		Transactions: txs,
	}, nil
}

func (ts *TransactionService) PostTransaction(
	chaintype chaintype.ChainType,
	req *model.PostTransactionRequest,
) (*model.Transaction, error) {
	txBytes := req.TransactionBytes
	// get unsigned bytes
	tx, err := util.ParseTransactionBytes(txBytes, true)
	if err != nil {
		return nil, err
	}
	// Validate Tx
	txType, err := ts.ActionTypeSwitcher.GetTransactionType(tx)
	if err != nil {
		return nil, err
	}
	// Save to mempool
	mpTx := &model.MempoolTransaction{
		FeePerByte:              constant.TxFeePerByte,
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

	ts.Observer.Notify(observer.TransactionAdded, mpTx.GetTransactionBytes(), chaintype)
	// return parsed transaction
	return tx, nil
}
