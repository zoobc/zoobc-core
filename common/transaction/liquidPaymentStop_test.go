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
	"strings"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	executorSetupLiquidPaymentStopSuccess struct {
		query.Executor
	}
	executorSetupLiquidPaymentStopFail struct {
		query.Executor
	}
	executorLiquidPaymentStopApplyConfirmed struct {
		query.Executor
	}

	mockLiquidPaymentTransactionQueryFail struct {
		query.LiquidPaymentTransactionQuery
	}

	mockLiquidPaymentTransactionQuerySuccess struct {
		Sender    []byte
		Recipient []byte
		Status    model.LiquidPaymentStatus
		query.LiquidPaymentTransactionQuery
	}

	mockTransactionQueryFail struct {
		query.TransactionQuery
	}

	mockTransactionQuerySuccess struct {
		query.TransactionQuery
	}

	mockTypeActionSwitcher struct {
		isError  bool
		returnTx TypeAction
		TypeActionSwitcher
	}

	mockLiquidPaymentTransaction struct {
		isError bool
		LiquidPaymentTransaction
	}
)

var (
	liquidPayStopAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
		72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	liquidPayStopAddress2 = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
	liquidPayStopAddress3 = []byte{0, 0, 0, 0, 33, 130, 42, 143, 177, 97, 43, 208, 76, 119, 240, 91, 41, 170, 240, 161, 55, 224, 8, 205,
		139, 227, 189, 146, 86, 211, 52, 194, 131, 126, 233, 100}
)

func (*executorSetupLiquidPaymentStopSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorSetupLiquidPaymentStopSuccess) ExecuteTransaction(qStr string, args ...interface{}) error {
	return nil
}

func (*executorSetupLiquidPaymentStopSuccess) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, nil
}

func (*executorSetupLiquidPaymentStopFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("executor mock error")
}

func (*executorSetupLiquidPaymentStopFail) ExecuteTransaction(qStr string, args ...interface{}) error {
	return errors.New("executor mock error")
}

func (*executorSetupLiquidPaymentStopFail) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return &sql.Row{}, errors.New("executor mock error")
}

func (*executorLiquidPaymentStopApplyConfirmed) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*executorLiquidPaymentStopApplyConfirmed) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*executorLiquidPaymentStopApplyConfirmed) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	var txID int64
	if len(args) > 0 {
		txID = args[0].(int64)
	}
	if strings.Contains(query, "liquid_payment_transaction") && txID == 666 {
		return nil, errors.New("mock error")
	} else if strings.Contains(query, "transaction") && strings.Contains(query, "666") {
		return nil, errors.New("mock error")
	}
	return &sql.Row{}, nil
}

func (*mockLiquidPaymentTransactionQueryFail) Scan(liquidPayment *model.LiquidPayment, row *sql.Row) error {
	return errors.New("mock error")
}

func (m *mockLiquidPaymentTransactionQuerySuccess) Scan(liquidPayment *model.LiquidPayment, row *sql.Row) error {
	liquidPayment.SenderAddress = m.Sender
	liquidPayment.RecipientAddress = m.Recipient
	liquidPayment.Status = m.Status
	return nil
}

func (*mockTransactionQueryFail) Scan(tx *model.Transaction, row *sql.Row) error {
	return errors.New("mock error")
}

func (*mockTransactionQuerySuccess) Scan(tx *model.Transaction, row *sql.Row) error {
	return nil
}

func (m *mockTypeActionSwitcher) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	if m.isError {
		return nil, errors.New("mock error")
	}
	return m.returnTx, nil
}

func (m *mockLiquidPaymentTransaction) CompletePayment(blockHeight uint32, blockTimestamp, firstAppliedTimestamp int64) error {
	if m.isError {
		return errors.New("mock error")
	}
	return nil
}

