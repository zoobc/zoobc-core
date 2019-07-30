package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	mockMempoolQueryExecutorSuccess struct {
		query.Executor
	}
)

var getTxByIDQuery = "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes FROM mempool WHERE id = :id"

func (*mockMempoolQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes FROM mempool":
		mockedRows := sqlmock.NewRows([]string{"id", "fee_per_byte", "arrival_timestamp", "transaction_bytes"})
		mockedRows.AddRow(1, 1, 1562893305, getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes)
		mockedRows.AddRow(2, 10, 1562893304, getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes)
		mockedRows.AddRow(3, 1, 1562893302, getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes)
		mockedRows.AddRow(4, 100, 1562893306, getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes)
		mockedRows.AddRow(5, 5, 1562893303, getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	case getTxByIDQuery:
		return nil, errors.New("MempoolTransactionNotFound")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockMempoolQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (*mockMempoolQueryExecutorSuccess) ExecuteTransaction(qe string, args ...interface{}) error {
	return nil
}

type mockMempoolQueryExecutorFail struct {
	query.Executor
}

func (*mockMempoolQueryExecutorFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	// before adding mempool transactions to db we check for duplicate transactions
	case getTxByIDQuery:
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id", "fee_per_byte", "arrival_timestamp", "transaction_bytes"},
		).AddRow(3, 1, 1562893302, []byte{}))
	default:
		return nil, errors.New("MockedError")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockMempoolQueryExecutorFail) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MockedError")
}

func (*mockMempoolQueryExecutorFail) ExecuteTransaction(qe string, args ...interface{}) error {
	return errors.New("MockedError")
}

func buildTransaction(timestamp int64, sender, recipient string) *model.Transaction {
	return &model.Transaction{
		Version:                 1,
		ID:                      2774809487,
		BlockID:                 1,
		Height:                  1,
		SenderAccountType:       0,
		SenderAccountAddress:    sender,
		RecipientAccountType:    0,
		RecipientAccountAddress: recipient,
		TransactionType:         0,
		Fee:                     1,
		Timestamp:               timestamp,
		TransactionHash:         make([]byte, 32),
		TransactionBodyLength:   0,
		TransactionBodyBytes:    make([]byte, 0),
		TransactionBody:         nil,
		Signature:               make([]byte, 64),
	}
}

func getTestSignedMempoolTransaction(id, timestamp int64) *model.MempoolTransaction {
	tx := buildTransaction(timestamp, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN")
	txBytes, _ := util.GetTransactionBytes(tx, true)
	return &model.MempoolTransaction{
		ID:               id,
		FeePerByte:       1,
		ArrivalTimestamp: timestamp,
		TransactionBytes: txBytes,
	}
}

func TestNewMempoolService(t *testing.T) {
	type args struct {
		ct                  contract.ChainType
		queryExecutor       query.ExecutorInterface
		mempoolQuery        query.MempoolQueryInterface
		actionTypeSwitcher  transaction.TypeActionSwitcher
		accountBalanceQuery query.AccountBalanceQueryInterface
	}

	test := struct {
		name string
		args args
		want *MempoolService
	}{
		name: "NewBlockService:success",
		args: args{
			ct: &chaintype.MainChain{},
		},
		want: &MempoolService{
			Chaintype: &chaintype.MainChain{},
		},
	}

	got := NewMempoolService(
		test.args.ct,
		test.args.queryExecutor,
		test.args.mempoolQuery,
		test.args.actionTypeSwitcher,
		test.args.accountBalanceQuery,
	)
	if !reflect.DeepEqual(got, test.want) {
		t.Errorf("NewMempoolService() = %v, want %v", got, test.want)
	}
}

func TestMempoolService_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*model.MempoolTransaction
		wantErr bool
	}{
		{
			name: "GetMempoolTransactions:Success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorSuccess{},
			},
			want: []*model.MempoolTransaction{
				{
					ID:               1,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893305,
					TransactionBytes: getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
				},
				{
					ID:               2,
					FeePerByte:       10,
					ArrivalTimestamp: 1562893304,
					TransactionBytes: getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
				},
				{
					ID:               3,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893302,
					TransactionBytes: getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
				},
				{
					ID:               4,
					FeePerByte:       100,
					ArrivalTimestamp: 1562893306,
					TransactionBytes: getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes,
				},
				{
					ID:               5,
					FeePerByte:       5,
					ArrivalTimestamp: 1562893303,
					TransactionBytes: getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes,
				},
			},
			wantErr: false,
		},
		{
			name: "GetMempoolTransactions:Fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorFail{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				MempoolQuery:  tt.fields.MempoolQuery,
			}
			got, err := mps.GetMempoolTransactions()
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.GetMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolService.GetMempoolTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMempoolService_AddMempoolTransaction(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		MempoolQuery       query.MempoolQueryInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	type args struct {
		mpTx *model.MempoolTransaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "AddMempoolTransaction:Success",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:      &mockMempoolQueryExecutorSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893302),
			},
			wantErr: false,
		},
		{
			name: "AddMempoolTransaction:DuplicateTransaction",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:      &mockMempoolQueryExecutorFail{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893303),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				MempoolQuery:       tt.fields.MempoolQuery,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			if err := mps.AddMempoolTransaction(tt.args.mpTx); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.AddMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolService_SelectTransactionsFromMempool(t *testing.T) {
	type fields struct {
		Chaintype          contract.ChainType
		QueryExecutor      query.ExecutorInterface
		MempoolQuery       query.MempoolQueryInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
	}
	type args struct {
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.MempoolTransaction
		wantErr bool
	}{
		{
			name: "SelectTransactionsFromMempool:Success",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:      &mockMempoolQueryExecutorSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
			},
			args: args{
				blockTimestamp: 1562893106,
			},
			want: []*model.MempoolTransaction{
				{
					ID:               4,
					FeePerByte:       100,
					ArrivalTimestamp: 1562893306,
					TransactionBytes: getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes,
				},
				{
					ID:               2,
					FeePerByte:       10,
					ArrivalTimestamp: 1562893304,
					TransactionBytes: getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
				},
				{
					ID:               5,
					FeePerByte:       5,
					ArrivalTimestamp: 1562893303,
					TransactionBytes: getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes,
				},
				{
					ID:               3,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893302,
					TransactionBytes: getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
				},
				{
					ID:               1,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893305,
					TransactionBytes: getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				MempoolQuery:       tt.fields.MempoolQuery,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
			}
			got, err := mps.SelectTransactionsFromMempool(tt.args.blockTimestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.SelectTransactionsFromMempool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolService.SelectTransactionsFromMempool() = %v, want %v", got, tt.want)
			}
		})
	}
}
