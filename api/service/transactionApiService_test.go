// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"reflect"
	"regexp"
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
	mockData := struct {
		TransactionSize   uint32
		TransactionHeight uint32
		Transactions      []*model.Transaction
	}{
		TransactionSize:   2,
		TransactionHeight: 0,
		Transactions: []*model.Transaction{
			{
				ID:                      1,
				PreviousTransactionHash: []byte{},
				Height:                  1,
				Timestamp:               10000,
				TransactionSeed:         []byte{},
				TransactionSignature:    []byte{},
				CumulativeDifficulty:    "",
				SmithScale:              1,
				PayloadLength:           2,
				PayloadHash:             []byte{},
				TransactionsmithID:      []byte{},
				TotalAmount:             0,
				TotalFee:                0,
				TotalCoinBase:           0,
				Version:                 1,
			},
			{
				ID:                      1,
				PreviousTransactionHash: []byte{},
				Height:                  2,
				Timestamp:               11000,
				TransactionSeed:         []byte{},
				TransactionSignature:    []byte{},
				CumulativeDifficulty:    "",
				SmithScale:              1,
				PayloadLength:           2,
				PayloadHash:             []byte{},
				TransactionsmithID:      []byte{},
				TotalAmount:             0,
				TotalFee:                0,
				TotalCoinBase:           0,
				Version:                 1,
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
		bs      *TransactionService
		want    *model.GetTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetTransactions:success",
			bs:   instance,
			want: &model.GetTransactionsResponse{
				Transactions:      mockData.Transactions,
				TransactionHeight: mockData.TransactionHeight,
				TransactionSize:   2,
			},
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	transactionQuery := query.NewTransactionQuery(chainType)
	queryStr := transactionQuery.GetTransactions(mockData.TransactionHeight, mockData.TransactionSize)

	mock.ExpectQuery(queryStr).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousTransactionHash", "Height", "Timestamp", "TransactionSeed", "TransactionSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "TransactionsmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version",
	}).AddRow(
		1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1).AddRow(
		1, []byte{}, 2, 11000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instance.GetTransactions(chainType, mockData.TransactionSize, mockData.TransactionHeight)
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

func Test_TransactionService_GetTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()
	var bl model.Transaction
	instance := NewTransactionService(query.NewQueryExecutor(db))
	defer ResetTransactionService()
	tests := []struct {
		name    string
		bs      *TransactionService
		want    *model.Transaction
		wantErr bool
	}{
		{
			name:    "GetTransactionByHeight:success",
			bs:      instance,
			want:    &bl,
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	transactionQuery := query.NewTransactionQuery(chainType)
	queryStr := transactionQuery.GetTransactionByHeight(0)
	mock.ExpectQuery(regexp.QuoteMeta(queryStr)).
		WillReturnRows(sqlmock.NewRows(transactionQuery.Fields))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instance.GetTransactionByHeight(chainType, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.GetTransactionByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.GetTransactionByHeight() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