func TestLiquidPaymentStop_ApplyConfirmed(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
			name: "wantErr:ExecuteTransactions_error_balances",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				QueryExecutor: &executorSetupLiquidPaymentStopFail{},
			},
			wantErr: true,
		},
		{
			name: "wantErr:ExecuteSelectRow_error_GetPendingLiquidPaymentTransactionByID",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 666,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:Scan_error_LiquidPaymentTransactionQuery",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQueryFail{},
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:ExecuteSelectRow_error_GetTransaction",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{},
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 666,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:Scan_error_GetTransaction",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{},
				TransactionQuery:              &mockTransactionQueryFail{},
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:GetTransactionType_error",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{},
				TransactionQuery:              &mockTransactionQuerySuccess{},
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					isError: true,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:casting_error_liquidPaymentTransaction",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{},
				TransactionQuery:              &mockTransactionQuerySuccess{},
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					returnTx: &TXEmpty{},
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
			wantErr: true,
		},
		{
			name: "wantErr:CompletePayment_error",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{},
				TransactionQuery:              &mockTransactionQuerySuccess{},
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					returnTx: &mockLiquidPaymentTransaction{
						isError: true,
					},
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:status_is_already_completed",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{
					Status: model.LiquidPaymentStatus_LiquidPaymentCompleted,
				},
				QueryExecutor: &executorLiquidPaymentStopApplyConfirmed{},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
		},
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorLiquidPaymentStopApplyConfirmed{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{},
				TransactionQuery:              &mockTransactionQuerySuccess{},
				QueryExecutor:                 &executorLiquidPaymentStopApplyConfirmed{},
				TypeActionSwitcher: &mockTypeActionSwitcher{
					returnTx: &mockLiquidPaymentTransaction{
						isError: false,
					},
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentStop.ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPaymentStop_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
					SenderAccountAddress:    liquidPayStopAddress1,
					RecipientAccountAddress: liquidPayStopAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
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
					SenderAccountAddress:    liquidPayStopAddress1,
					RecipientAccountAddress: liquidPayStopAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentStop.ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPaymentStop_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
					SenderAccountAddress:    liquidPayStopAddress1,
					RecipientAccountAddress: liquidPayStopAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
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
					SenderAccountAddress:    liquidPayStopAddress1,
					RecipientAccountAddress: liquidPayStopAddress2,
					Height:                  10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentStop.UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperLiquidPaymentStopValidateSuccess struct {
		AccountBalanceHelper
	}
)

var (
	mockFeeLiquidPaymentStopValidate int64 = 10
)

func (*mockAccountBalanceHelperLiquidPaymentStopValidateSuccess) HasEnoughSpendableBalance(
	dbTX bool,
	address []byte,
	compareBalance int64,
) (enough bool, err error) {
	return true, nil
}
func (*mockAccountBalanceHelperLiquidPaymentStopValidateSuccess) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeLiquidPaymentStopValidate + 1
	return nil
}

