package transaction

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"golang.org/x/crypto/sha3"
)

type (
	mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperFailed struct {
		AccountBalanceHelperInterface
	}
	mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperSuccess struct {
		AccountBalanceHelperInterface
	}
	mockCreateBlockchainObjectTransactionApplyConfirmedQueryExecutorFailed struct {
		query.Executor
	}
	mockCreateBlockchainObjectTransactionApplyConfirmedQueryExecutorSuccess struct {
		query.Executor
	}
)

var (
	mockBlockchainObjectTransactionBody = model.CreateBlockchainObjectTransactionBody{
		BlockchainObjectBalance: 1,
		BlockchainObjectImmutableProperties: map[string]string{
			"mockKey":  "mockValue",
			"mockKey2": "mockValue2",
		},
		BlockchainObjectMutableProperties: map[string]string{
			"mockKey":  "mockValue",
			"mockKey2": "mockValue2",
		},
	}
)

func (*mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperFailed) AddAccountBalance(
	address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64,
	blockTimestamp uint64,
) error {
	return errors.New("mockedErr")
}

func (*mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperSuccess) AddAccountBalance(
	address []byte, amount int64, event model.EventType, blockHeight uint32, transactionID int64,
	blockTimestamp uint64,
) error {
	return nil
}

func (*mockCreateBlockchainObjectTransactionApplyConfirmedQueryExecutorFailed) ExecuteTransaction(
	query string, args ...interface{}) error {
	return errors.New("mockedErr")
}

func (*mockCreateBlockchainObjectTransactionApplyConfirmedQueryExecutorSuccess) ExecuteTransaction(
	query string, args ...interface{}) error {
	return nil
}

