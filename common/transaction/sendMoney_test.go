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
package transaction

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var db, mock, _ = sqlmock.New()

type (
	executorAccountCountSuccess struct {
		query.Executor
	}
	executorAccountCountFail struct {
		query.Executor
	}

	executorApplyUnconfirmedSuccess struct {
		query.Executor
	}

	executorFailUpdateAccount struct {
		executorAccountCountSuccess
	}

	executorSuccessUpdateAccount struct {
		query.Executor
	}

	executorUnconfirmedFail struct {
		query.ExecutorInterface
	}
)

func (*executorAccountCountFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:accountInsertFail")
}

func (*executorAccountCountFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockError:accountInsertFail")
}

func (*executorFailUpdateAccount) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("mockError:accountbalanceFail")
}

func (*executorFailUpdateAccount) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockError:senderFail")
}

func (*executorSuccessUpdateAccount) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorSuccessUpdateAccount) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorAccountCountSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{
		"total_record",
	}).AddRow(2))

	return db.QueryRow(qStr, 1, 2), nil
}
func (*executorAccountCountSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewAccountBalanceQuery().Fields,
	).AddRow(1, 2, 3, 4, 5, 6))
	return db.Query(qStr, 1)
}

func (*executorAccountCountSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorUnconfirmedFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("MockedError")
}

func (*executorApplyUnconfirmedSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

type (
	mockQueryExecutorValidateSendMoneyHasEscrow struct {
		query.Executor
	}
	mockQueryExecutorValidateSendMoneyNeedEscrow struct {
		query.Executor
	}
	mockAccountBalanceValidateSendMoneySuccess struct {
		query.AccountBalanceQuery
	}
)

func (*mockQueryExecutorValidateSendMoneyHasEscrow) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := mock.NewRows(query.NewAccountDatasetsQuery().Fields)
	mock.ExpectQuery("").WillReturnRows(mockRow)
	row := db.QueryRow("")
	return row, nil
}

func (*mockQueryExecutorValidateSendMoneyHasEscrow) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	return &sql.Rows{}, nil
}

func (*mockQueryExecutorValidateSendMoneyNeedEscrow) ExecuteSelectRow(string, bool, ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := mock.NewRows(query.NewAccountDatasetsQuery().Fields)
	mockRow.AddRow(
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"AccountDatasetEscrowApproval",
		"You Welcome",
		true,
		true,
		5,
	)

	mock.ExpectQuery("").WillReturnRows(mockRow)
	row := db.QueryRow("")
	return row, nil
}

func (*mockAccountBalanceValidateSendMoneySuccess) Scan(accountBalance *model.AccountBalance, row *sql.Row) error {
	accountBalance.AccountAddress = senderAddress1
	accountBalance.BlockHeight = 10
	accountBalance.SpendableBalance = 10000
	accountBalance.Balance = 10
	accountBalance.PopRevenue = 0
	accountBalance.Latest = true
	return nil
}

