// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"github.com/zoobc/zoobc-core/common/storage"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
)

// resetTransactionService resets the singleton back to nil, used in test case teardown
func resetTransactionService() {
	transactionServiceInstance = nil
}

type (
	mockTypeSwitcherValidateFail struct {
		transaction.TypeSwitcher
	}
	mockTxTypeValidateFail struct {
		transaction.TXEmpty
	}
	mockTypeSwitcherApplyUnconfirmedFail struct {
		transaction.TypeSwitcher
	}
	mockTxTypeApplyUnconfirmedFail struct {
		transaction.TXEmpty
	}
	mockTypeSwitcherSuccess struct {
		transaction.TypeSwitcher
	}
	mockTxTypeSuccess struct {
		transaction.TXEmpty
	}
	mockMempoolServiceFailAdd struct {
		service.MempoolService
	}
	mockMempoolServiceFailValidate struct {
		service.MempoolService
	}
	mockMempoolServiceSuccess struct {
		service.MempoolService
	}
	mockGetTransactionExecutorTxsFail struct {
		query.Executor
	}
	mockGetTransactionExecutorTxNoRow struct {
		query.Executor
	}

	mockTransactionExecutorFailBeginTx struct {
		query.Executor
	}
	mockTransactionExecutorSuccess struct {
		query.Executor
	}
	mockTransactionExecutorRollbackFail struct {
		mockTransactionExecutorSuccess
	}
	mockTransactionExecutorCommitFail struct {
		mockTransactionExecutorSuccess
	}
)

var (
	mockLog             = logrus.New()
	txAPISenderAccount1 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
		28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
	txAPIRecipientAccount1 = []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
		116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162}
)

func (*mockTypeSwitcherValidateFail) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &mockTxTypeValidateFail{}, nil
}

func (*mockTypeSwitcherApplyUnconfirmedFail) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &mockTxTypeApplyUnconfirmedFail{}, nil
}

func (*mockTypeSwitcherSuccess) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &mockTxTypeSuccess{}, nil
}

func (*mockTxTypeValidateFail) Validate(bool) error {
	return errors.New("mockError:validateFail")
}

func (*mockTxTypeApplyUnconfirmedFail) Validate(bool) error {
	return nil
}

func (*mockTxTypeSuccess) Validate(bool) error {
	return nil
}

func (*mockTxTypeApplyUnconfirmedFail) ApplyUnconfirmed() error {
	return errors.New("mockError:ApplyUnconfirmedFail")
}

func (*mockTxTypeSuccess) ApplyUnconfirmed() error {
	return nil
}

func (*mockMempoolServiceFailAdd) AddMempoolTransaction(tx *model.Transaction, txBytes []byte) error {
	return errors.New("mockError:addTxFail")
}

func (*mockMempoolServiceFailAdd) ValidateMempoolTransaction(mpTx *model.Transaction) error {
	return nil
}

func (*mockMempoolServiceFailValidate) ValidateMempoolTransaction(mpTx *model.Transaction) error {
	return errors.New("mockedError")
}

func (*mockMempoolServiceSuccess) AddMempoolTransaction(tx *model.Transaction, txBytes []byte) error {
	return nil
}

func (*mockMempoolServiceSuccess) ValidateMempoolTransaction(mpTx *model.Transaction) error {
	return nil
}

func (*mockGetTransactionExecutorTxsFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:getTxsFail")
}
func (*mockGetTransactionExecutorTxsFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(
		query.NewTransactionQuery(chaintype.GetChainType(0)).Fields,
	))
	return db.QueryRow(""), nil
}

func (*mockGetTransactionExecutorTxNoRow) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "BlockID", "Height", "SenderAccountType", "SenderAccountAddress", "RecipientAccountType", "RecipientAccountAddress",
		"TransactionType", "Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature",
		"Version"}))
	return db.Query(qe)
}
func (*mockGetTransactionExecutorTxNoRow) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(
		query.NewTransactionQuery(chaintype.GetChainType(0)).Fields,
	))
	return db.QueryRow(""), nil
}

func (*mockTransactionExecutorFailBeginTx) BeginTx() error {
	return errors.New("mockedError")
}

func (*mockTransactionExecutorSuccess) BeginTx() error {
	return nil
}

func (*mockTransactionExecutorSuccess) CommitTx() error {
	return nil
}

