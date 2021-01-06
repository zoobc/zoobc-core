// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/feedbacksystem"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
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

func (*mockTxTypeValidateFail) Validate(bool, bool) error {
	return errors.New("mockError:validateFail")
}

func (*mockTxTypeApplyUnconfirmedFail) Validate(bool, bool) error {
	return nil
}

func (*mockTxTypeSuccess) Validate(bool, bool) error {
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
		"Version", "message"}))
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
				nil,
				nil,
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

	var (
		sendMoneyTxBytes = []byte{1, 0, 0, 0, 1, 45, 230, 135, 95, 0, 0, 0, 0, 0, 0, 0, 0, 22, 42, 66, 34, 152, 48, 253, 178,
			113, 165, 192, 70, 53, 235, 121, 157, 138, 101, 3, 61, 204, 73, 16, 90, 211, 203, 42, 245, 241, 134, 173, 131, 0,
			0, 0, 0, 209, 39, 149, 255, 194, 205, 12, 110, 147, 76, 232, 143, 197, 139, 71, 162, 195, 147, 119, 235, 115, 12,
			231, 73, 49, 234, 207, 187, 242, 63, 97, 58, 65, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 128, 150, 152, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 210, 101, 51, 243, 100, 27,
			194, 204, 144, 1, 175, 209, 142, 115, 121, 46, 40, 121, 135, 142, 71, 154, 17, 95, 71, 146, 84, 32, 118, 159, 18,
			34, 130, 212, 36, 74, 216, 185, 83, 52, 230, 253, 195, 38, 52, 167, 16, 65, 208, 53, 216, 114, 168, 219, 57, 140,
			251, 189, 213, 101, 58, 65, 89, 11}
	)

	txTypeSuccess, transactionBytes := transaction.GetFixtureForSpecificTransaction(
		5298837107897007947,
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
		-62373445000112233,
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
		FeedbackStrategy   feedbacksystem.FeedbackStrategyInterface
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: sendMoneyTxBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
			},
			args: args{
				chaintype: &chaintype.MainChain{},
				req: &model.PostTransactionRequest{
					TransactionBytes: transactionBytes,
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
				FeedbackStrategy: &feedbacksystem.DummyFeedbackStrategy{},
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
				FeedbackStrategy:   tt.fields.FeedbackStrategy,
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
				[]byte{1, 2, 3},
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
						Message:                 []byte{1, 2, 3},
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
				"",
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
				Message:                 []byte{},
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