func TestSendMoney_Validate(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        []byte
		SenderAccountType    uint32
		RecipientAddress     []byte
		RecipientAccountType uint32
		Height               uint32
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:AmountNotEnough",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: -1,
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "wantError:RecipientInvalid",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				RecipientAddress:     nil,
				RecipientAccountType: 0,
			},
			wantErr: true,
		},
		{
			name: "wantError:SenderInvalid",
			fields: fields{
				QueryExecutor: &mockQueryExecutorValidateSendMoneyHasEscrow{},
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        nil,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
			},
			wantErr: true,
		},
		{
			name: "wantError:SenderNotExists",
			fields: fields{
				QueryExecutor: &mockQueryExecutorValidateSendMoneyHasEscrow{},
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        nil,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &mockQueryExecutorValidateSendMoneyHasEscrow{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(false); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSendMoney_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        []byte
		SenderAccountType    uint32
		RecipientAddress     []byte
		RecipientAccountType uint32
		Height               uint32
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:ExecuteTransactionFail",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorUnconfirmedFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:ApplySuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorApplyUnconfirmedSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSendMoney_ApplyConfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        []byte
		SenderAccountType    uint32
		RecipientAddress     []byte
		RecipientAccountType uint32
		Height               uint32
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantFail:undoUnconfirmedFail",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorFailUpdateAccount{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "ExecuteTransactionFail",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Height:               0,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorFailUpdateAccount{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Height:               0,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorSuccessUpdateAccount{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(0); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendMoney_GetAmount(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        []byte
		SenderAccountType    uint32
		RecipientAddress     []byte
		RecipientAccountType uint32
		Height               uint32
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "GetAmount:success",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 100,
				},
				Height:               0,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorSuccessUpdateAccount{},
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:             tt.fields.Body,
				SenderAddress:    tt.fields.SenderAddress,
				RecipientAddress: tt.fields.RecipientAddress,
				Height:           tt.fields.Height,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("SendMoney.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMoney_GetSize(t *testing.T) {
	t.Run("SendMoney:GetSize", func(t *testing.T) {
		tx := &SendMoney{}
		size, _ := tx.GetSize()
		if size != 8 {
			t.Errorf("SendMoney size should be 8\nget: %d instead", size)
		}
	})
}

func TestSendMoney_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        []byte
		SenderAccountType    uint32
		RecipientAddress     []byte
		RecipientAccountType uint32
		Height               uint32
		QueryExecutor        query.ExecutorInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed:success",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorAccountCountSuccess{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed:executeTransactionFail",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height:               1,
				SenderAccountType:    0,
				SenderAddress:        senderAddress1,
				RecipientAccountType: 0,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &executorAccountCountFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				QueryExecutor:        tt.fields.QueryExecutor,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendMoney_GetBodyBytes(t *testing.T) {
	type fields struct {
		Body             *model.SendMoneyTransactionBody
		Fee              int64
		SenderAddress    []byte
		RecipientAddress []byte
		Height           uint32
		QueryExecutor    query.ExecutorInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1000,
				},
				Fee:              0,
				SenderAddress:    nil,
				RecipientAddress: nil,
				Height:           0,
				QueryExecutor:    nil,
			},
			want: []byte{
				232, 3, 0, 0, 0, 0, 0, 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:             tt.fields.Body,
				Fee:              tt.fields.Fee,
				SenderAddress:    tt.fields.SenderAddress,
				RecipientAddress: tt.fields.RecipientAddress,
				Height:           tt.fields.Height,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMoney_ParseBodyBytes(t *testing.T) {
	type fields struct {
		Body             *model.SendMoneyTransactionBody
		Fee              int64
		SenderAddress    []byte
		RecipientAddress []byte
		Height           uint32
		QueryExecutor    query.ExecutorInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TransactionBodyInterface
		wantErr bool
	}{
		{
			name: "SendMoney:ParseBodyBytes - error (no amount)",
			fields: fields{
				Body:             nil,
				Fee:              0,
				SenderAddress:    nil,
				RecipientAddress: nil,
				Height:           0,
				QueryExecutor:    nil,
			},
			args:    args{txBodyBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendMoney:ParseBodyBytes - error (wrong amount bytes lengths)",
			fields: fields{
				Body:             nil,
				Fee:              0,
				SenderAddress:    nil,
				RecipientAddress: nil,
				Height:           0,
				QueryExecutor:    nil,
			},
			args:    args{txBodyBytes: []byte{1, 2}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "SendMoney:ParseBodyBytes - success",
			fields: fields{
				Body:             nil,
				Fee:              0,
				SenderAddress:    nil,
				RecipientAddress: nil,
				Height:           0,
				QueryExecutor:    nil,
			},
			args: args{txBodyBytes: []byte{1, 0, 0, 0, 0, 0, 0, 0}},
			want: &model.SendMoneyTransactionBody{
				Amount: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:             tt.fields.Body,
				Fee:              tt.fields.Fee,
				SenderAddress:    tt.fields.SenderAddress,
				RecipientAddress: tt.fields.RecipientAddress,
				Height:           tt.fields.Height,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			got, err := tx.ParseBodyBytes(tt.args.txBodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBodyBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMoney_GetTransactionBody(t *testing.T) {
	type fields struct {
		Body             *model.SendMoneyTransactionBody
		Fee              int64
		SenderAddress    []byte
		RecipientAddress []byte
		Height           uint32
		QueryExecutor    query.ExecutorInterface
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
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
			},
			args: args{
				transaction: &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:             tt.fields.Body,
				Fee:              tt.fields.Fee,
				SenderAddress:    tt.fields.SenderAddress,
				RecipientAddress: tt.fields.RecipientAddress,
				Height:           tt.fields.Height,
				QueryExecutor:    tt.fields.QueryExecutor,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestSendMoney_Escrowable(t *testing.T) {
	type fields struct {
		ID               int64
		Fee              int64
		SenderAddress    []byte
		RecipientAddress []byte
		Height           uint32
		Body             *model.SendMoneyTransactionBody
		Escrow           *model.Escrow
		QueryExecutor    query.ExecutorInterface
		EscrowQuery      query.EscrowTransactionQueryInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   EscrowTypeAction
		want1  bool
	}{
		{
			name: "wantNonEscrow",
			fields: fields{
				ID:               0,
				Fee:              0,
				SenderAddress:    nil,
				RecipientAddress: nil,
				Height:           0,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				QueryExecutor: nil,
				EscrowQuery:   nil,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "wantEscrow",
			fields: fields{
				ID:               1,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           0,
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Escrow: &model.Escrow{
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					Commission:       10,
					Timeout:          1,
				},
				QueryExecutor: nil,
				EscrowQuery:   nil,
			},
			want: EscrowTypeAction(&SendMoney{
				ID:               1,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           0,
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Escrow: &model.Escrow{
					ID:               1,
					Amount:           1,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					Commission:       10,
					Timeout:          1,
					Latest:           true,
				},
				QueryExecutor: nil,
				EscrowQuery:   nil,
			}),
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				ID:               tt.fields.ID,
				Fee:              tt.fields.Fee,
				SenderAddress:    tt.fields.SenderAddress,
				RecipientAddress: tt.fields.RecipientAddress,
				Height:           tt.fields.Height,
				Body:             tt.fields.Body,
				Escrow:           tt.fields.Escrow,
				QueryExecutor:    tt.fields.QueryExecutor,
				EscrowQuery:      tt.fields.EscrowQuery,
			}
			got, got1 := tx.Escrowable()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Escrowable() got = \n%v, want \n%v", got, tt.want)
				return
			}
			if got1 != tt.want1 {
				t.Errorf("Escrowable() got1 = \n%v, want \n%v", got1, tt.want1)
			}
		})
	}
}

type (
	mockExecutorEscrowValidateValid struct {
		query.Executor
	}
	mockBlockQueryValidBlockHeight struct {
		query.BlockQuery
	}
	mockExecutorEscrowValidateInvalidBlockHeight struct {
		query.Executor
	}
	mockBlockQueryInvalidBlockHeight struct {
		query.BlockQuery
	}
)

func (*mockExecutorEscrowValidateValid) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnError(sql.ErrNoRows)
	row := db.QueryRow(qe)
	return row, nil
}

func (*mockBlockQueryValidBlockHeight) GetLastBlock() string {
	return ""
}

func (*mockBlockQueryValidBlockHeight) Scan(block *model.Block, row *sql.Row) error {
	block.Height = 1
	return nil
}
func (*mockExecutorEscrowValidateInvalidBlockHeight) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, nil
}
func (*mockBlockQueryInvalidBlockHeight) GetLastBlock() string {
	return ""
}

func (*mockBlockQueryInvalidBlockHeight) Scan(block *model.Block, row *sql.Row) error {
	block.Height = 1
	return nil
}

func TestSendMoney_EscrowValidate(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		RecipientAddress     []byte
		Height               uint32
		Body                 *model.SendMoneyTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		BlockQuery           query.BlockQueryInterface
		AccountBalanceHelper AccountBalanceHelperInterface
		AccountDatasetQuery  query.AccountDatasetQueryInterface
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
			name: "wantError:AmountNotEnough",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: -1,
				},
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
		{
			name: "wantError:ApproverAddressRequired",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Escrow: &model.Escrow{
					Commission: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "wantError:RecipientRequired",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Escrow: &model.Escrow{
					Commission:      1,
					ApproverAddress: senderAddress2,
				},
			},
			wantErr: true,
		},
		{
			name: "wantError:InvalidTimeout",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Escrow: &model.Escrow{
					Commission:       1,
					ApproverAddress:  senderAddress2,
					RecipientAddress: recipientAddress1,
				},
				QueryExecutor: &mockExecutorEscrowValidateInvalidBlockHeight{},
				BlockQuery:    &mockBlockQueryInvalidBlockHeight{},
			},
			wantErr: true,
		},
		{
			name: "wantError:SenderAddressRequired",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				Escrow: &model.Escrow{
					Commission:       1,
					ApproverAddress:  senderAddress2,
					RecipientAddress: recipientAddress1,
					Timeout:          10,
				},
				QueryExecutor: &mockExecutorEscrowValidateValid{},
				BlockQuery:    &mockBlockQueryValidBlockHeight{},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 1,
				},
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Escrow: &model.Escrow{
					Commission:       1,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					Timeout:          10,
				},
				QueryExecutor:        &mockExecutorEscrowValidateValid{},
				BlockQuery:           &mockBlockQueryValidBlockHeight{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				BlockQuery:           tt.fields.BlockQuery,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.EscrowValidate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("EscrowValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockEscrowApplyUnconfirmedOK struct {
		query.Executor
	}
)

func (*mockEscrowApplyUnconfirmedOK) ExecuteTransaction(string, ...interface{}) error {
	return nil
}

func TestSendMoney_EscrowApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		RecipientAddress     []byte
		Height               uint32
		Body                 *model.SendMoneyTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		BlockQuery           query.BlockQueryInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:               1234567890,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           1,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Escrow: &model.Escrow{
					ID:               1234567890,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					BlockHeight:      1,
				},
				QueryExecutor:        &mockEscrowApplyUnconfirmedOK{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				BlockQuery:           tt.fields.BlockQuery,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.EscrowApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("EscrowApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendMoney_EscrowUndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		RecipientAddress     []byte
		Height               uint32
		Body                 *model.SendMoneyTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		BlockQuery           query.BlockQueryInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ID:               1234567890,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           1,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Escrow: &model.Escrow{
					ID:               1234567890,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					BlockHeight:      1,
				},
				QueryExecutor:        &mockEscrowApplyUnconfirmedOK{},
				BlockQuery:           query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				BlockQuery:           tt.fields.BlockQuery,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.EscrowUndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("EscrowUndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockBlockQueryApplyConfirmedOK struct {
		query.BlockQuery
	}
	mockQueryExecutorApplyConfirmedOK struct {
		query.Executor
	}
)

func (*mockBlockQueryApplyConfirmedOK) GetLastBlock() string {
	return ""
}
func (*mockBlockQueryApplyConfirmedOK) Scan(block *model.Block, row *sql.Row) error {
	block.Height = 100
	return nil
}

func (*mockQueryExecutorApplyConfirmedOK) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, nil
}
func (*mockQueryExecutorApplyConfirmedOK) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestSendMoney_EscrowApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		RecipientAddress     []byte
		Height               uint32
		Body                 *model.SendMoneyTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		BlockQuery           query.BlockQueryInterface
		AccountBalanceHelper AccountBalanceHelperInterface
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
			name: "wantSuccess",
			fields: fields{
				ID:               1234567890,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           1,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Escrow: &model.Escrow{
					ID:               1234567890,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					BlockHeight:      1,
				},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				BlockQuery:           &mockBlockQueryApplyConfirmedOK{},
				QueryExecutor:        &mockQueryExecutorApplyConfirmedOK{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			args: args{blockTimestamp: 123456789},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				BlockQuery:           tt.fields.BlockQuery,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.EscrowApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("EscrowApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryEscrowApprovalOK struct {
		query.Executor
	}
)

func (*mockQueryEscrowApprovalOK) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}
func TestSendMoney_EscrowApproval(t *testing.T) {
	type fields struct {
		ID                   int64
		Fee                  int64
		SenderAddress        []byte
		RecipientAddress     []byte
		Height               uint32
		Body                 *model.SendMoneyTransactionBody
		Escrow               *model.Escrow
		QueryExecutor        query.ExecutorInterface
		EscrowQuery          query.EscrowTransactionQueryInterface
		BlockQuery           query.BlockQueryInterface
		AccountBalanceHelper AccountBalanceHelperInterface
	}
	type args struct {
		blockTimestamp int64
		txBody         *model.ApprovalEscrowTransactionBody
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess:Approved",
			fields: fields{
				ID:               1234567890,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           1,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Escrow: &model.Escrow{
					ID:               1234567890,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					BlockHeight:      1,
				},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				QueryExecutor:        &mockQueryEscrowApprovalOK{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			args: args{
				blockTimestamp: 100,
				txBody: &model.ApprovalEscrowTransactionBody{
					Approval:      0,
					TransactionID: 1234567890,
				}},
		},
		{
			name: "wantSuccess:Rejected",
			fields: fields{
				ID:               1234567890,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           1,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Escrow: &model.Escrow{
					ID:               1234567890,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					BlockHeight:      1,
				},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				QueryExecutor:        &mockQueryEscrowApprovalOK{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			args: args{
				blockTimestamp: 100,
				txBody: &model.ApprovalEscrowTransactionBody{
					Approval:      1,
					TransactionID: 1234567890,
				}},
		},
		{
			name: "WantSuccess:Expired",
			fields: fields{
				ID:               1234567890,
				Fee:              1,
				SenderAddress:    senderAddress1,
				RecipientAddress: recipientAddress1,
				Height:           1,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Escrow: &model.Escrow{
					ID:               1234567890,
					SenderAddress:    senderAddress1,
					RecipientAddress: recipientAddress1,
					ApproverAddress:  senderAddress2,
					Amount:           10,
					Commission:       1,
					Timeout:          123456789,
					Status:           model.EscrowStatus_Expired,
					BlockHeight:      1,
					Latest:           true,
					Instruction:      "Do this",
				},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				QueryExecutor:        &mockQueryEscrowApprovalOK{},
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
			},
			args: args{
				blockTimestamp: 100,
				txBody: &model.ApprovalEscrowTransactionBody{
					Approval:      model.EscrowApproval_Expire,
					TransactionID: 1234567890,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				ID:                   tt.fields.ID,
				Fee:                  tt.fields.Fee,
				SenderAddress:        tt.fields.SenderAddress,
				RecipientAddress:     tt.fields.RecipientAddress,
				Height:               tt.fields.Height,
				Body:                 tt.fields.Body,
				Escrow:               tt.fields.Escrow,
				QueryExecutor:        tt.fields.QueryExecutor,
				EscrowQuery:          tt.fields.EscrowQuery,
				BlockQuery:           tt.fields.BlockQuery,
				AccountBalanceHelper: tt.fields.AccountBalanceHelper,
			}
			if err := tx.EscrowApproval(tt.args.blockTimestamp, tt.args.txBody); (err != nil) != tt.wantErr {
				t.Errorf("EscrowApproval() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