func (*mockTransactionExecutorSuccess) RollbackTx() error {
	return nil
}

func (*mockTransactionExecutorSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*mockTransactionExecutorRollbackFail) RollbackTx() error {
	return errors.New("mockedError")
}

func (*mockTransactionExecutorCommitFail) CommitTx() error {
	return errors.New("mockedError")
}

func TestNewTransactionService(t *testing.T) {
	transactionUtil := &transaction.Util{}
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
			want: &TransactionService{
				Query:           query.NewQueryExecutor(db),
				TransactionUtil: transactionUtil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTransactionService(
				query.NewQueryExecutor(db),
				nil,
				nil,
				nil,
				nil,
				transactionUtil,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransactionService() = %v, want %v", got, tt.want)
			}
			defer resetTransactionService()
		})
	}
}

type (
	mockQueryExecutorPostApprovalEscrowTX struct {
		query.Executor
	}
	mockMempoolServicePostApprovalEscrowTX struct {
		service.MempoolService
	}
	mockMempoolServicePostApprovalEscrowTXSuccess struct {
		service.MempoolService
	}
)

func (*mockQueryExecutorPostApprovalEscrowTX) BeginTx() error {
	return nil
}
func (*mockQueryExecutorPostApprovalEscrowTX) CommitTx() error {
	return nil
}
func (*mockQueryExecutorPostApprovalEscrowTX) RollbackTx() error {
	return nil
}
func (*mockQueryExecutorPostApprovalEscrowTX) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockMempoolServicePostApprovalEscrowTX) ValidateMempoolTransaction(mpTx *model.Transaction) error {
	return errors.New("test")
}
func (*mockMempoolServicePostApprovalEscrowTXSuccess) ValidateMempoolTransaction(mpTx *model.Transaction) error {
	return nil
}
func (*mockMempoolServicePostApprovalEscrowTXSuccess) AddMempoolTransaction(tx *model.Transaction, txBytes []byte) error {
	return nil
}

type (
	mockCacheStorageAlwaysSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockCacheStorageAlwaysSuccess) SetItem(key, item interface{}) error { return nil }
func (*mockCacheStorageAlwaysSuccess) GetItem(key, item interface{}) error { return nil }
func (*mockCacheStorageAlwaysSuccess) GetAllItems(item interface{}) error  { return nil }
func (*mockCacheStorageAlwaysSuccess) RemoveItem(key interface{}) error    { return nil }
func (*mockCacheStorageAlwaysSuccess) GetSize() int64                      { return 0 }
func (*mockCacheStorageAlwaysSuccess) ClearCache() error                   { return nil }