func TestCreateBlockchainObjectTransaction_ApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	type args struct {
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantFail-AddAccountBalance",
			fields: fields{
				AccountBalanceHelper: &mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperFailed{},
				Body:                 &mockBlockchainObjectTransactionBody,
				Fee:                  1,
			},
			args: args{
				blockTimestamp: 0,
			},
			wantErr: true,
		},
		{
			name: "wantFail-InsertBlockcahinObject",
			fields: fields{
				AccountBalanceHelper:  &mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperSuccess{},
				Body:                  &mockBlockchainObjectTransactionBody,
				Fee:                   1,
				TransactionHash:       make([]byte, sha3.New256().Size()),
				BlockchainObjectQuery: query.NewBlockchainObjectQuery(),
				QueryExecutor:         &mockCreateBlockchainObjectTransactionApplyConfirmedQueryExecutorFailed{},
			},
			args: args{
				blockTimestamp: 0,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				AccountBalanceHelper:          &mockCreateBlockchainObjectTransactionApplyConfirmedAccountBalanceHelperSuccess{},
				Body:                          &mockBlockchainObjectTransactionBody,
				Fee:                           1,
				TransactionHash:               make([]byte, sha3.New256().Size()),
				BlockchainObjectQuery:         query.NewBlockchainObjectQuery(),
				BlockchainObjectPropertyQuery: query.NewBlockchainObjectPropertyQuery(),
				AccountDatasetQuery:           query.NewAccountDatasetsQuery(),
				QueryExecutor:                 &mockCreateBlockchainObjectTransactionApplyConfirmedQueryExecutorSuccess{},
			},
			args: args{
				blockTimestamp: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("CreateBlockchainObjectTransaction.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockCreateBlockchainObjectTransactionApplyUnconfirmedAccountBalanceHelperFailed struct {
		AccountBalanceHelperInterface
	}
	mockCreateBlockchainObjectTransactionApplyUnconfirmedAccountBalanceHelperSuccess struct {
		AccountBalanceHelperInterface
	}
)

func (*mockCreateBlockchainObjectTransactionApplyUnconfirmedAccountBalanceHelperFailed) AddAccountSpendableBalance(
	address []byte, amount int64,
) error {
	return errors.New("mockedErr")
}
func (*mockCreateBlockchainObjectTransactionApplyUnconfirmedAccountBalanceHelperSuccess) AddAccountSpendableBalance(
	address []byte, amount int64,
) error {
	return nil
}

func TestCreateBlockchainObjectTransaction_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail-AddAccountSpendableBalance",
			fields: fields{
				SenderAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				Body:                 &mockBlockchainObjectTransactionBody,
				Fee:                  1,
				AccountBalanceHelper: &mockCreateBlockchainObjectTransactionApplyUnconfirmedAccountBalanceHelperFailed{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				SenderAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				Body:                 &mockBlockchainObjectTransactionBody,
				Fee:                  1,
				AccountBalanceHelper: &mockCreateBlockchainObjectTransactionApplyUnconfirmedAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("CreateBlockchainObjectTransaction.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockCreateBlockchainObjectTransactionUndoApplyUnconfirmedAccountBalanceHelperFailed struct {
		AccountBalanceHelperInterface
	}
	mockCreateBlockchainObjectTransactionUndoApplyUnconfirmedAccountBalanceHelperSuccess struct {
		AccountBalanceHelperInterface
	}
)

func (*mockCreateBlockchainObjectTransactionUndoApplyUnconfirmedAccountBalanceHelperFailed) AddAccountSpendableBalance(
	address []byte, amount int64,
) error {
	return errors.New("mockedErr")
}
func (*mockCreateBlockchainObjectTransactionUndoApplyUnconfirmedAccountBalanceHelperSuccess) AddAccountSpendableBalance(
	address []byte, amount int64,
) error {
	return nil
}

func TestCreateBlockchainObjectTransaction_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail-AddAccountSpendableBalance",
			fields: fields{
				SenderAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				Body:                 &mockBlockchainObjectTransactionBody,
				Fee:                  1,
				AccountBalanceHelper: &mockCreateBlockchainObjectTransactionUndoApplyUnconfirmedAccountBalanceHelperFailed{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				SenderAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				Body:                 &mockBlockchainObjectTransactionBody,
				Fee:                  1,
				AccountBalanceHelper: &mockCreateBlockchainObjectTransactionUndoApplyUnconfirmedAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("CreateBlockchainObjectTransaction.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockCreateBlockchainObjectTransactionValidateQueryExecutorFailed struct {
		query.Executor
	}
	mockCreateBlockchainObjectTransactionValidateQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockCreateBlockchainObjectTransactionValidateQueryExecutorFailed) ExecuteSelectRow(
	query string, tx bool, args ...interface{},
) (*sql.Row, error) {
	return nil, errors.New("mockedErr")
}

func (*mockCreateBlockchainObjectTransactionValidateQueryExecutorSuccess) ExecuteSelectRow(
	string, bool, ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockRow := mock.NewRows(query.NewBlockchainObjectQuery().Fields)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	return db.QueryRow(""), nil
}

func TestCreateBlockchainObjectTransaction_Validate(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	type args struct {
		dbTx bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantFail-BlockchainObjectBalance0",
			fields: fields{
				Body: &model.CreateBlockchainObjectTransactionBody{},
			},
			args:    args{false},
			wantErr: true,
		},
		{
			name: "wantFail-InvalidTransactionHash",
			fields: fields{
				Body: &model.CreateBlockchainObjectTransactionBody{
					BlockchainObjectBalance: 1,
				},
				TransactionHash: make([]byte, 1),
			},
			args:    args{false},
			wantErr: true,
		},
		{
			name: "wantFail-ZeroImmutableProperty",
			fields: fields{
				Body: &model.CreateBlockchainObjectTransactionBody{
					BlockchainObjectBalance: 1,
				},
				TransactionHash: make([]byte, sha3.New256().Size()),
			},
			args:    args{false},
			wantErr: true,
		},
		{
			name: "wantFail-ExecuteSelectRow",
			fields: fields{
				Body:                  &mockBlockchainObjectTransactionBody,
				TransactionHash:       make([]byte, sha3.New256().Size()),
				BlockchainObjectQuery: query.NewBlockchainObjectQuery(),
				QueryExecutor:         &mockCreateBlockchainObjectTransactionValidateQueryExecutorFailed{},
			},
			args:    args{false},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Body:                  &mockBlockchainObjectTransactionBody,
				TransactionHash:       make([]byte, sha3.New256().Size()),
				BlockchainObjectQuery: query.NewBlockchainObjectQuery(),
				QueryExecutor:         &mockCreateBlockchainObjectTransactionValidateQueryExecutorSuccess{},
			},
			args:    args{false},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("CreateBlockchainObjectTransaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateBlockchainObjectTransaction_GetAmount(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: &mockBlockchainObjectTransactionBody,
			},
			want: mockBlockchainObjectTransactionBody.BlockchainObjectBalance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("CreateBlockchainObjectTransaction.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateBlockchainObjectTransaction_GetSize(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    uint32
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: &mockBlockchainObjectTransactionBody,
			},
			want:    116,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			got, err := tx.GetSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBlockchainObjectTransaction.GetSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateBlockchainObjectTransaction.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateBlockchainObjectTransaction_GetTransactionBody(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	type args struct {
		transaction *model.Transaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: &mockBlockchainObjectTransactionBody,
			},
			args: args{
				transaction: &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestCreateBlockchainObjectTransaction_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		ID                            int64
		Fee                           int64
		SenderAddress                 []byte
		Height                        uint32
		TransactionHash               []byte
		Body                          *model.CreateBlockchainObjectTransactionBody
		Escrow                        *model.Escrow
		AccountBalanceHelper          AccountBalanceHelperInterface
		QueryExecutor                 query.ExecutorInterface
		EscrowQuery                   query.EscrowTransactionQueryInterface
		BlockchainObjectQuery         query.BlockchainObjectQueryInterface
		BlockchainObjectPropertyQuery query.BlockchainObjectPropertyQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowFee                     fee.FeeModelInterface
		NormalFee                     fee.FeeModelInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
		blockTimestamp       int64
		blockHeight          uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "wantSuccess",
			fields:  fields{},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &CreateBlockchainObjectTransaction{
				ID:                            tt.fields.ID,
				Fee:                           tt.fields.Fee,
				SenderAddress:                 tt.fields.SenderAddress,
				Height:                        tt.fields.Height,
				TransactionHash:               tt.fields.TransactionHash,
				Body:                          tt.fields.Body,
				Escrow:                        tt.fields.Escrow,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				QueryExecutor:                 tt.fields.QueryExecutor,
				EscrowQuery:                   tt.fields.EscrowQuery,
				BlockchainObjectQuery:         tt.fields.BlockchainObjectQuery,
				BlockchainObjectPropertyQuery: tt.fields.BlockchainObjectPropertyQuery,
				AccountDatasetQuery:           tt.fields.AccountDatasetQuery,
				EscrowFee:                     tt.fields.EscrowFee,
				NormalFee:                     tt.fields.NormalFee,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions, tt.args.blockTimestamp, tt.args.blockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBlockchainObjectTransaction.SkipMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateBlockchainObjectTransaction.SkipMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
