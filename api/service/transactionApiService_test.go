// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
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
	mockGetTransactionExecutorTotalFail struct {
		query.Executor
	}
	mockGetTransactionExecutorTotalFailRow struct {
		query.Executor
	}
	mockGetTransactionExecutorSuccess struct {
		query.Executor
	}
	mockGetTransactionExecutorTxNoRow struct {
		query.Executor
	}
	mockGetTransactionExecutorTxSuccess struct {
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

var mockLog = logrus.New()

func (*mockTypeSwitcherValidateFail) GetTransactionType(tx *model.Transaction) transaction.TypeAction {
	return &mockTxTypeValidateFail{}
}

func (*mockTypeSwitcherApplyUnconfirmedFail) GetTransactionType(tx *model.Transaction) transaction.TypeAction {
	return &mockTxTypeApplyUnconfirmedFail{}
}

func (*mockTypeSwitcherSuccess) GetTransactionType(tx *model.Transaction) transaction.TypeAction {
	return &mockTxTypeSuccess{}
}

func (*mockTxTypeValidateFail) Validate() error {
	return errors.New("mockError:validateFail")
}

func (*mockTxTypeApplyUnconfirmedFail) Validate() error {
	return nil
}

func (*mockTxTypeSuccess) Validate() error {
	return nil
}

func (*mockTxTypeApplyUnconfirmedFail) ApplyUnconfirmed() error {
	return errors.New("mockError:ApplyUnconfirmedFail")
}

func (*mockTxTypeSuccess) ApplyUnconfirmed() error {
	return nil
}

func (*mockMempoolServiceFailAdd) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return errors.New("mockError:addTxFail")
}

func (*mockMempoolServiceFailAdd) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return nil
}

func (*mockMempoolServiceFailValidate) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return errors.New("mockedError")
}

func (*mockMempoolServiceSuccess) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return nil
}

func (*mockMempoolServiceSuccess) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return nil
}

func (*mockGetTransactionExecutorTxsFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:getTxsFail")
}

func (*mockGetTransactionExecutorTotalFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_id, block_height, sender_account_type, sender_account_address, recipient_account_type, " +
		"recipient_account_address, transaction_type, fee, timestamp, transaction_hash, transaction_body_length, " +
		"transaction_body_bytes, signature, version from \"transaction\" ORDER BY block_height, timestamp LIMIT 0,2":
		mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockID", "Height", "SenderAccountType", "SenderAccountAddress", "RecipientAccountType", "RecipientAccountAddress",
			"TransactionType", "Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature",
			"Version",
		}).AddRow(4545420970999433273, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1,
		).AddRow(
			4545420970999433274, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1))
	default:
		return nil, errors.New("mockError:totalFail")
	}
	return db.Query(qe)
}

func (*mockGetTransactionExecutorTxNoRow) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "BlockID", "Height", "SenderAccountType", "SenderAccountAddress", "RecipientAccountType", "RecipientAccountAddress",
		"TransactionType", "Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature",
		"Version"}))
	return db.Query(qe)
}

func (*mockGetTransactionExecutorTxSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "BlockID", "Height", "SenderAccountType", "SenderAccountAddress", "RecipientAccountType", "RecipientAccountAddress",
		"TransactionType", "Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature",
		"Version",
	}).AddRow(4545420970999433273, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
		[]byte{0, 0, 0, 0, 0, 0, 0}, 1,
	))
	return db.Query(qe)
}

func (*mockGetTransactionExecutorTotalFailRow) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_id, block_height, sender_account_type, sender_account_address, recipient_account_type, " +
		"recipient_account_address, transaction_type, fee, timestamp, transaction_hash, transaction_body_length, " +
		"transaction_body_bytes, signature, version from \"transaction\" ORDER BY block_height, timestamp LIMIT 0,2":
		mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockID", "Height", "SenderAccountType", "SenderAccountAddress", "RecipientAccountType", "RecipientAccountAddress",
			"TransactionType", "Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature", "Version",
		}).AddRow(4545420970999433273, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1,
		).AddRow(
			4545420970999433274, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1))
	default:
		mock.ExpectQuery("wrongRow").WillReturnRows(sqlmock.NewRows([]string{
			"total_record",
		}).AddRow("abc"))
		return db.Query("wrongRow")
	}
	return db.Query(qe)
}

func (*mockGetTransactionExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, block_id, block_height, sender_account_type, sender_account_address, recipient_account_type, " +
		"recipient_account_address, transaction_type, fee, timestamp, transaction_hash, transaction_body_length, " +
		"transaction_body_bytes, signature, version from \"transaction\" ORDER BY block_height, timestamp LIMIT 0,2":
		mock.ExpectQuery(qe).WillReturnRows(sqlmock.NewRows([]string{
			"ID", "BlockID", "Height", "SenderAccountType", "SenderAccountAddress", "RecipientAccountType", "RecipientAccountAddress",
			"TransactionType", "Fee", "Timestamp", "TransactionHash", "TransactionBodyLength", "TransactionBodyBytes", "Signature",
			"Version",
		}).AddRow(4545420970999433273, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1,
		).AddRow(
			4545420970999433274, 1, 1, 0, "senderA", 0, "recipientA", 1, 1, 10000, []byte{1, 1}, 8, []byte{1, 2, 3, 4, 5, 6, 7, 8},
			[]byte{0, 0, 0, 0, 0, 0, 0}, 1))
	default:
		mock.ExpectQuery("total-success").WillReturnRows(sqlmock.NewRows([]string{
			"total_record",
		}).AddRow(2))
		return db.Query("total-success")
	}
	return db.Query(qe)
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
			if got := NewTransactionService(query.NewQueryExecutor(db),
				nil, nil, nil, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransactionService() = %v, want %v", got, tt.want)
			}
			defer resetTransactionService()
		})
	}
}

