package service

import (
	"database/sql"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zoobc/zoobc-core/observer"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
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
		err    error
		rows   *sql.Rows
		txTemp []*model.Transaction
		tx     *model.Transaction
	)

	txQuery := query.NewTransactionQuery(chainType)
	rows, err = ts.Query.ExecuteSelect(txQuery.GetTransaction(params.ID), false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	txTemp = txQuery.BuildModel(txTemp, rows)
	if len(txTemp) != 0 {
		tx = txTemp[0]
		txType, err := ts.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}
		parsedBody, err := txType.ParseBodyBytes(tx.GetTransactionBodyBytes())
		if err != nil {
			return nil, err
		}
		// TODO: need enhancement when parsing body bytes into body
		switch tx.GetTransactionType() {
		case uint32(model.TransactionType_SendMoneyTransaction):
			tx.TransactionBody = &model.Transaction_SendMoneyTransactionBody{
				SendMoneyTransactionBody: parsedBody.(*model.SendMoneyTransactionBody),
			}
		case uint32(model.TransactionType_NodeRegistrationTransaction):
			tx.TransactionBody = &model.Transaction_NodeRegistrationTransactionBody{
				NodeRegistrationTransactionBody: parsedBody.(*model.NodeRegistrationTransactionBody),
			}
		case uint32(model.TransactionType_UpdateNodeRegistrationTransaction):
			tx.TransactionBody = &model.Transaction_UpdateNodeRegistrationTransactionBody{
				UpdateNodeRegistrationTransactionBody: parsedBody.(*model.UpdateNodeRegistrationTransactionBody),
			}
		case uint32(model.TransactionType_RemoveNodeRegistrationTransaction):
			tx.TransactionBody = &model.Transaction_RemoveNodeRegistrationTransactionBody{
				RemoveNodeRegistrationTransactionBody: parsedBody.(*model.RemoveNodeRegistrationTransactionBody),
			}
		case uint32(model.TransactionType_ClaimNodeRegistrationTransaction):
			tx.TransactionBody = &model.Transaction_ClaimNodeRegistrationTransactionBody{
				ClaimNodeRegistrationTransactionBody: parsedBody.(*model.ClaimNodeRegistrationTransactionBody),
			}
		case uint32(model.TransactionType_SetupAccountDatasetTransaction):
			tx.TransactionBody = &model.Transaction_SetupAccountDatasetTransactionBody{
				SetupAccountDatasetTransactionBody: parsedBody.(*model.SetupAccountDatasetTransactionBody),
			}
		case uint32(model.TransactionType_RemoveAccountDatasetTransaction):
			tx.TransactionBody = &model.Transaction_RemoveAccountDatasetTransactionBody{
				RemoveAccountDatasetTransactionBody: parsedBody.(*model.RemoveAccountDatasetTransactionBody),
			}
		default:
			tx.TransactionBody = nil
		}

		return tx, nil
	}
	return nil, status.Error(codes.NotFound, "transaction not found")
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
		txs          []*model.Transaction
		selectQuery  string
		args         []interface{}
		totalRecords uint64
	)

	txQuery := query.NewTransactionQuery(chainType)
	caseQuery := query.NewCaseQuery()
	caseQuery.Select(txQuery.TableName, txQuery.Fields...)

	// Represent transaction fields
	txFields := map[string]string{
		"Height":  "block_height",
		"BlockID": "block_id",
	}

	accountAddress := params.GetAccountAddress()
	page := params.GetPagination()
	height := params.GetHeight()

	if height != 0 {
		caseQuery.Where(caseQuery.Equal("block_height", height))
		if page != nil && page.GetLimit() == 0 {
			page.Limit = math.MaxUint32
		}
	}

	if accountAddress != "" {
		caseQuery.And(caseQuery.Equal("sender_account_address", accountAddress)).
			Or(caseQuery.Equal("recipient_account_address", accountAddress))
	}
	timestampStart := params.GetTimestampStart()
	timestampEnd := params.GetTimestampEnd()
	if timestampStart > 0 {
		caseQuery.And(caseQuery.Between("timestamp", timestampStart, timestampEnd))
	}

	transcationType := params.GetTransactionType()
	if transcationType > 0 {
		caseQuery.And(caseQuery.Equal("transaction_type", transcationType))
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

	rows, err = ts.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	txs = txQuery.BuildModel(txs, rows)

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
