package service

import (
	"database/sql"
	"errors"
	"github.com/zoobc/zoobc-core/common/crypto"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	mockMempoolQueryExecutorSuccess struct {
		query.Executor
	}
)

var getTxByIDQuery = "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, " +
	"recipient_account_address FROM mempool WHERE id = :id"

func (*mockMempoolQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {

	case "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, recipient_account_address FROM mempool":
		mockedRows := sqlmock.NewRows([]string{"id", "fee_per_byte", "arrival_timestamp", "transaction_bytes", "sender_account_address",
			"recipient_account_address"})
		mockedRows.AddRow(1, 1, 1562893305, getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes, "A", "B")
		mockedRows.AddRow(2, 10, 1562893304, getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes, "A", "B")
		mockedRows.AddRow(3, 1, 1562893302, getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes, "A", "B")
		mockedRows.AddRow(4, 100, 1562893306, getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes, "A", "B")
		mockedRows.AddRow(5, 5, 1562893303, getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes, "A", "B")
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

func (*mockMempoolQueryExecutorSuccess) BeginTx() error {
	return nil
}

func (*mockMempoolQueryExecutorSuccess) CommitTx() error {
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
			"id", "fee_per_byte", "arrival_timestamp", "transaction_bytes", "sender_account_address", "recipient_account_address"},
		).AddRow(3, 1, 1562893302, []byte{}, []byte{1}, []byte{2}))
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
		SenderAccountAddress:    sender,
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
		ID:                      id,
		FeePerByte:              1,
		ArrivalTimestamp:        timestamp,
		TransactionBytes:        txBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	}
}

func TestNewMempoolService(t *testing.T) {
	type args struct {
		ct                  chaintype.ChainType
		queryExecutor       query.ExecutorInterface
		mempoolQuery        query.MempoolQueryInterface
		actionTypeSwitcher  transaction.TypeActionSwitcher
		accountBalanceQuery query.AccountBalanceQueryInterface
		obsr                *observer.Observer
		signature           crypto.SignatureInterface
	}

	test := struct {
		name string
		args args
		want *MempoolService
	}{
		name: "NewBlockService:success",
		args: args{
			ct:   &chaintype.MainChain{},
			obsr: observer.NewObserver(),
		},
		want: &MempoolService{
			Chaintype: &chaintype.MainChain{},
			Observer:  observer.NewObserver(),
		},
	}

	got := NewMempoolService(
		test.args.ct,
		test.args.queryExecutor,
		test.args.mempoolQuery,
		test.args.actionTypeSwitcher,
		test.args.accountBalanceQuery,
		test.args.signature,
		test.args.obsr,
	)
	if !reflect.DeepEqual(got, test.want) {
		t.Errorf("NewMempoolService() = %v, want %v", got, test.want)
	}
}

func TestMempoolService_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype     chaintype.ChainType
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
					ID:                      1,
					FeePerByte:              1,
					ArrivalTimestamp:        1562893305,
					TransactionBytes:        getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      2,
					FeePerByte:              10,
					ArrivalTimestamp:        1562893304,
					TransactionBytes:        getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      3,
					FeePerByte:              1,
					ArrivalTimestamp:        1562893302,
					TransactionBytes:        getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      4,
					FeePerByte:              100,
					ArrivalTimestamp:        1562893306,
					TransactionBytes:        getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      5,
					FeePerByte:              5,
					ArrivalTimestamp:        1562893303,
					TransactionBytes:        getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
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
		Chaintype          chaintype.ChainType
		QueryExecutor      query.ExecutorInterface
		MempoolQuery       query.MempoolQueryInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		Observer           *observer.Observer
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
				Observer:           observer.NewObserver(),
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
				Observer:           observer.NewObserver(),
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
				Observer:           tt.fields.Observer,
			}
			if err := mps.AddMempoolTransaction(tt.args.mpTx); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.AddMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMempoolService_SelectTransactionsFromMempool(t *testing.T) {
	type fields struct {
		Chaintype          chaintype.ChainType
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
					ID:                      4,
					FeePerByte:              100,
					ArrivalTimestamp:        1562893306,
					TransactionBytes:        getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      2,
					FeePerByte:              10,
					ArrivalTimestamp:        1562893304,
					TransactionBytes:        getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      5,
					FeePerByte:              5,
					ArrivalTimestamp:        1562893303,
					TransactionBytes:        getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      3,
					FeePerByte:              1,
					ArrivalTimestamp:        1562893302,
					TransactionBytes:        getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      1,
					FeePerByte:              1,
					ArrivalTimestamp:        1562893305,
					TransactionBytes:        getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
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

type (
	ReceivedTransactionListenerMockTypeAction struct {
		transaction.SendMoney
	}
	ReceivedTransactionListenerMockTypeActionSuccess struct {
		ReceivedTransactionListenerMockTypeAction
	}
)

// mockTypeAction
func (*ReceivedTransactionListenerMockTypeAction) ApplyConfirmed() error {
	return nil
}
func (*ReceivedTransactionListenerMockTypeAction) Validate() error {
	return nil
}
func (*ReceivedTransactionListenerMockTypeAction) GetAmount() int64 {
	return 10
}

func (*ReceivedTransactionListenerMockTypeAction) ApplyUnconfirmed() error {
	return nil
}

func (*ReceivedTransactionListenerMockTypeActionSuccess) GetTransactionType(tx *model.Transaction) transaction.TypeAction {
	return &ReceivedTransactionListenerMockTypeAction{}
}
