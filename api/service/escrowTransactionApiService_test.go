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
					ApproverAddress: []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
						213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
				},
			},
			want: &model.GetEscrowTransactionsResponse{
				Total: 1,
				Escrows: []*model.Escrow{
					{
						ID: 1,
						SenderAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
							72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
						RecipientAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
							202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						ApproverAddress: []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
							213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
						Amount:     10,
						Commission: 1,
						Timeout:    120,
						Status:     model.EscrowStatus_Approved,
						Latest:     true,
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
		[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
			72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		[]byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
			202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
		[]byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
			213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
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
				ID: 1,
				SenderAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
					72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				RecipientAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
					202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
				ApproverAddress: []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75,
					213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54},
				Amount:     10,
				Commission: 1,
				Timeout:    120,
				Status:     model.EscrowStatus_Approved,
				Latest:     true,
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