func TestTransactionService_PostTransaction(t *testing.T) {
	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Log                *logrus.Logger
	}
	type args struct {
		chaintype contract.ChainType
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
				Log:                mockLog,
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
			wantErr: false,
			want: &model.Transaction{
				ID:                      -4430102867797816008,
				Version:                 1,
				TransactionType:         1,
				Timestamp:               1564112693,
				SenderAccountType:       0,
				SenderAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
				RecipientAccountType:    0,
				RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				Fee:                     1,
				TransactionBodyLength:   8,
				TransactionHash: []byte{56, 205, 120, 214, 233, 28, 133, 194, 224, 240, 192, 247, 227, 35, 183, 63, 118, 111, 147,
					55, 104, 54, 76, 108, 224, 194, 232, 88, 36, 93, 104, 76},
				Signature: []byte{32, 85, 34, 198, 89, 78, 166, 142, 59, 148, 243, 133, 69, 66, 123, 219, 2, 3, 229, 172, 221, 35, 185,
					208, 43, 44, 172, 96, 166, 116, 205, 93, 78, 194, 153, 95, 243, 145, 108, 96, 42, 6, 186, 128, 59, 117, 83, 196, 26,
					9, 15, 157, 215, 108, 180, 35, 195, 100, 7, 142, 47, 96, 108, 10},
				TransactionBodyBytes: []byte{16, 39, 0, 0, 0, 0, 0, 0},
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
				Log:                tt.fields.Log,
			}
			got, err := ts.PostTransaction(tt.args.chaintype, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.PostTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.PostTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionService_GetTransactions(t *testing.T) {
	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
	}
	type args struct {
		chainType contract.ChainType
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
			name: "GetTransactions:executeSelectGetTxsFail",
			fields: fields{
				Query: &mockGetTransactionExecutorTxsFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Limit:  2,
					Offset: 0,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactions:executeSelectGetTotalFail",
			fields: fields{
				Query: &mockGetTransactionExecutorTotalFail{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Limit:  2,
					Offset: 0,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactions:executeSelectGetTotalFailRow",
			fields: fields{
				Query: &mockGetTransactionExecutorTotalFailRow{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Limit:  2,
					Offset: 0,
				},
			},
			want:    &model.GetTransactionsResponse{},
			wantErr: true,
		},
		{
			name: "GetTransactions:executeSelectGetTotalFailRow",
			fields: fields{
				Query: &mockGetTransactionExecutorSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				params: &model.GetTransactionsRequest{
					Limit:  2,
					Offset: 0,
				},
			},
			want: &model.GetTransactionsResponse{
				Count: 2,
				Total: 2,
				Transactions: []*model.Transaction{
					{
						ID:                      4545420970999433273,
						BlockID:                 1,
						Height:                  1,
						SenderAccountType:       0,
						SenderAccountAddress:    "senderA",
						RecipientAccountType:    0,
						RecipientAccountAddress: "recipientA",
						TransactionType:         1,
						Fee:                     1,
						Timestamp:               10000,
						TransactionHash:         []byte{1, 1},
						TransactionBodyLength:   8,
						TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Signature:               []byte{0, 0, 0, 0, 0, 0, 0},
						Version:                 1,
					},
					{
						ID:                      4545420970999433274,
						BlockID:                 1,
						Height:                  1,
						SenderAccountType:       0,
						SenderAccountAddress:    "senderA",
						RecipientAccountType:    0,
						RecipientAccountAddress: "recipientA",
						TransactionType:         1,
						Fee:                     1,
						Timestamp:               10000,
						TransactionHash:         []byte{1, 1},
						TransactionBodyLength:   8,
						TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
						Signature:               []byte{0, 0, 0, 0, 0, 0, 0},
						Version:                 1,
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
			}
			got, err := ts.GetTransactions(tt.args.chainType, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.GetTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.GetTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactionService_GetTransaction(t *testing.T) {
	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
	}
	type args struct {
		chainType contract.ChainType
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
				Query: &mockGetTransactionExecutorTxSuccess{},
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
				SenderAccountType:       0,
				SenderAccountAddress:    "senderA",
				RecipientAccountType:    0,
				RecipientAccountAddress: "recipientA",
				TransactionType:         1,
				Fee:                     1,
				Timestamp:               10000,
				TransactionHash:         []byte{1, 1},
				TransactionBodyLength:   8,
				TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Signature:               []byte{0, 0, 0, 0, 0, 0, 0},
				Version:                 1,
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
			}
			got, err := ts.GetTransaction(tt.args.chainType, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransactionService.GetTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransactionService.GetTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
