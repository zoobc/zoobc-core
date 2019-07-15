package service

import (
	"database/sql"
	"errors"
	"math"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

type mockMempoolQueryExecutorSuccess struct {
	query.Executor
}

func (*mockMempoolQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT ID, FeePerByte, ArrivalTimestamp, TransactionBytes FROM mempool":
		mockedRows := sqlmock.NewRows([]string{"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes"})
		mockedRows.AddRow(1, 1, 1562893305, []byte{5, 5, 5, 5, 5})
		mockedRows.AddRow(2, 10, 1562893304, []byte{2, 2, 2, 2, 2})
		mockedRows.AddRow(3, 1, 1562893302, []byte{1, 1, 1, 1, 1})
		mockedRows.AddRow(4, 100, 1562893306, []byte{4, 4, 4, 4, 4})
		mockedRows.AddRow(5, 5, 1562893303, []byte{4, 4, 4, 4, 4})
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	case "SELECT ID, FeePerByte, ArrivalTimestamp, TransactionBytes FROM mempool WHERE id = :id":
		return nil, errors.New("MempoolTransactionNotFound")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockMempoolQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

type mockMempoolQueryExecutorFail struct {
	query.Executor
}

func (*mockMempoolQueryExecutorFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	// before adding mempool transactions to db we check for duplicate transactions
	case "SELECT ID, FeePerByte, ArrivalTimestamp, TransactionBytes FROM mempool WHERE id = :id":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes"},
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

type mockMempoolQueryExecutorSQLFail struct {
	query.Executor
}

func (*mockMempoolQueryExecutorSQLFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

func buildTransaction(id int64, sender, recipient string) *model.Transaction {
	return &model.Transaction{
		Version:                 1,
		ID:                      id,
		BlockID:                 1,
		Height:                  1,
		SenderAccountType:       0,
		SenderAccountAddress:    sender,
		RecipientAccountType:    0,
		RecipientAccountAddress: recipient,
		TransactionType:         0,
		Fee:                     1,
		Timestamp:               1562893302,
		TransactionHash:         make([]byte, 32),
		TransactionBodyLength:   0,
		TransactionBodyBytes:    make([]byte, 0),
		TransactionBody:         nil,
		Signature:               make([]byte, 256),
	}
}

func getTestSignedMempoolTransaction() *model.MempoolTransaction {
	tx := buildTransaction(3, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN")
	txBytes, _ := util.GetTransactionBytes(tx, true)
	return &model.MempoolTransaction{
		ID:               1,
		FeePerByte:       1,
		ArrivalTimestamp: 1562893302,
		TransactionBytes: txBytes,
	}
}

func TestNewMempoolService(t *testing.T) {
	type args struct {
		ct            contract.ChainType
		queryExecutor query.ExecutorInterface
		mempoolQuery  query.MempoolQueryInterface
	}

	test := struct {
		name string
		args args
		want *MempoolService
	}{
		name: "NewBlockService:success",
		args: args{
			ct:            &chaintype.MainChain{},
			queryExecutor: nil,
			mempoolQuery:  nil,
		},
		want: &MempoolService{
			Chaintype:     &chaintype.MainChain{},
			QueryExecutor: nil,
			MempoolQuery:  nil,
		},
	}

	got := NewMempoolService(test.args.ct, test.args.queryExecutor, test.args.mempoolQuery)

	if !cmp.Equal(got, test.want) {
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
					TransactionBytes: []byte{5, 5, 5, 5, 5},
				},
				{
					ID:               2,
					FeePerByte:       10,
					ArrivalTimestamp: 1562893304,
					TransactionBytes: []byte{2, 2, 2, 2, 2},
				},
				{
					ID:               3,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893302,
					TransactionBytes: []byte{1, 1, 1, 1, 1},
				},
				{
					ID:               4,
					FeePerByte:       100,
					ArrivalTimestamp: 1562893306,
					TransactionBytes: []byte{4, 4, 4, 4, 4},
				},
				{
					ID:               5,
					FeePerByte:       5,
					ArrivalTimestamp: 1562893303,
					TransactionBytes: []byte{4, 4, 4, 4, 4},
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
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
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
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorSuccess{},
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(),
			},
			wantErr: false,
		},
		{
			name: "AddMempoolTransaction:DuplicateTransaction",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorFail{},
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(),
			},
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
			if err := mps.AddMempoolTransaction(tt.args.mpTx); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.AddMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolService_RemoveMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
	}
	type args struct {
		transactions []*model.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "RemoveMempoolTransaction:Success",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorSuccess{},
			},
			args: args{
				transactions: []*model.Transaction{
					buildTransaction(3, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
				},
			},
			wantErr: false,
		},
		{
			name: "RemoveMempoolTransaction:Fail",
			fields: fields{
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorFail{},
			},
			args: args{
				transactions: []*model.Transaction{
					buildTransaction(3, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"),
				},
			},
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
			if err := mps.RemoveMempoolTransactions(tt.args.transactions); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.RemoveMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolService_SelectTransactionsFromMempool(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
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
				Chaintype:     &chaintype.MainChain{},
				MempoolQuery:  query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor: &mockMempoolQueryExecutorSuccess{},
			},
			args: args{
				blockTimestamp: math.MaxInt64,
			},
			want: []*model.MempoolTransaction{
				{
					ID:               4,
					FeePerByte:       100,
					ArrivalTimestamp: 1562893306,
					TransactionBytes: []byte{4, 4, 4, 4, 4},
				},
				{
					ID:               2,
					FeePerByte:       10,
					ArrivalTimestamp: 1562893304,
					TransactionBytes: []byte{2, 2, 2, 2, 2},
				},
				{
					ID:               3,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893302,
					TransactionBytes: []byte{1, 1, 1, 1, 1},
				},
				{
					ID:               1,
					FeePerByte:       1,
					ArrivalTimestamp: 1562893305,
					TransactionBytes: []byte{5, 5, 5, 5, 5},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:     tt.fields.Chaintype,
				QueryExecutor: tt.fields.QueryExecutor,
				MempoolQuery:  tt.fields.MempoolQuery,
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
