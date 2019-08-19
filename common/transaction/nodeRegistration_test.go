package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// validate mock
	mockExecutorValidateFailExecuteSelectFail struct {
		query.Executor
	}
	mockExecutorValidateFailBalanceNotEnough struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectNodeFail struct {
		query.Executor
	}
	mockExecutorValidateFailExecuteSelectNodeExist struct {
		query.Executor
	}
	mockExecutorValidateSuccess struct {
		query.Executor
	}

	// undo unconfirmed mock
	mockExecutorUndoUnconfirmedExecuteTransactionsFail struct {
		query.Executor
	}

	mockExecutorUndoUnconfirmedSuccess struct {
		query.Executor
	}

	// apply unconfirmed mock
	mockExecutorApplyUnconfirmedExecuteTransactionFail struct {
		mockExecutorValidateSuccess
	}
	mockExecutorApplyUnconfirmedSuccess struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedUndoUnconfirmedFail struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedExecuteTransactionsFail struct {
		mockExecutorValidateSuccess
	}
	mockApplyConfirmedSuccess struct {
		mockExecutorValidateSuccess
	}
)

func (*mockExecutorValidateFailExecuteSelectFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:selectFail")
}

func (*mockExecutorValidateFailBalanceNotEnough) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
		"AccountID",
		"BlockHeight",
		"SpendableBalance",
		"Balance",
		"PopRevenue",
		"Latest",
	}).AddRow(
		[]byte{1},
		1,
		10,
		10,
		0,
		true,
	),
	)
	return db.Query("")
}

func (*mockExecutorValidateFailExecuteSelectNodeFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_id,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_id = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountID",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			[]byte{1},
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	return nil, errors.New("mockError:nodeFail")
}

func (*mockExecutorValidateFailExecuteSelectNodeExist) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_id,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_id = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountID",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			[]byte{1},
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	mock.ExpectQuery("B").WillReturnRows(sqlmock.NewRows([]string{
		"NodePublicKey",
		"AccountId",
		"RegistrationHeight",
		"NodeAddress",
		"LockedBalance",
		"Queued",
		"Latest",
		"Height",
	}).AddRow(
		[]byte{1},
		[]byte{2},
		1,
		"127.0.0.1",
		1000000,
		true,
		true,
		1,
	))
	return db.Query("B")
}

func (*mockExecutorValidateSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	if qe == "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance WHERE "+
		"account_address = ? AND latest = 1" {
		mock.ExpectQuery("A").WillReturnRows(sqlmock.NewRows([]string{
			"AccountAddress",
			"BlockHeight",
			"SpendableBalance",
			"Balance",
			"PopRevenue",
			"Latest",
		}).AddRow(
			"BCZ",
			1,
			1000000,
			1000000,
			0,
			true,
		))
		return db.Query("A")
	}
	mock.ExpectQuery("B").WillReturnRows(sqlmock.NewRows([]string{
		"NodePublicKey",
		"AccountAddress",
		"RegistrationHeight",
		"NodeAddress",
		"LockedBalance",
		"Queued",
		"Latest",
		"Height",
	}))
	return db.Query("B")
}

func (*mockExecutorUndoUnconfirmedExecuteTransactionsFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:undoFail")
}

func (*mockExecutorUndoUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockExecutorApplyUnconfirmedExecuteTransactionFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:ApplyUnconfirmedFail")
}

func (*mockExecutorApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockApplyConfirmedUndoUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockUndoUnconfirmedFail")
}

func (*mockApplyConfirmedExecuteTransactionsFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockError:ExecuteTransactionsFail")
}

func (*mockApplyConfirmedSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestNodeRegistration_ApplyConfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "ApplyConfirmed:fail-{undoUnconfirmedFail}",
			wantErr: true,
			fields: fields{
				Height:                1,
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockApplyConfirmedUndoUnconfirmedFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
		{
			name:    "ApplyConfirmed:fail-{executeTransactionsFail}",
			wantErr: true,
			fields: fields{
				Height:                0,
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockApplyConfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
		{
			name:    "ApplyConfirmed:success",
			wantErr: false,
			fields: fields{
				Height:                0,
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockApplyConfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.ApplyConfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "ApplyUnconfirmed:fail-{ExecuteTransactionFail}",
			wantErr: true,
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorApplyUnconfirmedExecuteTransactionFail{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10,
					NodePublicKey: []byte{1},
				},
				Fee: 1,
			},
		},
		{
			name:    "ApplyUnconfirmed:success",
			wantErr: false,
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorApplyUnconfirmedSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10,
					NodePublicKey: []byte{1},
				},
				Fee: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:fail-{executeTransactionsFail}",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorUndoUnconfirmedExecuteTransactionsFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorUndoUnconfirmedSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_Validate(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Validate:fail-{executeSelectFail}",
			fields: fields{
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:       &mockExecutorValidateFailExecuteSelectFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{balanceNotEnough}",
			fields: fields{
				SenderAddress:       "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:       &mockExecutorValidateFailBalanceNotEnough{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				Fee:                 1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{failGetNode}",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectNodeFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{nodeExist}",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorValidateFailExecuteSelectNodeExist{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "Validate:success",
			fields: fields{
				SenderAddress:         "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				QueryExecutor:         &mockExecutorValidateSuccess{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				Fee:                   1,
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 10000,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if err := tx.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistration_GetAmount(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			fields: fields{
				Body: &model.NodeRegistrationTransactionBody{
					LockedBalance: 1000,
				},
			},
			want: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("NodeRegistration.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_GetSize(t *testing.T) {
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSize:success",
			fields: fields{
				Body: &model.NodeRegistrationTransactionBody{
					NodeAddress: "127.0.0.1",
				},
			},
			want: 277,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := n.GetSize(); got != tt.want {
				t.Errorf("NodeRegistration.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_ParseBodyBytes(t *testing.T) {
	_, _, body, bodyBytes := GetFixturesForNoderegistration()
	// bodyBytes :=
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.NodeRegistrationTransactionBody
	}{
		{
			name:   "ParseBodyBytes:success",
			fields: fields{},
			args: args{
				txBodyBytes: bodyBytes,
			},
			want: body,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				Height:                tt.fields.Height,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := n.ParseBodyBytes(tt.args.txBodyBytes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistration.ParseBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeRegistration_GetBodyBytes(t *testing.T) {
	_, _, body, bodyBytes := GetFixturesForNoderegistration()
	type fields struct {
		Body                  *model.NodeRegistrationTransactionBody
		Fee                   int64
		SenderAddress         string
		Height                uint32
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		txBody *model.NodeRegistrationTransactionBody
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body: body,
			},
			args: args{
				txBody: body,
			},
			want: bodyBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeRegistration{
				Body:                  tt.fields.Body,
				Fee:                   tt.fields.Fee,
				SenderAddress:         tt.fields.SenderAddress,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			if got := n.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistration.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