func TestTransactionService_PostTransaction(t *testing.T) {

	txTypeSuccess, transactionHashed := transaction.GetFixtureForSpecificTransaction(
		1655828751895385352,
		1562806389280,
		txAPISenderAccount1,
		txAPIRecipientAccount1,
		8,
		model.TransactionType_SendMoneyTransaction,
		&model.SendMoneyTransactionBody{
			Amount: 10,
		},
		false,
		true,
	)
	escrowApprovalTX, escrowApprovalTXBytes := transaction.GetFixtureForSpecificTransaction(
		8880850336851038037,
		1581301507,
		txAPISenderAccount1,
		nil,
		12,
		model.TransactionType_ApprovalEscrowTransaction,
		&model.ApprovalEscrowTransactionBody{
			Approval:      0,
			TransactionID: 0,
		},
		false,
		true,
	)

	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Log                *logrus.Logger
		Observer           *observer.Observer
		TransactionUtil    transaction.UtilInterface
	}
	type args struct {
		chaintype chaintype.ChainType
		req       *model.PostTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "PostTransaction:txBytesInvalid",
			fields: fields{
				Query: nil,
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.ValidateFail",
			fields: fields{
				Query:              nil,
				ActionTypeSwitcher: &mockTypeSwitcherValidateFail{},
				MempoolService:     &mockMempoolServiceFailValidate{},
				Log:                mockLog,
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:beginTxFail",
			fields: fields{
				Query:              &mockTransactionExecutorFailBeginTx{},
				ActionTypeSwitcher: &mockTypeSwitcherApplyUnconfirmedFail{},
				Log:                mockLog,
				MempoolService:     &mockMempoolServiceSuccess{},
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.ApplyUnconfirmedFail",
			fields: fields{
				Query:              &mockTransactionExecutorSuccess{},
				ActionTypeSwitcher: &mockTypeSwitcherApplyUnconfirmedFail{},
				Log:                mockLog,
				MempoolService:     &mockMempoolServiceSuccess{},
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.ApplyUnconfirmedFail-RollbackFail",
			fields: fields{
				Query:              &mockTransactionExecutorRollbackFail{},
				ActionTypeSwitcher: &mockTypeSwitcherApplyUnconfirmedFail{},
				Log:                mockLog,
				MempoolService:     &mockMempoolServiceSuccess{},
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.AddMempoolTransactionFail",
			fields: fields{
				Query:              &mockTransactionExecutorSuccess{},
				ActionTypeSwitcher: &mockTypeSwitcherSuccess{},
				MempoolService:     &mockMempoolServiceFailAdd{},
				Log:                mockLog,
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.AddMempoolTransactionFail-RollbackFail",
			fields: fields{
				Query:              &mockTransactionExecutorRollbackFail{},
				ActionTypeSwitcher: &mockTypeSwitcherSuccess{},
				MempoolService:     &mockMempoolServiceFailAdd{},
				Log:                mockLog,
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.AddMempoolTransactionFail-RollbackFail",
			fields: fields{
				Query:              &mockTransactionExecutorCommitFail{},
				ActionTypeSwitcher: &mockTypeSwitcherSuccess{},
				MempoolService:     &mockMempoolServiceSuccess{},
				Log:                mockLog,
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50,
						83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
						57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106,
						116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122,
						68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 4, 38, 103, 73, 250, 169, 63,
						155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11,
						4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
						77, 80, 80, 39, 254, 173, 28, 169},
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "PostTransaction:txType.Success",
			fields: fields{
				Query:              &mockTransactionExecutorSuccess{},
				ActionTypeSwitcher: &mockTypeSwitcherSuccess{},
				MempoolService:     &mockMempoolServiceSuccess{},
				Observer:           observer.NewObserver(),
				Log:                mockLog,
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: transactionHashed,
				},
			},
			wantErr: false,
			want:    txTypeSuccess,
		},
		{
			name: "WantError:ValidateMempoolFail1",
			fields: fields{
				Query:     &mockQueryExecutorPostApprovalEscrowTX{},
				Signature: nil,
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					Executor: &mockQueryExecutorPostApprovalEscrowTX{},
				},
				MempoolService: &mockMempoolServicePostApprovalEscrowTXSuccess{},
				Observer:       observer.NewObserver(),
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: escrowApprovalTXBytes,
				},
			},
			want: escrowApprovalTX,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TransactionService{
				Query:              tt.fields.Query,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				MempoolService:     tt.fields.MempoolService,
				Observer:           tt.fields.Observer,
				TransactionUtil:    tt.fields.TransactionUtil,
			}
			got, err := ts.PostTransaction(tt.args.chaintype, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.PostTransaction() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.PostTransaction() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetTransactionsFail struct {
		query.Executor
	}
	mockQueryGetTransactionsSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetTransactionsFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("want error")
}
func (*mockQueryGetTransactionsFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("want error")
}
func (*mockQueryGetTransactionsSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows(query.NewTransactionQuery(&chaintype.MainChain{}).Fields).
			AddRow(
				4545420970999433273,
				1,
				1,
				txAPISenderAccount1,
				txAPIRecipientAccount1,
				1,
				1,
				10000,
				[]byte{1, 1},
				8,
				[]byte{1, 2, 3, 4, 5, 6, 7, 8},
				[]byte{0, 0, 0, 0, 0, 0, 0},
				1,
				1,
				false,
			),
		)
	return db.Query("")
}
func (*mockQueryGetTransactionsSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total_record"}).AddRow(1))
	return db.QueryRow(""), nil
}
func TestTransactionService_GetTransactions(t *testing.T) {
	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		params    *model.GetTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetTransactionsResponse
		wantErr bool
	}{
		{
			name: "RequestFilledExecuteSelectGetTxsFail",
			fields: fields{
				Query: &mockQueryGetTransactionsFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Pagination: &model.Pagination{
						Limit: 2,
						Page:  0,
					},
					AccountAddress: txAPISenderAccount1,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "RequestExceptAccountAddressExecuteSelectGetTxsFail",
			fields: fields{
				Query: &mockQueryGetTransactionsFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Pagination: &model.Pagination{
						Limit: 2,
						Page:  0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "RequestSuccess",
			fields: fields{
				Query:              &mockQueryGetTransactionsSuccess{},
				ActionTypeSwitcher: &mockTypeSwitcherSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Pagination: &model.Pagination{
						Limit: 1,
						Page:  0,
					},
					AccountAddress: txAPISenderAccount1,
					Height:         1,
				},
			},
			want: &model.GetTransactionsResponse{
				Total: 1,
				Transactions: []*model.Transaction{
					{
						ID:                      4545420970999433273,
						BlockID:                 1,
						Height:                  1,
						SenderAccountAddress:    txAPISenderAccount1,
						RecipientAccountAddress: txAPIRecipientAccount1,
						TransactionType:         1,
						Fee:                     1,
						Timestamp:               10000,
						TransactionHash:         []byte{1, 1},
						TransactionBodyLength:   8,
						TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Signature:               []byte{0, 0, 0, 0, 0, 0, 0},
						Version:                 1,
						TransactionIndex:        1,
						MultisigChild:           false,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TransactionService{
				Query:              tt.fields.Query,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				MempoolService:     tt.fields.MempoolService,
				TransactionUtil:    &transaction.Util{},
			}
			got, err := ts.GetTransactions(tt.args.chainType, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.GetTransactions() error = \n %v, wantErr = \n %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.GetTransactions() got = \n %v, want = \n %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetTransactionSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetTransactionSuccess) ExecuteSelect(
	qe string,
	tx bool,
	args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewTransactionQuery(&chaintype.MainChain{}).Fields).AddRow(
			4545420970999433273,
			1,
			1,
			txAPISenderAccount1,
			txAPIRecipientAccount1,
			0,
			1,
			10000,
			[]byte{1, 1},
			8,
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1, 1,
			false,
		),
	)
	return db.Query("")
}
func (*mockQueryGetTransactionSuccess) ExecuteSelectRow(qstr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewTransactionQuery(chaintype.GetChainType(0)).Fields).
			AddRow(
				4545420970999433273,
				1,
				1,
				txAPISenderAccount1,
				txAPIRecipientAccount1,
				0,
				1,
				10000,
				[]byte{1, 1},
				8,
				[]byte{1, 2, 3, 4, 5, 6, 7, 8},
				[]byte{0, 0, 0, 0, 0, 0, 0}, 1, 1,
				false,
			),
	)
	return db.QueryRow(""), nil
}
func TestTransactionService_GetTransaction(t *testing.T) {
	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		params    *model.GetTransactionRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "GetTransaction:failExecuteSelect",
			fields: fields{
				Query: &mockGetTransactionExecutorTxsFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionRequest{
					ID: 1,
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetTransaction:noRowExecuteSelect",
			fields: fields{
				Query: &mockGetTransactionExecutorTxNoRow{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionRequest{
					ID: 1,
				},
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetTransaction:success",
			fields: fields{
				Query:              &mockQueryGetTransactionSuccess{},
				ActionTypeSwitcher: &mockTypeSwitcherSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionRequest{
					ID: 1,
				},
			},
			wantErr: false,
			want: &model.Transaction{
				ID:                      4545420970999433273,
				BlockID:                 1,
				Height:                  1,
				SenderAccountAddress:    txAPISenderAccount1,
				RecipientAccountAddress: txAPIRecipientAccount1,
				TransactionType:         0,
				Fee:                     1,
				Timestamp:               10000,
				TransactionHash:         []byte{1, 1},
				TransactionBodyLength:   8,
				TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Signature:               []byte{0, 0, 0, 0, 0, 0, 0},
				Version:                 1,
				TransactionIndex:        1,
				MultisigChild:           false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TransactionService{
				Query:              tt.fields.Query,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				MempoolService:     tt.fields.MempoolService,
				TransactionUtil:    &transaction.Util{},
			}
			got, err := ts.GetTransaction(tt.args.chainType, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.GetTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.GetTransaction() got = \n%v, want = \n%v", got, tt.want)
			}
		})
	}
}
