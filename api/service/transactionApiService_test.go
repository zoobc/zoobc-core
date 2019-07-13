// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

// ResetTransactionService resets the singleton back to nil, used in test case teardown
func ResetTransactionService() {
	transactionServiceInstance = nil
}

func TestNewTransactionervice(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	tests := []struct {
		name string
		want *TransactionService
	}{
		{
			name: "NewTransactionService:InitiateTransactionServiceInstance",
			want: &TransactionService{Query: query.NewQueryExecutor(db)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTransactionService(query.NewQueryExecutor(db)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransactionService() = %v, want %v", got, tt.want)
			}
			defer ResetTransactionService()
		})
	}
}

func Test_TransactionService_GetTransactions(t *testing.T) {
	params := &model.GetTransactionsRequest{
		Limit:  2,
		Offset: 0,
	}

	mockData := struct {
		Total        uint64
		Count        uint32
		Transactions []*model.Transaction
	}{
		Total: 2,
		Count: 0,
		Transactions: []*model.Transaction{
			{
				ID:                    1,
				BlockID:               1,
				Height:                1,
				TransactionType:       1,
				Fee:                   1,
				Timestamp:             11000,
				TransactionHash:       []byte{},
				TransactionBodyLength: 1,
				TransactionBodyBytes:  []byte{},
				Signature:             []byte{},
			},
			{
				ID:                    2,
				BlockID:               2,
				Height:                2,
				TransactionType:       2,
				Fee:                   2,
				Timestamp:             21000,
				TransactionHash:       []byte{},
				TransactionBodyLength: 2,
				TransactionBodyBytes:  []byte{},
				Signature:             []byte{},
			},
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()
	instance := NewTransactionService(query.NewQueryExecutor(db))
	defer ResetTransactionService()
	tests := []struct {
		name    string
		want    *model.GetTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetTransactions:success",
			want: &model.GetTransactionsResponse{
				Transactions: mockData.Transactions,
				Total:        mockData.Total,
				Count:        mockData.Count,
			},
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	transactionQuery := query.NewTransactionQuery(chainType)
	queryStr := transactionQuery.GetTransactions(params.Limit, params.Offset, 0, 0)

	mock.ExpectQuery(queryStr).WillReturnRows(sqlmock.NewRows([]string{
		"id",
		"block_id",
		"block_height",
		"sender_account_id",
		"recipient_account_id",
		"transaction_type",
		"fee",
		"timestamp",
		"transaction_hash",
		"transaction_body_length",
		"transaction_body_bytes",
		"signature",
	}).AddRow(
		1, 1, 1, []byte{}, []byte{}, 1, 1, 11000, []byte{}, 1, []byte{}, []byte{}).AddRow(
		2, 2, 2, []byte{}, []byte{}, 2, 2, 21000, []byte{}, 2, []byte{}, []byte{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instance.GetTransactions(chainType, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.GetTransactions() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
