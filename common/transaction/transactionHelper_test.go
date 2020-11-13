package transaction

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockTransactionQuery = query.NewTransactionQuery(chaintype.GetChainType(0))
	mockTransaction      = &model.Transaction{
		ID:      -1273123123,
		BlockID: -123123123123,
		Version: 1,
		Height:  1,
		SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		TransactionType:       binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
		Fee:                   1,
		Timestamp:             10000,
		TransactionHash:       make([]byte, 200),
		TransactionBodyLength: 88,
		TransactionBodyBytes:  make([]byte, 88),
		Signature:             make([]byte, 68),
		TransactionIndex:      1,
	}
)

type (
	mockExecuteTransactionError struct {
		query.ExecutorInterface
	}
	mockExecuteTransactionSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockExecuteTransactionError) ExecuteTransaction(query string, args ...interface{}) error {
	return errors.New("Error ExecuteTransaction")
}

func (*mockExecuteTransactionSuccess) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func TestTransactionHelper_InsertTransaction(t *testing.T) {
	type fields struct {
		TransactionQuery query.TransactionQueryInterface
		QueryExecutor    query.ExecutorInterface
	}
	type args struct {
		transaction *model.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "InsertTransaction:Error",
			args: args{
				transaction: mockTransaction,
			},
			fields: fields{
				TransactionQuery: mockTransactionQuery,
				QueryExecutor:    &mockExecuteTransactionError{},
			},
			wantErr: true,
		},
		{
			name: "InsertTransaction:Success",
			args: args{
				transaction: mockTransaction,
			},
			fields: fields{
				TransactionQuery: mockTransactionQuery,
				QueryExecutor:    &mockExecuteTransactionSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := &TransactionHelper{
				TransactionQuery: tt.fields.TransactionQuery,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			if err := th.InsertTransaction(tt.args.transaction); (err != nil) != tt.wantErr {
				t.Errorf("TransactionHelper.InsertTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTransactionHelper(t *testing.T) {
	type args struct {
		transactionQuery query.TransactionQueryInterface
		queryExecutor    query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *TransactionHelper
	}{
		{
			name: "NewTransactionHelper:Success",
			args: args{
				transactionQuery: mockTransactionQuery,
				queryExecutor:    &mockExecuteTransactionSuccess{},
			},
			want: &TransactionHelper{
				TransactionQuery: mockTransactionQuery,
				QueryExecutor:    &mockExecuteTransactionSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTransactionHelper(tt.args.transactionQuery, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransactionHelper() = %v, want %v", got, tt.want)
			}
		})
	}
}