func TestLiquidPaymentStop_Validate(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
			name: "wantError:sender_address_is_empty",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  10,
					SenderAccountAddress: nil,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:transactionID_is_empty",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  10,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 0,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:select_LiquidPaymentTransactionQuery_executor_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  10,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopFail{},
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:select_liquid_payment_scan_error",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  10,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor:                 &executorSetupLiquidPaymentStopFail{},
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQueryFail{},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopFail{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:transaction_sender_dont_match_with_sender_and_recipient",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  10,
					SenderAccountAddress: liquidPayStopAddress3,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor: &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{
					Sender:    liquidPayStopAddress1,
					Recipient: liquidPayStopAddress2,
				},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantError:status_is_not_pending",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  10,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor: &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{
					Sender: liquidPayStopAddress1,
					Status: model.LiquidPaymentStatus_LiquidPaymentCompleted,
				},
				AccountBalanceHelper: NewAccountBalanceHelper(
					&executorSetupLiquidPaymentStopSuccess{},
					query.NewAccountBalanceQuery(),
					query.NewAccountLedgerQuery(),
				),
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:sender_match_sender",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  mockFeeLiquidPaymentStopValidate,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor: &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{
					Sender: liquidPayStopAddress1,
					Status: model.LiquidPaymentStatus_LiquidPaymentPending,
				},
				AccountBalanceHelper: &mockAccountBalanceHelperLiquidPaymentStopValidateSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:sender_match_sender",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  mockFeeLiquidPaymentStopValidate,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor: &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{
					Sender: liquidPayStopAddress1,
					Status: model.LiquidPaymentStatus_LiquidPaymentPending,
				},
				AccountBalanceHelper: &mockAccountBalanceHelperLiquidPaymentStopValidateSuccess{},
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:sender_match_recipient",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   10,
					Fee:                  mockFeeLiquidPaymentStopValidate,
					SenderAccountAddress: liquidPayStopAddress1,
					Height:               10,
				},
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
				QueryExecutor: &executorSetupLiquidPaymentStopSuccess{},
				LiquidPaymentTransactionQuery: &mockLiquidPaymentTransactionQuerySuccess{
					Recipient: liquidPayStopAddress1,
					Status:    model.LiquidPaymentStatus_LiquidPaymentPending,
				},
				AccountBalanceHelper: &mockAccountBalanceHelperLiquidPaymentStopValidateSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentStop.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiquidPaymentStop_GetAmount(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "wantSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: 100,
				},
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if got := tx.GetAmount(); got != tt.want {
				t.Errorf("LiquidPaymentStop.GetAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentStop_GetSize(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "wantSuccess",
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("LiquidPaymentStop.GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentStop_ParseBodyBytes(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				QueryExecutor: nil,
			},
			args:    args{txBodyBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantErr:ParseBodyBytes - error (wrong amount bytes lengths)",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				QueryExecutor: nil,
			},
			args:    args{txBodyBytes: []byte{1, 2}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess:ParseBodyBytes - success",
			fields: fields{
				Body: nil,
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				QueryExecutor: nil,
			},
			args: args{txBodyBytes: []byte{1, 0, 0, 0, 0, 0, 0, 0}},
			want: &model.LiquidPaymentStopTransactionBody{
				TransactionID: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			got, err := tx.ParseBodyBytes(tt.args.txBodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentStop.ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPaymentStop.ParseBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentStop_GetBodyBytes(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytes:success",
			fields: fields{
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 1000,
				},
				TransactionObject: &model.Transaction{
					Fee:                     0,
					SenderAccountAddress:    nil,
					RecipientAccountAddress: nil,
					Height:                  0,
				},
				QueryExecutor: nil,
			},
			want: []byte{
				232, 3, 0, 0, 0, 0, 0, 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			if got, _ := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPaymentStop.GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentStop_GetTransactionBody(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
				Body: &model.LiquidPaymentStopTransactionBody{
					TransactionID: 123,
				},
			},
			args: args{
				transaction: &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			tx.GetTransactionBody(tt.args.transaction)
		})
	}
}

func TestLiquidPaymentStop_SkipMempoolTransaction(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			got, err := tx.SkipMempoolTransaction(tt.args.selectedTransactions, tt.args.newBlockTimestamp, tt.args.newBlockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("LiquidPaymentStop.SkipMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LiquidPaymentStop.SkipMempoolTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidPaymentStop_Escrowable(t *testing.T) {
	type fields struct {
		TransactionObject             *model.Transaction
		Body                          *model.LiquidPaymentStopTransactionBody
		QueryExecutor                 query.ExecutorInterface
		TransactionQuery              query.TransactionQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		AccountBalanceHelper          AccountBalanceHelperInterface
		TypeActionSwitcher            TypeActionSwitcher
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
			tx := &LiquidPaymentStopTransaction{
				TransactionObject:             tt.fields.TransactionObject,
				Body:                          tt.fields.Body,
				QueryExecutor:                 tt.fields.QueryExecutor,
				TransactionQuery:              tt.fields.TransactionQuery,
				LiquidPaymentTransactionQuery: tt.fields.LiquidPaymentTransactionQuery,
				AccountBalanceHelper:          tt.fields.AccountBalanceHelper,
				TypeActionSwitcher:            tt.fields.TypeActionSwitcher,
			}
			got, got1 := tx.Escrowable()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidPaymentStop.Escrowable() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("LiquidPaymentStop.Escrowable() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
