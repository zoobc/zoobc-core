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
	"testing"

	"github.com/zoobc/zoobc-core/common/fee"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	executorSetupLiquidPaymentSuccess struct {
		query.Executor
	}
	executorSetupLiquidPaymentFail struct {
		query.Executor
	}
)

var (
	liquidPayAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
		72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	liquidPayAddress2 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
)

func (*executorSetupLiquidPaymentSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorSetupLiquidPaymentSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupLiquidPaymentSuccess) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, nil
}

func (*executorSetupLiquidPaymentFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("executor mock error")
}

func (*executorSetupLiquidPaymentFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("executor mock error")
}

func (*executorSetupLiquidPaymentFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, errors.New("executor mock error")
}

func TestLiquidPayment_ApplyConfirmed(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
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
			name: "wantError:executor_returns_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPayment_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:executor_returns_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPayment_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantError:executor_returns_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceQueryForLiquidPaymentSuccess struct {
		mockSpendableBalance int64
		query.AccountBalanceQuery
	}
	mockAccountBalanceQueryForLiquidPaymentFail struct {
		query.AccountBalanceQuery
	}
)

func (m *mockAccountBalanceQueryForLiquidPaymentSuccess) Scan(accountBalance *model.AccountBalance, row *sql.Row) error {
	accountBalance.SpendableBalance = m.mockSpendableBalance
	return nil
}

func (*mockAccountBalanceQueryForLiquidPaymentFail) Scan(accountBalance *model.AccountBalance, row *sql.Row) error {
	return errors.New("error mock")
}

func TestLiquidPayment_Validate(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
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
			name: "wantError:amount_is_equal_to_0",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          0,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:amount_is_less_than_0",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          -1,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:sender_address_is_empty",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:recipient_address_is_empty",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress2,
					RecipientAccountAddress: nil,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:select_account_balance_executor_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress2,
					RecipientAccountAddress: liquidPayAddress1,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:select_account_balance_scan_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress2,
					RecipientAccountAddress: liquidPayAddress1,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					&mockAccountBalanceQueryForLiquidPaymentFail{},
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:spendableBalance_is_less_than_amount+fee",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(&executorSetupLiquidPaymentSuccess{}, &mockAccountBalanceQueryForLiquidPaymentSuccess{
					mockSpendableBalance: 1,
				}, query.NewAccountLedgerQuery()),
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(&executorSetupLiquidPaymentSuccess{}, &mockAccountBalanceQueryForLiquidPaymentSuccess{
					mockSpendableBalance: 20,
				}, query.NewAccountLedgerQuery()),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPayment_GetMinimumFee(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		FeeScaleService               fee.FeeScaleServiceInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{},
				FeeScaleService:   &mockFeeScaleServiceValidateSuccess{},
			},
			want: constant.OneZBC / 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				FeeScaleService:               tt.fields.FeeScaleService,
			}
			got, err := tx.GetMinimumFee()
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.GetMinimumFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LiquidPayment.GetMinimumFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPayment_GetAmount(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.LiquidPaymentTransactionBody{
					Amount: 10,
				},
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("LiquidPayment.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPayment_GetSize(t *testing.T) {
	tests := []struct {
		name string
		want uint32
	}{
		{
			name: "wantSuccess",
			want: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("LiquidPayment.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPayment_ParseBodyBytes(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
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
			name: "wantErr:ParseBodyBytes - error (no amount)",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				Body:                 nil,
				AccountBalanceHelper: nil,
				QueryExecutor:        nil,
			},
			args:    args{txBodyBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantErr:ParseBodyBytes - error (wrong amount bytes lengths)",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				Body:                 nil,
				AccountBalanceHelper: nil,
				QueryExecutor:        nil,
			},
			args:    args{txBodyBytes: []byte{1, 2}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess:ParseBodyBytes - success",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				Body:                 nil,
				AccountBalanceHelper: nil,
				QueryExecutor:        nil,
			},
			args: args{txBodyBytes: []byte{1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}},
			want: &model.LiquidPaymentTransactionBody{
				Amount:          1,
				CompleteMinutes: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			got, err := tx.ParseBodyBytes(tt.args.txBodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPayment.ParseBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPayment_GetBodyBytes(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          1000,
					CompleteMinutes: 200,
				},
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				AccountBalanceHelper: nil,
				QueryExecutor:        nil,
			},
			want: []byte{
				232, 3, 0, 0, 0, 0, 0, 0, 200, 0, 0, 0, 0, 0, 0, 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPayment.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPayment_GetTransactionBody(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
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
				Body: &model.LiquidPaymentTransactionBody{
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
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestLiquidPayment_CompletePayment(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	type args struct {
		blockHeight           uint32
		blockTimestamp        int64
		firstAppliedTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantErr:blockTimeStamp_is_less_than_firstAppliedTimestamp",
			args: args{
				blockTimestamp:        1257894000,
				firstAppliedTimestamp: 1257894004,
			},
			wantErr: true,
		},
		{
			name: "wantErr:ExecuteTransactions_error",
			args: args{
				blockTimestamp:        1257894004,
				firstAppliedTimestamp: 1257894004,
			},
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantSuccess",
			args: args{
				blockTimestamp:        1257894004,
				firstAppliedTimestamp: 1257894004,
			},
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                      10,
					Fee:                     10,
					SenderAccountAddress:    liquidPayAddress1,
					RecipientAccountAddress: liquidPayAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentTransactionBody{
					Amount:          10,
					CompleteMinutes: 100,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			if err := tx.CompletePayment(tt.args.blockHeight, tt.args.blockTimestamp, tt.args.firstAppliedTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.CompletePayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPayment_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	type args struct {
		selectedTransactions []*model.Transaction
		newBlockTimestamp    int64
		newBlockHeight       uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "wantNoSkip",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions, tt.args.newBlockTimestamp, tt.args.newBlockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidPayment.SkipMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LiquidPayment.SkipMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPayment_Escrowable(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentTransactionBody
		QueryExecutor                 query.ExecutorInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   EscrowTypeAction
		want1  bool
	}{
		{
			name: "wantNonEscrowable",
			fields: fields{
				TransactionObject: &model.Transaction{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
			}
			got, got1 := tx.Escrowable()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPayment.Escrowable() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("LiquidPayment.Escrowable() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
