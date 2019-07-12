package service

import (
	"database/sql"
	"errors"
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
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "FeePerByte", "ArrivalTimestamp", "TransactionBytes"},
		).AddRow(make([]byte, 32), 1, 1562893302, []byte{}))
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
		).AddRow(make([]byte, 32), 1, 1562893302, []byte{}))
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

func getTestSignedMempoolTransaction() *model.MempoolTransaction {
	tx := &model.Transaction{
		BlockID:               1,
		Fee:                   1,
		Height:                1,
		ID:                    []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		RecipientAccountID:    []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		SenderAccountID:       []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
		Signature:             make([]byte, 256),
		Timestamp:             1562893302,
		TransactionBodyBytes:  make([]byte, 0),
		TransactionBodyLength: 0,
		TransactionHash:       make([]byte, 32),
	}
	txBytes, _ := util.GetTransactionBytes(tx, true)
	return &model.MempoolTransaction{
		ID:               make([]byte, 32),
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
					ID:               make([]byte, 32),
					FeePerByte:       1,
					ArrivalTimestamp: 1562893302,
					TransactionBytes: []byte{},
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

func TestMempoolService_RemoveMempoolTransaction(t *testing.T) {
	type fields struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
	}
	type args struct {
		id []byte
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
				id: make([]byte, 32),
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
				id: make([]byte, 32),
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
			if err := mps.RemoveMempoolTransaction(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.RemoveMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
