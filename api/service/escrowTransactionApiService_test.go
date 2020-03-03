package service

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockQueryExecutorGetEscrowTransactionsError struct {
		query.ExecutorInterface
	}
	mockQueryExecutorGetEscrowTransactionsSuccess struct {
		query.ExecutorInterface
	}
)

func (*mockQueryExecutorGetEscrowTransactionsError) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, sql.ErrNoRows
}
func (*mockQueryExecutorGetEscrowTransactionsError) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, sql.ErrNoRows
}

func (*mockQueryExecutorGetEscrowTransactionsSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	dbMocked, mock, _ := sqlmock.New()
	mockedRows := mock.NewRows(query.NewEscrowTransactionQuery().Fields)
	mockedRows.AddRow(
		int64(1),
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	)

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockedRows)
	return dbMocked.Query(qStr)

}

func (*mockQueryExecutorGetEscrowTransactionsSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockedRow := mock.NewRows([]string{"count"})
	mockedRow.AddRow(1)
	mock.ExpectQuery("").WillReturnRows(mockedRow)
	row := db.QueryRow("")
	return row, nil
}

func TestEscrowTransactionService_GetEscrowTransactions(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		params *model.GetEscrowTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetEscrowTransactionsResponse
		wantErr bool
	}{
		{
			name: "WantError",
			fields: fields{
				Query: &mockQueryExecutorGetEscrowTransactionsError{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockQueryExecutorGetEscrowTransactionsSuccess{},
			},
			args: args{
				params: &model.GetEscrowTransactionsRequest{
					ApproverAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				},
			},
			want: &model.GetEscrowTransactionsResponse{
				Total: 1,
				Escrows: []*model.Escrow{
					{
						ID:               1,
						SenderAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
						RecipientAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
						ApproverAddress:  "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
						Amount:           10,
						Commission:       1,
						Timeout:          120,
						Status:           model.EscrowStatus_Approved,
						Latest:           true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &escrowTransactionService{
				Query: tt.fields.Query,
			}
			got, err := es.GetEscrowTransactions(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEscrowTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEscrowTransactions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockExecutorGetEscrow struct {
		query.ExecutorInterface
	}
)

func (*mockExecutorGetEscrow) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockedRow := mock.NewRows(query.NewEscrowTransactionQuery().Fields)
	mockedRow.AddRow(
		int64(1),
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		"BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		int64(10),
		int64(1),
		uint64(120),
		model.EscrowStatus_Approved,
		uint32(0),
		true,
		"",
	)
	mock.ExpectQuery("").WillReturnRows(mockedRow)
	row := db.QueryRow("")
	return row, nil
}

func Test_escrowTransactionService_GetEscrowTransaction(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		request *model.GetEscrowTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Escrow
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Query: &mockExecutorGetEscrow{},
			},
			args: args{
				request: &model.GetEscrowTransactionRequest{
					ID: 918263123,
				},
			},
			want: &model.Escrow{
				ID:               1,
				SenderAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				RecipientAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
				ApproverAddress:  "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Amount:           10,
				Commission:       1,
				Timeout:          120,
				Status:           model.EscrowStatus_Approved,
				Latest:           true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &escrowTransactionService{
				Query: tt.fields.Query,
			}
			got, err := es.GetEscrowTransaction(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEscrowTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEscrowTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}
