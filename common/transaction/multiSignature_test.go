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

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// multisignatureInfoHelper mocks
	multisignatureInfoHelperMultisignatureInfoQueryScanFail struct {
		query.MultisignatureInfoQuery
	}
	multisignatureInfoHelperQueryExecutorSuccess struct {
		emptyResultSet bool
		query.Executor
	}
	// multisignatureInfoHelper mocks

)

var (
	// multisignatureInfoHelper mocks
	mockMultisignatureInfoHelperMultisigInfoSuccess = &model.MultiSignatureInfo{
		MinimumSignatures: 2,
		Nonce:             1,
		Addresses: [][]byte{
			{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79,
				28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
			{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
				81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		},
		MultisigAddress: []byte{0, 0, 0, 0, 178, 223, 128, 179, 51, 150, 104, 6, 181, 133, 185, 121, 163, 139, 51, 120, 246, 15, 250, 56,
			118, 159, 166, 97, 98, 40, 70, 130, 35, 164, 104, 182},
		BlockHeight: 720,
		Latest:      true,
	}
	// multisignatureInfoHelper mocks
)

func (*multisignatureInfoHelperMultisignatureInfoQueryScanFail) GetMultisignatureInfoByAddressWithParticipants(
	[]byte, uint32, uint32,
) (str string, args []interface{}) {
	return "", []interface{}{}
}
func (*multisignatureInfoHelperMultisignatureInfoQueryScanFail) Scan(*model.MultiSignatureInfo, *sql.Row) error {
	return errors.New("mockedError")
}

func (*multisignatureInfoHelperQueryExecutorSuccess) ExecuteSelectRow(
	string, bool, ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	multisigInfoQuery := query.NewMultisignatureInfoQuery()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(append(multisigInfoQuery.Fields, "account_address")).AddRow(
		mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
		mockMultisignatureInfoHelperMultisigInfoSuccess.MinimumSignatures,
		mockMultisignatureInfoHelperMultisigInfoSuccess.Nonce,
		mockMultisignatureInfoHelperMultisigInfoSuccess.BlockHeight,
		mockMultisignatureInfoHelperMultisigInfoSuccess.Latest,
		mockMultisignatureInfoHelperMultisigInfoSuccess.Addresses[0],
		[]byte{},
	))
	row := db.QueryRow("")
	return row, nil
}

func (mockMsi *multisignatureInfoHelperQueryExecutorSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	dbMocked, mock, _ := sqlmock.New()
	defer db.Close()

	multisigInfoQuery := query.NewMultisignatureInfoQuery()
	if mockMsi.emptyResultSet {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(append(multisigInfoQuery.Fields, "account_address")))
	} else {
		mockRows := mock.NewRows(append(multisigInfoQuery.Fields, "account_address"))
		mockRows.AddRow(
			mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
			mockMultisignatureInfoHelperMultisigInfoSuccess.MinimumSignatures,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Nonce,
			mockMultisignatureInfoHelperMultisigInfoSuccess.BlockHeight,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Latest,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Addresses[0],
		)
		mockRows.AddRow(
			mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
			mockMultisignatureInfoHelperMultisigInfoSuccess.MinimumSignatures,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Nonce,
			mockMultisignatureInfoHelperMultisigInfoSuccess.BlockHeight,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Latest,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Addresses[1],
		)
		mockRows.AddRow(
			mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
			mockMultisignatureInfoHelperMultisigInfoSuccess.MinimumSignatures,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Nonce,
			mockMultisignatureInfoHelperMultisigInfoSuccess.BlockHeight,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Latest,
			mockMultisignatureInfoHelperMultisigInfoSuccess.Addresses[2],
		)
		mock.ExpectQuery("").WillReturnRows(mockRows)
	}
	return dbMocked.Query("")

}

func TestMultisignatureInfoHelper_GetMultisigInfoByAddress(t *testing.T) {
	var (
		multisigInfoSuccess model.MultiSignatureInfo
	)
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		multisigInfo    *model.MultiSignatureInfo
		multisigAddress []byte
		blockHeight     uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GetMultisigInfo - success",
			fields: fields{
				MultisignatureInfoQuery: query.NewMultisignatureInfoQuery(),
				QueryExecutor:           &multisignatureInfoHelperQueryExecutorSuccess{},
			},
			args: args{
				multisigInfo:    &multisigInfoSuccess,
				multisigAddress: mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
				blockHeight:     720,
			},
			wantErr: false,
		},
		{
			name: "GetMultisigInfo - buildModel empty",
			fields: fields{
				MultisignatureInfoQuery: &multisignatureInfoHelperMultisignatureInfoQueryScanFail{},
				QueryExecutor: &multisignatureInfoHelperQueryExecutorSuccess{
					emptyResultSet: true,
				},
			},
			args: args{
				multisigInfo:    &multisigInfoSuccess,
				multisigAddress: mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
				blockHeight:     720,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := msi.GetMultisigInfoByAddress(tt.args.multisigInfo, tt.args.multisigAddress, tt.args.blockHeight); (err != nil) != tt.wantErr {
				t.Errorf("GetMultisigInfoByAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock multisignatureInfoHelperInsertMultisignatureInfo
	multisignatureInfoHelperInsertMultisignatureInfoExecutorSuccess struct {
		query.Executor
	}
	multisignatureInfoHelperInsertMultisignatureInfoExecutorFail struct {
		query.Executor
	}
	// mock multisignatureInfoHelperInsertMultisignatureInfo
)

func (*multisignatureInfoHelperInsertMultisignatureInfoExecutorSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*multisignatureInfoHelperInsertMultisignatureInfoExecutorFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockedError")
}

func TestMultisignatureInfoHelper_InsertMultisignatureInfo(t *testing.T) {
	var (
		multisigInfoSuccess model.MultiSignatureInfo
	)
	type fields struct {
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		QueryExecutor                  query.ExecutorInterface
	}
	type args struct {
		multisigInfo *model.MultiSignatureInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "InsertMultisignatureInfo - success",
			fields: fields{
				MultisignatureInfoQuery:        query.NewMultisignatureInfoQuery(),
				MultiSignatureParticipantQuery: query.NewMultiSignatureParticipantQuery(),
				QueryExecutor:                  &multisignatureInfoHelperInsertMultisignatureInfoExecutorSuccess{},
			},
			args: args{
				multisigInfo: &multisigInfoSuccess,
			},
			wantErr: false,
		},
		{
			name: "InsertMultisignatureInfo - fail",
			fields: fields{
				MultisignatureInfoQuery:        query.NewMultisignatureInfoQuery(),
				MultiSignatureParticipantQuery: query.NewMultiSignatureParticipantQuery(),
				QueryExecutor:                  &multisignatureInfoHelperInsertMultisignatureInfoExecutorFail{},
			},
			args: args{
				multisigInfo: &multisigInfoSuccess,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msi := &MultisignatureInfoHelper{
				MultisignatureInfoQuery:        tt.fields.MultisignatureInfoQuery,
				MultiSignatureParticipantQuery: tt.fields.MultiSignatureParticipantQuery,
				QueryExecutor:                  tt.fields.QueryExecutor,
			}
			if err := msi.InsertMultisignatureInfo(tt.args.multisigInfo); (err != nil) != tt.wantErr {
				t.Errorf("InsertMultisignatureInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock multisignatureInfoHelperInsertPendingSignatureExecutor
	multisignatureInfoHelperInsertPendingSignatureExecutorSuccess struct {
		query.Executor
	}
	multisignatureInfoHelperInsertPendingSignatureExecutorFail struct {
		query.Executor
	}
	// mock multisignatureInfoHelperInsertPendingSignatureExecutor
)

func (*multisignatureInfoHelperInsertPendingSignatureExecutorSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*multisignatureInfoHelperInsertPendingSignatureExecutorFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockedError")
}

func TestSignatureInfoHelper_InsertPendingSignature(t *testing.T) {
	var pendingSignatureSuccess model.PendingSignature
	type fields struct {
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		QueryExecutor           query.ExecutorInterface
		Signature               crypto.SignatureInterface
	}
	type args struct {
		pendingSignature *model.PendingSignature
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "InsertPendingSignature - success",
			fields: fields{
				PendingSignatureQuery: query.NewPendingSignatureQuery(),
				QueryExecutor:         &multisignatureInfoHelperInsertPendingSignatureExecutorSuccess{},
			},
			args: args{
				pendingSignature: &pendingSignatureSuccess,
			},
			wantErr: false,
		},
		{
			name: "InsertPendingSignature - fail",
			fields: fields{
				PendingSignatureQuery: query.NewPendingSignatureQuery(),
				QueryExecutor:         &multisignatureInfoHelperInsertPendingSignatureExecutorFail{},
			},
			args: args{
				pendingSignature: &pendingSignatureSuccess,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sih := &SignatureInfoHelper{
				PendingSignatureQuery:   tt.fields.PendingSignatureQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				Signature:               tt.fields.Signature,
			}
			if err := sih.InsertPendingSignature(tt.args.pendingSignature); (err != nil) != tt.wantErr {
				t.Errorf("InsertPendingSignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutor
	multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorSuccess struct {
		query.Executor
	}
	multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorFail struct {
		query.Executor
	}
	multisignatureInfoHelperGetPendingSignatureByTransactionHashPendingSignatureQueryBuildFail struct {
		query.PendingSignatureQuery
	}
	// mock multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutor
)

var (
	// mock multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutor
	mockGetPendingSignatureByTransactionHashSuccessPendingSignatures = []*model.PendingSignature{
		{
			TransactionHash: make([]byte, 32),
			AccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			Signature:   make([]byte, 68),
			BlockHeight: 720,
			Latest:      true,
		},
	}
	// mock multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutor
)

func (*multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorFail) ExecuteSelect(
	string, bool, ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorSuccess) ExecuteSelect(
	string, bool, ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	pendingSignatureQuery := query.NewPendingSignatureQuery()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(pendingSignatureQuery.Fields).AddRow(
		mockGetPendingSignatureByTransactionHashSuccessPendingSignatures[0].TransactionHash,
		mockGetPendingSignatureByTransactionHashSuccessPendingSignatures[0].AccountAddress,
		mockGetPendingSignatureByTransactionHashSuccessPendingSignatures[0].Signature,
		mockGetPendingSignatureByTransactionHashSuccessPendingSignatures[0].BlockHeight,
		mockGetPendingSignatureByTransactionHashSuccessPendingSignatures[0].Latest,
	))
	rows, _ := db.Query("")
	return rows, nil
}

func (*multisignatureInfoHelperGetPendingSignatureByTransactionHashPendingSignatureQueryBuildFail) BuildModel(
	[]*model.PendingSignature, *sql.Rows,
) ([]*model.PendingSignature, error) {
	return nil, errors.New("mockedError")
}

func TestSignatureInfoHelper_GetPendingSignatureByTransactionHash(t *testing.T) {
	var pendingSigsSuccess []*model.PendingSignature
	type fields struct {
		PendingSignatureQuery   query.PendingSignatureQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		QueryExecutor           query.ExecutorInterface
		Signature               crypto.SignatureInterface
	}
	type args struct {
		pendingSigs     []*model.PendingSignature
		transactionHash []byte
		txHeight        uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GetPendingSignatureByTransactionHash - success",
			fields: fields{
				PendingSignatureQuery: query.NewPendingSignatureQuery(),
				QueryExecutor:         &multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorSuccess{},
			},
			args: args{
				pendingSigs:     pendingSigsSuccess,
				transactionHash: make([]byte, 32),
				txHeight:        720,
			},
			wantErr: false,
		},
		{
			name: "GetPendingSignatureByTransactionHash - fail - Executor",
			fields: fields{
				PendingSignatureQuery: query.NewPendingSignatureQuery(),
				QueryExecutor:         &multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorFail{},
			},
			args: args{
				pendingSigs:     pendingSigsSuccess,
				transactionHash: make([]byte, 32),
				txHeight:        720,
			},
			wantErr: true,
		},
		{
			name: "GetPendingSignatureByTransactionHash - fail - Build",
			fields: fields{
				PendingSignatureQuery: &multisignatureInfoHelperGetPendingSignatureByTransactionHashPendingSignatureQueryBuildFail{},
				QueryExecutor:         &multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutorSuccess{},
			},
			args: args{
				pendingSigs:     pendingSigsSuccess,
				transactionHash: make([]byte, 32),
				txHeight:        720,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sih := &SignatureInfoHelper{
				PendingSignatureQuery:   tt.fields.PendingSignatureQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				QueryExecutor:           tt.fields.QueryExecutor,
				Signature:               tt.fields.Signature,
			}
			if _, err := sih.GetPendingSignatureByTransactionHash(
				tt.args.transactionHash, tt.args.txHeight); (err != nil) != tt.wantErr {
				t.Errorf("GetPendingSignatureByTransactionHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock multisignatureInfoHelperInsertPendingTransactionExecutor
	multisignatureInfoHelperInsertPendingTransactionExecutorSuccess struct {
		query.Executor
	}
	multisignatureInfoHelperInsertPendingTransactionExecutorFail struct {
		query.Executor
	}
	// mock multisignatureInfoHelperInsertPendingTransactionExecutor
)

func (*multisignatureInfoHelperInsertPendingTransactionExecutorSuccess) ExecuteTransactions([][]interface{}) error {
	return nil
}

func (*multisignatureInfoHelperInsertPendingTransactionExecutorFail) ExecuteTransactions([][]interface{}) error {
	return errors.New("mockedError")
}

func TestPendingTransactionHelper_InsertPendingTransaction(t *testing.T) {
	var pendingTransactionSuccess model.PendingTransaction
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		pendingTransaction *model.PendingTransaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "InsertPendingSignature - success",
			fields: fields{
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           &multisignatureInfoHelperInsertPendingSignatureExecutorSuccess{},
			},
			args: args{
				pendingTransaction: &pendingTransactionSuccess,
			},
			wantErr: false,
		},
		{
			name: "InsertPendingSignature - fail",
			fields: fields{
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           &multisignatureInfoHelperInsertPendingSignatureExecutorFail{},
			},
			args: args{
				pendingTransaction: &pendingTransactionSuccess,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pth := &PendingTransactionHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				TransactionUtil:         tt.fields.TransactionUtil,
				TypeSwitcher:            tt.fields.TypeSwitcher,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := pth.InsertPendingTransaction(tt.args.pendingTransaction); (err != nil) != tt.wantErr {
				t.Errorf("InsertPendingTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock PendingTransactionHelperGetPendingTransactionByHash
	pendingTransactionHelperPendingTransactionQueryScanFail struct {
		query.PendingTransactionQuery
	}
	pendingTransactionHelperQueryExecutorSuccess struct {
		query.Executor
	}
	// mock PendingTransactionHelperGetPendingTransactionByHash
)

var (
	mockGetPendingTransactionByHashPendingTransactionSuccess = &model.PendingTransaction{
		SenderAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		TransactionHash:  make([]byte, 32),
		TransactionBytes: make([]byte, 30),
		Status:           model.PendingTransactionStatus_PendingTransactionPending,
		BlockHeight:      720,
		Latest:           true,
	}
)

func (*pendingTransactionHelperQueryExecutorSuccess) ExecuteSelectRow(
	string, bool, ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	pendingTransactionQuery := query.NewPendingTransactionQuery()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(pendingTransactionQuery.Fields).AddRow(
		mockGetPendingTransactionByHashPendingTransactionSuccess.SenderAddress,
		mockGetPendingTransactionByHashPendingTransactionSuccess.TransactionHash,
		mockGetPendingTransactionByHashPendingTransactionSuccess.TransactionBytes,
		mockGetPendingTransactionByHashPendingTransactionSuccess.Status,
		mockGetPendingTransactionByHashPendingTransactionSuccess.BlockHeight,
		mockGetPendingTransactionByHashPendingTransactionSuccess.Latest,
	))
	row := db.QueryRow("")
	return row, nil
}

func (*pendingTransactionHelperPendingTransactionQueryScanFail) Scan(*model.PendingTransaction, *sql.Row) error {
	return errors.New("mockedError")
}

func TestPendingTransactionHelper_GetPendingTransactionByHash(t *testing.T) {
	var pendingTransactionSuccess model.PendingTransaction
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		pendingTx                  *model.PendingTransaction
		pendingTransactionHash     []byte
		pendingTransactionStatuses []model.PendingTransactionStatus
		blockHeight                uint32
		dbTx                       bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GetPendingTransactionByHash - success",
			fields: fields{
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           &pendingTransactionHelperQueryExecutorSuccess{},
			},
			args: args{
				pendingTx:              &pendingTransactionSuccess,
				pendingTransactionHash: make([]byte, 32),
				pendingTransactionStatuses: []model.PendingTransactionStatus{
					model.PendingTransactionStatus_PendingTransactionPending,
					model.PendingTransactionStatus_PendingTransactionExecuted,
				},
			},
			wantErr: false,
		},
		{
			name: "GetPendingTransactionByHash - fail",
			fields: fields{
				PendingTransactionQuery: &pendingTransactionHelperPendingTransactionQueryScanFail{},
				QueryExecutor:           &pendingTransactionHelperQueryExecutorSuccess{},
			},
			args: args{
				pendingTx:              &pendingTransactionSuccess,
				pendingTransactionHash: make([]byte, 32),
				pendingTransactionStatuses: []model.PendingTransactionStatus{
					model.PendingTransactionStatus_PendingTransactionPending,
					model.PendingTransactionStatus_PendingTransactionExecuted,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pth := &PendingTransactionHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				TransactionUtil:         tt.fields.TransactionUtil,
				TypeSwitcher:            tt.fields.TypeSwitcher,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := pth.GetPendingTransactionByHash(
				tt.args.pendingTx, tt.args.pendingTransactionHash,
				tt.args.pendingTransactionStatuses, tt.args.blockHeight, tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("GetPendingTransactionByHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock pendingSignatureHelperGetPendingTransactionBySenderAddressExecutor
	pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorSuccess struct {
		query.Executor
	}
	pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorFail struct {
		query.Executor
	}
	pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorPendingTransactionQueryBuildFail struct {
		query.PendingTransactionQuery
	}
	// mock pendingSignatureHelperGetPendingTransactionBySenderAddressExecutor
)

var (
	// mock multisignatureInfoHelperGetPendingTransactionByTransactionHashExecutor
	mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions = []*model.PendingTransaction{
		{
			SenderAddress: []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
				202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
			TransactionHash:  make([]byte, 32),
			TransactionBytes: make([]byte, 30),
			Status:           model.PendingTransactionStatus_PendingTransactionPending,
			BlockHeight:      720,
			Latest:           true,
		},
	}
	// mock multisignatureInfoHelperGetPendingSignatureByTransactionHashExecutor
)

func (*pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorFail) ExecuteSelect(
	string, bool, ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorSuccess) ExecuteSelect(
	string, bool, ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	pendingTransactionQuery := query.NewPendingTransactionQuery()
	mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(pendingTransactionQuery.Fields).AddRow(
		mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions[0].SenderAddress,
		mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions[0].TransactionHash,
		mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions[0].TransactionBytes,
		mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions[0].Status,
		mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions[0].BlockHeight,
		mockPendingTransactionHelperGetPendingTransactionsBySenderAddressSuccessPendingTransactions[0].Latest,
	))
	rows, _ := db.Query("")
	return rows, nil
}

func (*pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorPendingTransactionQueryBuildFail) BuildModel(
	[]*model.PendingTransaction, *sql.Rows,
) ([]*model.PendingTransaction, error) {
	return nil, errors.New("mockedError")
}

func TestPendingTransactionHelper_GetPendingTransactionBySenderAddress(t *testing.T) {
	var pendingTxs []*model.PendingTransaction
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		pendingTxs    []*model.PendingTransaction
		senderAddress []byte
		txHeight      uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "GetPendingTransactionBySenderAddress - success",
			fields: fields{
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           &pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorSuccess{},
			},
			args: args{
				pendingTxs: pendingTxs,
			},
			wantErr: false,
		},
		{
			name: "GetPendingTransactionBySenderAddress - executeSelectFail fail",
			fields: fields{
				PendingTransactionQuery: query.NewPendingTransactionQuery(),
				QueryExecutor:           &pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorFail{},
			},
			args: args{
				pendingTxs: pendingTxs,
			},
			wantErr: true,
		},
		{
			name: "GetPendingTransactionBySenderAddress - buildModelFail",
			fields: fields{
				PendingTransactionQuery: &pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorPendingTransactionQueryBuildFail{},
				QueryExecutor:           &pendingTransactionHelperGetPendingTransactionBySenderAddressExecutorSuccess{},
			},
			args: args{
				pendingTxs: pendingTxs,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pth := &PendingTransactionHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				TransactionUtil:         tt.fields.TransactionUtil,
				TypeSwitcher:            tt.fields.TypeSwitcher,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if _, err := pth.GetPendingTransactionBySenderAddress(
				tt.args.senderAddress, tt.args.txHeight); (err != nil) != tt.wantErr {
				t.Errorf("GetPendingTransactionBySenderAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
	pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess struct {
		Util
	}
	pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseFail struct {
		Util
	}
	pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionSuccess struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyUnconfirmedPendingTransactionActionTypeFail struct {
		TypeAction
	}
	pendingTransactionhelperApplyUnconfirmedPendingTransactionActionTypeSuccess struct {
		TypeAction
	}
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
)

var (
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
	pendingTransactionHelperApplyUnconfirmedPendingTransactionTransaction = &model.Transaction{
		ID:      1,
		BlockID: 1,
		Height:  720,
		SenderAccountAddress: []byte{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
	}
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
)

func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseFail) ParseTransactionBytes(
	transactionBytes []byte, sign bool,
) (*model.Transaction, error) {
	return nil, errors.New("mockedError")
}
func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess) ParseTransactionBytes(
	transactionBytes []byte, sign bool,
) (*model.Transaction, error) {
	return pendingTransactionHelperApplyUnconfirmedPendingTransactionTransaction, nil
}

func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherFail) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	return nil, errors.New("mockedError")
}

func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionFail) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperApplyUnconfirmedPendingTransactionActionTypeFail{}, nil
}

func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionSuccess) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperApplyUnconfirmedPendingTransactionActionTypeSuccess{}, nil
}

func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionActionTypeSuccess) ApplyUnconfirmed() error {
	return nil
}
func (*pendingTransactionhelperApplyUnconfirmedPendingTransactionActionTypeFail) ApplyUnconfirmed() error {
	return errors.New("mockedError")
}

func TestPendingTransactionHelper_ApplyUnconfirmedPendingTransaction(t *testing.T) {
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		pendingTransactionBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ApplyUnconfirmedPendingTransaction - parseFail",
			fields: fields{
				TransactionUtil: &pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseFail{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - getTransactionTypeFail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherFail{},
				TransactionUtil: &pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - applyUnconfirmedFail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionFail{},
				TransactionUtil: &pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - applyUnconfirmedSuccess",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionSuccess{},
				TransactionUtil: &pendingTransactionhelperApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pth := &PendingTransactionHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				TransactionUtil:         tt.fields.TransactionUtil,
				TypeSwitcher:            tt.fields.TypeSwitcher,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := pth.ApplyUnconfirmedPendingTransaction(tt.args.pendingTransactionBytes); (err != nil) != tt.wantErr {
				t.Errorf("ApplyUnconfirmedPendingTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// start here

type (
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
	pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseSuccess struct {
		Util
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseFail struct {
		Util
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionUndoFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionApplyConfirmedFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionApplyConfirmedSuccess struct {
		TypeSwitcher
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedFail struct {
		TypeAction
	}
	pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedSuccess struct {
		TypeAction
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeUndoUnconfirmedFail struct {
		TypeAction
	}

	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
)

var (
	mockApplyConfirmedPendingTransactionParseSuccess = &model.Transaction{}
)

func (*pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseFail) ParseTransactionBytes(
	transactionBytes []byte, sign bool) (*model.Transaction, error) {
	return nil, errors.New("mockedError")
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseSuccess) ParseTransactionBytes(
	transactionBytes []byte, sign bool) (*model.Transaction, error) {
	return mockApplyConfirmedPendingTransactionParseSuccess, nil
}

func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeUndoUnconfirmedFail) UndoApplyUnconfirmed() error {
	return errors.New("mockError")
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherFail) GetTransactionType(tx *model.Transaction) (TypeAction, error) {
	return nil, errors.New("mockedError")
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionUndoFail) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeUndoUnconfirmedFail{}, nil
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionApplyConfirmedFail) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedFail{}, nil
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionApplyConfirmedSuccess) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedSuccess{}, nil
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedSuccess) UndoApplyUnconfirmed() error {
	return nil
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedSuccess) ApplyConfirmed(int64) error {
	return nil
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedFail) UndoApplyUnconfirmed() error {
	return nil
}

func (*pendingTransactionhelperApplyConfirmedPendingTransactionActionTypeApplyConfirmedFail) ApplyConfirmed(int64) error {
	return errors.New("mockedError")
}

func TestPendingTransactionHelper_ApplyConfirmedPendingTransaction(t *testing.T) {
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		pendingTransactionBytes []byte
		txHeight                uint32
		blockTimestamp          int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "ApplyUnconfirmedPendingTransaction - parseFail",
			fields: fields{
				TransactionUtil: &pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseFail{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - getTransactionTypeFail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherFail{},
				TransactionUtil: &pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			want:    mockApplyConfirmedPendingTransactionParseSuccess,
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - UndoApplyUnconfirmedFail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionUndoFail{},
				TransactionUtil: &pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			want:    mockApplyConfirmedPendingTransactionParseSuccess,
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - apply confirmed fail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionApplyConfirmedFail{},
				TransactionUtil: &pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			want:    mockApplyConfirmedPendingTransactionParseSuccess,
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmedPendingTransaction - apply confirmed success",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperApplyConfirmedPendingTransactionTypeSwitcherSuccessTypeActionApplyConfirmedSuccess{},
				TransactionUtil: &pendingTransactionhelperApplyConfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			want:    mockApplyConfirmedPendingTransactionParseSuccess,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pth := &PendingTransactionHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				TransactionUtil:         tt.fields.TransactionUtil,
				TypeSwitcher:            tt.fields.TypeSwitcher,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			got, err := pth.ApplyConfirmedPendingTransaction(tt.args.pendingTransactionBytes, tt.args.txHeight, tt.args.blockTimestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyConfirmedPendingTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyConfirmedPendingTransaction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// mock multisignatureTransactionValidate
	mockMultisignatureValidateMultisigUtilValidateFail struct {
		MultisigTransactionUtilInterface
	}
	mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessPendingTxFail struct {
		MultisigTransactionUtilInterface
	}
	mockMultisignatureValidateMultisigUtilValidatePendingTxFail struct {
		MultisigTransactionUtilInterface
	}
	mockMultisignatureValidateMultisigUtilValidatePendingTxSuccessValidateSignatureInfoFail struct {
		MultisigTransactionUtilInterface
	}
	mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessSignatureInfoFail struct {
		MultisigTransactionUtilInterface
	}

	mockMultisignatureValidateMultisigUtilPendingTransactionHelperNoPendingTransaction struct {
		PendingTransactionHelperInterface
	}

	mockMultisignatureValidateMultisigUtilPendingTransactionHelperErrorPendingTransaction struct {
		PendingTransactionHelperInterface
	}
	mockMultisignatureValidateMultisigUtilPendingTransactionHelperPendingTransactionExist struct {
		PendingTransactionHelperInterface
	}

	mockMultisignatureValidateMultisigInfoNotExist struct {
		MultisignatureInfoHelperInterface
	}
	mockMultisignatureValidateMultisigInfoError struct {
		MultisignatureInfoHelperInterface
	}
	mockMultisignatureValidateMultisigInfoExist struct {
		MultisignatureInfoHelperInterface
	}
	mockValidateMultisigUtilValidateSignatureInfoSucess struct {
		MultisigTransactionUtilInterface
	}

	mockAccountBalanceHelperMultisignatureValidateSuccess struct {
		AccountBalanceHelper
	}
)

var (
	mockFeeMultisignatureValidate int64 = 10
)

func (*mockValidateMultisigUtilValidateSignatureInfoSucess) ValidateSignatureInfo(
	signature crypto.SignatureInterface, signatureInfo *model.SignatureInfo, multisignatureAddresses map[string]bool,
) error {
	return nil
}

func (*mockMultisignatureValidateMultisigInfoNotExist) GetMultisigInfoByAddress(
	multisigInfo *model.MultiSignatureInfo,
	multisigAddress []byte,
	blockHeight uint32,
) error {
	return sql.ErrNoRows
}

func (*mockMultisignatureValidateMultisigInfoError) GetMultisigInfoByAddress(
	multisigInfo *model.MultiSignatureInfo,
	multisigAddress []byte,
	blockHeight uint32,
) error {
	return errors.New("mockedError")
}

func (*mockMultisignatureValidateMultisigInfoExist) GetMultisigInfoByAddress(
	multisigInfo *model.MultiSignatureInfo,
	multisigAddress []byte,
	blockHeight uint32,
) error {
	*multisigInfo = model.MultiSignatureInfo{
		Addresses: make([][]byte, 3),
	}
	return nil
}

func (*mockMultisignatureValidateMultisigUtilPendingTransactionHelperErrorPendingTransaction) GetPendingTransactionByHash(
	pendingTransaction *model.PendingTransaction,
	pendingTransactionHash []byte,
	pendingTransactionStatuses []model.PendingTransactionStatus,
	blockHeight uint32,
	dbTx bool,
) error {
	return errors.New("mockedError")
}

func (*mockMultisignatureValidateMultisigUtilPendingTransactionHelperNoPendingTransaction) GetPendingTransactionByHash(
	pendingTransaction *model.PendingTransaction,
	pendingTransactionHash []byte,
	pendingTransactionStatuses []model.PendingTransactionStatus,
	blockHeight uint32,
	dbTx bool,
) error {
	pendingTransaction.TransactionBytes = make([]byte, 0)
	return nil
}

func (*mockMultisignatureValidateMultisigUtilPendingTransactionHelperPendingTransactionExist) GetPendingTransactionByHash(
	pendingTransaction *model.PendingTransaction,
	pendingTransactionHash []byte,
	pendingTransactionStatuses []model.PendingTransactionStatus,
	blockHeight uint32,
	dbTx bool,
) error {
	pendingTransaction.TransactionBytes = make([]byte, 32)
	return nil
}

func (*mockMultisignatureValidateMultisigUtilValidateFail) ValidateMultisignatureInfo(info *model.MultiSignatureInfo) error {
	return errors.New("mockedError")
}

func (*mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessPendingTxFail) ValidateMultisignatureInfo(
	info *model.MultiSignatureInfo) error {
	return nil
}

func (*mockMultisignatureValidateMultisigUtilValidatePendingTxFail) ValidatePendingTransactionBytes(
	transactionUtil UtilInterface,
	typeSwitcher TypeActionSwitcher,
	multisigInfoHelper MultisignatureInfoHelperInterface,
	pendingTransactionHelper PendingTransactionHelperInterface,
	multisigInfo *model.MultiSignatureInfo,
	senderAddress, unsignedTxBytes []byte,
	blockHeight uint32,
	dbTx bool,
) error {
	return errors.New("mockedError")
}

func (*mockMultisignatureValidateMultisigUtilValidatePendingTxSuccessValidateSignatureInfoFail) ValidatePendingTransactionBytes(
	transactionUtil UtilInterface,
	typeSwitcher TypeActionSwitcher,
	multisigInfoHelper MultisignatureInfoHelperInterface,
	pendingTransactionHelper PendingTransactionHelperInterface,
	multisigInfo *model.MultiSignatureInfo,
	senderAddress, unsignedTxBytes []byte,
	blockHeight uint32,
	dbTx bool,
) error {
	*multisigInfo = model.MultiSignatureInfo{
		Addresses: make([][]byte, 2),
	}
	return nil
}

func (*mockMultisignatureValidateMultisigUtilValidatePendingTxSuccessValidateSignatureInfoFail) ValidateSignatureInfo(
	signature crypto.SignatureInterface, signatureInfo *model.SignatureInfo, multisignatureAddresses map[string]bool,
) error {
	return errors.New("mockedError")
}

func (*mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessPendingTxFail) ValidatePendingTransactionBytes(
	transactionUtil UtilInterface,
	typeSwitcher TypeActionSwitcher,
	multisigInfoHelper MultisignatureInfoHelperInterface,
	pendingTransactionHelper PendingTransactionHelperInterface,
	multisigInfo *model.MultiSignatureInfo,
	senderAddress, unsignedTxBytes []byte,
	blockHeight uint32,
	dbTx bool,
) error {
	return errors.New("mockedError")
}

func (*mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessSignatureInfoFail) ValidateMultisignatureInfo(
	info *model.MultiSignatureInfo) error {
	return nil
}

func (*mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessSignatureInfoFail) ValidateSignatureInfo(
	signature crypto.SignatureInterface, signatureInfo *model.SignatureInfo, multisignatureAddresses map[string]bool,
) error {
	return errors.New("mockedError")
}

func (*mockAccountBalanceHelperMultisignatureValidateSuccess) GetBalanceByAccountAddress(
	accountBalance *model.AccountBalance, address []byte, dbTx bool,
) error {
	accountBalance.SpendableBalance = mockFeeMultisignatureValidate + 1
	return nil
}

func TestMultiSignatureTransaction_Validate(t *testing.T) {
	type fields struct {
		TransactionObject        *model.Transaction
		Body                     *model.MultiSignatureTransactionBody
		NormalFee                fee.FeeModelInterface
		TransactionUtil          UtilInterface
		TypeSwitcher             TypeActionSwitcher
		Signature                crypto.SignatureInterface
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		TransactionHelper        TransactionHelperInterface
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
			name: "Validate - none provided",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            nil,
				},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:exist - multisignatureInfo invalid",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       &model.MultiSignatureInfo{},
					UnsignedTransactionBytes: nil,
					SignatureInfo:            nil,
				},
				MultisigUtil:         &mockMultisignatureValidateMultisigUtilValidateFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:exist - multisignatureInfo valid - unsignedTransactionBytes invalid",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       &model.MultiSignatureInfo{},
					UnsignedTransactionBytes: make([]byte, 32),
					SignatureInfo:            nil,
				},
				MultisigUtil:         &mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessPendingTxFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:exist - multisignatureInfo valid - " +
				"signatureInfo invalid",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       &model.MultiSignatureInfo{},
					UnsignedTransactionBytes: nil,
					SignatureInfo:            &model.SignatureInfo{},
				},
				MultisigUtil:         &mockMultisignatureValidateMultisigUtilValidateMultisigInfoSuccessSignatureInfoFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - unsignedTransactionBytes invalid",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 32),
					SignatureInfo:            nil,
				},
				MultisigUtil:         &mockMultisignatureValidateMultisigUtilValidatePendingTxFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - unsignedTransactionBytes valid and return multisigInfo - " +
				"signatureInfo invalid",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 32),
					SignatureInfo:            &model.SignatureInfo{},
				},
				MultisigUtil:         &mockMultisignatureValidateMultisigUtilValidatePendingTxSuccessValidateSignatureInfoFail{},
				AccountBalanceHelper: &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - error getting pending transaction - " +
				"signatureInfo provided",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            &model.SignatureInfo{},
				},
				PendingTransactionHelper: &mockMultisignatureValidateMultisigUtilPendingTransactionHelperErrorPendingTransaction{},
				AccountBalanceHelper:     &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - no pending transaction - " +
				"signatureInfo provided",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            &model.SignatureInfo{},
				},
				PendingTransactionHelper: &mockMultisignatureValidateMultisigUtilPendingTransactionHelperNoPendingTransaction{},
				AccountBalanceHelper:     &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - pending transaction exist - " +
				"multisigInfo not exist",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            &model.SignatureInfo{},
				},
				PendingTransactionHelper: &mockMultisignatureValidateMultisigUtilPendingTransactionHelperPendingTransactionExist{},
				MultisignatureInfoHelper: &mockMultisignatureValidateMultisigInfoNotExist{},
				AccountBalanceHelper:     &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - pending transaction exist - " +
				"get multisigInfo error",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            &model.SignatureInfo{},
				},
				PendingTransactionHelper: &mockMultisignatureValidateMultisigUtilPendingTransactionHelperPendingTransactionExist{},
				MultisignatureInfoHelper: &mockMultisignatureValidateMultisigInfoError{},
				AccountBalanceHelper:     &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: true,
		},
		{
			name: "Validate - multisignatureInfo:notExist - pending transaction exist - " +
				"get multisigInfo exist - ValidateSignatureSuccess",
			fields: fields{
				TransactionObject: &model.Transaction{
					Fee: mockFeeMultisignatureValidate,
				},
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            &model.SignatureInfo{},
				},
				PendingTransactionHelper: &mockMultisignatureValidateMultisigUtilPendingTransactionHelperPendingTransactionExist{},
				MultisignatureInfoHelper: &mockMultisignatureValidateMultisigInfoExist{},
				MultisigUtil:             &mockValidateMultisigUtilValidateSignatureInfoSucess{},
				AccountBalanceHelper:     &mockAccountBalanceHelperMultisignatureValidateSuccess{},
			},
			args: args{
				dbTx: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				TransactionObject:        tt.fields.TransactionObject,
				Body:                     tt.fields.Body,
				NormalFee:                tt.fields.NormalFee,
				TransactionUtil:          tt.fields.TransactionUtil,
				TypeSwitcher:             tt.fields.TypeSwitcher,
				Signature:                tt.fields.Signature,
				MultisigUtil:             tt.fields.MultisigUtil,
				SignatureInfoHelper:      tt.fields.SignatureInfoHelper,
				MultisignatureInfoHelper: tt.fields.MultisignatureInfoHelper,
				PendingTransactionHelper: tt.fields.PendingTransactionHelper,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				TransactionHelper:        tt.fields.TransactionHelper,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockAccountBalanceHelperAddAccountSpendableBalanceFail struct {
		AccountBalanceHelperInterface
	}
	mockAccountBalanceHelperAddAccountSpendableBalanceSuccess struct {
		AccountBalanceHelperInterface
	}
	mockUndoApplyUnconfirmedPendingTransactionHelperUndoPendingFail struct {
		PendingTransactionHelperInterface
	}
	mockUndoApplyUnconfirmedPendingTransactionHelperUndoPendingSuccess struct {
		PendingTransactionHelperInterface
	}
)

func (*mockAccountBalanceHelperAddAccountSpendableBalanceFail) AddAccountSpendableBalance(
	address []byte, amount int64) error {
	return errors.New("mockedError")
}

func (*mockAccountBalanceHelperAddAccountSpendableBalanceSuccess) AddAccountSpendableBalance(
	address []byte, amount int64) error {
	return nil
}

func (*mockUndoApplyUnconfirmedPendingTransactionHelperUndoPendingFail) UndoApplyUnconfirmedPendingTransaction(
	pendingTransactionBytes []byte) error {
	return errors.New("mockedError")
}

func (*mockUndoApplyUnconfirmedPendingTransactionHelperUndoPendingSuccess) UndoApplyUnconfirmedPendingTransaction(
	pendingTransactionBytes []byte) error {
	return nil
}

func TestMultiSignatureTransaction_UndoApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject        *model.Transaction
		Body                     *model.MultiSignatureTransactionBody
		NormalFee                fee.FeeModelInterface
		TransactionUtil          UtilInterface
		TypeSwitcher             TypeActionSwitcher
		Signature                crypto.SignatureInterface
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		TransactionHelper        TransactionHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmed - AddAccountSpendableBalance fail",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperAddAccountSpendableBalanceFail{},
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed - AddAccountSpendableBalance success",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperAddAccountSpendableBalanceSuccess{},
				Body: &model.MultiSignatureTransactionBody{
					UnsignedTransactionBytes: make([]byte, 0),
				},
			},
			wantErr: false,
		},
		{
			name: "UndoApplyUnconfirmed - AddAccountSpendableBalance success, UndoPendingTransactionFail",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperAddAccountSpendableBalanceSuccess{},
				Body: &model.MultiSignatureTransactionBody{
					UnsignedTransactionBytes: make([]byte, 32),
				},
				PendingTransactionHelper: &mockUndoApplyUnconfirmedPendingTransactionHelperUndoPendingFail{},
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmed - AddAccountSpendableBalance success, UndoPendingTransactionSuccess",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperAddAccountSpendableBalanceSuccess{},
				Body: &model.MultiSignatureTransactionBody{
					UnsignedTransactionBytes: make([]byte, 32),
				},
				PendingTransactionHelper: &mockUndoApplyUnconfirmedPendingTransactionHelperUndoPendingSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				TransactionObject:        tt.fields.TransactionObject,
				Body:                     tt.fields.Body,
				NormalFee:                tt.fields.NormalFee,
				TransactionUtil:          tt.fields.TransactionUtil,
				TypeSwitcher:             tt.fields.TypeSwitcher,
				Signature:                tt.fields.Signature,
				MultisigUtil:             tt.fields.MultisigUtil,
				SignatureInfoHelper:      tt.fields.SignatureInfoHelper,
				MultisignatureInfoHelper: tt.fields.MultisignatureInfoHelper,
				PendingTransactionHelper: tt.fields.PendingTransactionHelper,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				TransactionHelper:        tt.fields.TransactionHelper,
			}
			if err := tx.UndoApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("UndoApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockApplyUnconfirmedPendingTransactionHelperApplyUnconfirmedFail struct {
		PendingTransactionHelperInterface
	}
	mockApplyUnconfirmedPendingTransactionHelperApplyUnconfirmedSuccess struct {
		PendingTransactionHelperInterface
	}
)

func (*mockApplyUnconfirmedPendingTransactionHelperApplyUnconfirmedFail) ApplyUnconfirmedPendingTransaction(
	pendingTransactionBytes []byte) error {
	return errors.New("mockedError")
}

func (*mockApplyUnconfirmedPendingTransactionHelperApplyUnconfirmedSuccess) ApplyUnconfirmedPendingTransaction(
	pendingTransactionBytes []byte) error {
	return nil
}

func TestMultiSignatureTransaction_ApplyUnconfirmed(t *testing.T) {
	type fields struct {
		TransactionObject        *model.Transaction
		Body                     *model.MultiSignatureTransactionBody
		NormalFee                fee.FeeModelInterface
		TransactionUtil          UtilInterface
		TypeSwitcher             TypeActionSwitcher
		Signature                crypto.SignatureInterface
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		TransactionHelper        TransactionHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ApplyUnconfirmed - AddAccountSpendableBalance-Fail",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperAddAccountSpendableBalanceFail{},
			},
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmed - AddAccountSpendableBalance-Success",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					UnsignedTransactionBytes: make([]byte, 0),
				},
				AccountBalanceHelper: &mockAccountBalanceHelperAddAccountSpendableBalanceSuccess{},
			},
			wantErr: false,
		},
		{
			name: "ApplyUnconfirmed - ApplyUnconfirmedPendingTransactionFail",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					UnsignedTransactionBytes: make([]byte, 32),
				},
				AccountBalanceHelper:     &mockAccountBalanceHelperAddAccountSpendableBalanceSuccess{},
				PendingTransactionHelper: &mockApplyUnconfirmedPendingTransactionHelperApplyUnconfirmedFail{},
			},
			wantErr: true,
		},
		{
			name: "ApplyUnconfirmed - ApplyUnconfirmedPendingTransactionSuccess",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					UnsignedTransactionBytes: make([]byte, 32),
				},
				AccountBalanceHelper:     &mockAccountBalanceHelperAddAccountSpendableBalanceSuccess{},
				PendingTransactionHelper: &mockApplyUnconfirmedPendingTransactionHelperApplyUnconfirmedSuccess{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				TransactionObject:        tt.fields.TransactionObject,
				Body:                     tt.fields.Body,
				NormalFee:                tt.fields.NormalFee,
				TransactionUtil:          tt.fields.TransactionUtil,
				TypeSwitcher:             tt.fields.TypeSwitcher,
				Signature:                tt.fields.Signature,
				MultisigUtil:             tt.fields.MultisigUtil,
				SignatureInfoHelper:      tt.fields.SignatureInfoHelper,
				MultisignatureInfoHelper: tt.fields.MultisignatureInfoHelper,
				PendingTransactionHelper: tt.fields.PendingTransactionHelper,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				TransactionHelper:        tt.fields.TransactionHelper,
			}
			if err := tx.ApplyUnconfirmed(); (err != nil) != tt.wantErr {
				t.Errorf("ApplyUnconfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// here

type (
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess struct {
		Util
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseFail struct {
		Util
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionFail struct {
		TypeSwitcher
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionSuccess struct {
		TypeSwitcher
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeFail struct {
		TypeAction
	}
	pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeSuccess struct {
		TypeAction
	}
	// mock pendingTransactionHelperApplyUnconfirmedPendingTransaction
)

func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseFail) ParseTransactionBytes(
	transactionBytes []byte, sign bool,
) (*model.Transaction, error) {
	return nil, errors.New("mockedError")
}
func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess) ParseTransactionBytes(
	transactionBytes []byte, sign bool,
) (*model.Transaction, error) {
	return pendingTransactionHelperApplyUnconfirmedPendingTransactionTransaction, nil
}

func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherFail) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return nil, errors.New("mockedError")
}

func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionFail) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeFail{}, nil
}

func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionSuccess) GetTransactionType(
	tx *model.Transaction) (TypeAction, error) {
	return &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeSuccess{}, nil
}

func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeSuccess) UndoApplyUnconfirmed() error {
	return nil
}
func (*pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionActionTypeFail) UndoApplyUnconfirmed() error {
	return errors.New("mockedError")
}

func TestPendingTransactionHelper_UndoApplyUnconfirmedPendingTransaction(t *testing.T) {
	type fields struct {
		MultisignatureInfoQuery query.MultisignatureInfoQueryInterface
		PendingTransactionQuery query.PendingTransactionQueryInterface
		TransactionUtil         UtilInterface
		TypeSwitcher            TypeActionSwitcher
		QueryExecutor           query.ExecutorInterface
	}
	type args struct {
		pendingTransactionBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "UndoApplyUnconfirmedPendingTransaction - parseFail",
			fields: fields{
				TransactionUtil: &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseFail{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmedPendingTransaction - getTransactionTypeFail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherFail{},
				TransactionUtil: &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmedPendingTransaction - undoApplyUnconfirmedFail",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionFail{},
				TransactionUtil: &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "UndoApplyUnconfirmedPendingTransaction - applyUnconfirmedSuccess",
			fields: fields{
				TypeSwitcher:    &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTypeSwitcherSuccessTypeActionSuccess{},
				TransactionUtil: &pendingTransactionhelperUndoApplyUnconfirmedPendingTransactionTransactionUtilParseSuccess{},
			},
			args: args{
				pendingTransactionBytes: make([]byte, 32),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pth := &PendingTransactionHelper{
				MultisignatureInfoQuery: tt.fields.MultisignatureInfoQuery,
				PendingTransactionQuery: tt.fields.PendingTransactionQuery,
				TransactionUtil:         tt.fields.TransactionUtil,
				TypeSwitcher:            tt.fields.TypeSwitcher,
				QueryExecutor:           tt.fields.QueryExecutor,
			}
			if err := pth.UndoApplyUnconfirmedPendingTransaction(tt.args.pendingTransactionBytes); (err != nil) != tt.wantErr {
				t.Errorf("UndoApplyUnconfirmedPendingTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultiSignatureTransaction_GetBodyBytes(t *testing.T) {
	type fields struct {
		TransactionObject        *model.Transaction
		Body                     *model.MultiSignatureTransactionBody
		NormalFee                fee.FeeModelInterface
		TransactionUtil          UtilInterface
		TypeSwitcher             TypeActionSwitcher
		Signature                crypto.SignatureInterface
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		TransactionHelper        TransactionHelperInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "GetMultisignatureBodyBytes - success",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 2,
						Nonce:             1,
						Addresses: [][]byte{
							{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
								45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
							{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
								81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
							{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98,
								47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						},
						MultisigAddress: mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
						BlockHeight:     720,
						Latest:          true,
					},
					UnsignedTransactionBytes: make([]byte, 64),
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures: map[string][]byte{
							"00000000b2df80b333966806b585b979a38b3378f60ffa38769fa6616228468223a468b6": make([]byte, 32),
						},
					},
				},
			},
			want: []byte{
				1, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125,
				75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88,
				220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 0, 0,
				0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77,
				206, 170, 27, 124, 232, 110, 14, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0,
				0, 178, 223, 128, 179, 51, 150, 104, 6, 181, 133, 185, 121, 163, 139, 51, 120, 246, 15, 250, 56, 118, 159, 166, 97, 98,
				40, 70, 130, 35, 164, 104, 182, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
		},
		{
			name: "GetMultisignatureBodyBytes - success - multisigInfo missing",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 64),
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures: map[string][]byte{
							"00000000b2df80b333966806b585b979a38b3378f60ffa38769fa6616228468223a468b6": make([]byte, 32),
						},
					},
				},
			},
			want: []byte{
				0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 178, 223, 128, 179,
				51, 150, 104, 6, 181, 133, 185, 121, 163, 139, 51, 120, 246, 15, 250, 56, 118, 159, 166, 97, 98, 40, 70, 130, 35, 164, 104,
				182, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name: "GetMultisignatureBodyBytes - success - signatureInfo missing",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 2,
						Nonce:             1,
						Addresses: [][]byte{
							{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
								45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
							{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
								81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
							{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98,
								47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						},
						MultisigAddress: mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
						BlockHeight:     720,
						Latest:          true,
					},
					UnsignedTransactionBytes: make([]byte, 64),
					SignatureInfo:            nil,
				},
			},
			want: []byte{
				1, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125,
				75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88,
				220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 0, 0,
				0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77,
				206, 170, 27, 124, 232, 110, 14, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0,
			},
		},
		{
			name: "GetMultisignatureBodyBytes - fail",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 2,
						Nonce:             1,
						Addresses: [][]byte{
							{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
								45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
							{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
								81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
							{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98,
								47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						},
						MultisigAddress: mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
						BlockHeight:     720,
						Latest:          true,
					},
					UnsignedTransactionBytes: make([]byte, 64),
					SignatureInfo: &model.SignatureInfo{
						Signatures: map[string][]byte{
							"0000IamAnInvalidAccount": make([]byte, 64),
						},
						TransactionHash: make([]byte, 32),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GetMultisignatureBodyBytes - successParsingSignatures",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 2,
						Nonce:             1,
						Addresses: [][]byte{
							{4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
								45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
							{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
								81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
							{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98,
								47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14},
						},
						MultisigAddress: mockMultisignatureInfoHelperMultisigInfoSuccess.MultisigAddress,
						BlockHeight:     720,
						Latest:          true,
					},
					UnsignedTransactionBytes: make([]byte, 64),
					SignatureInfo: &model.SignatureInfo{
						Signatures: map[string][]byte{
							"00000000b2df80b333966806b585b979a38b3378f60ffa38769fa6616228468223a468b6": make([]byte, 64),
						},
						TransactionHash: make([]byte, 32),
					},
				},
			},
			want: []byte{1, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21,
				99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 0, 0, 0, 0, 4, 38, 68, 24, 230,
				247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28,
				169, 0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243,
				132, 77, 206, 170, 27, 124, 232, 110, 14, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0,
				0, 0, 0, 0, 178, 223, 128, 179, 51, 150, 104, 6, 181, 133, 185, 121, 163, 139, 51, 120, 246, 15, 250, 56, 118, 159, 166,
				97, 98, 40, 70, 130, 35, 164, 104, 182, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				TransactionObject:        tt.fields.TransactionObject,
				Body:                     tt.fields.Body,
				TypeSwitcher:             tt.fields.TypeSwitcher,
				Signature:                tt.fields.Signature,
				MultisignatureInfoHelper: tt.fields.MultisignatureInfoHelper,
				PendingTransactionHelper: tt.fields.PendingTransactionHelper,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				TransactionHelper:        tt.fields.TransactionHelper,
			}
			got, err := tx.GetBodyBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiSignatureTransaction_ParseBodyBytes(t *testing.T) {
	var (
		multisigTxBody = &model.MultiSignatureTransactionBody{
			MultiSignatureInfo: &model.MultiSignatureInfo{
				MinimumSignatures: 2,
				Nonce:             1,
				Addresses: [][]byte{
					senderAddress1,
					senderAddress2,
					senderAddress3,
				},
			},
			UnsignedTransactionBytes: make([]byte, 64),
			SignatureInfo: &model.SignatureInfo{
				TransactionHash: make([]byte, 32),
				Signatures: map[string][]byte{
					"00000000b2df80b333966806b585b979a38b3378f60ffa38769fa6616228468223a468b6": make([]byte, 64),
				},
			},
		}
		tx1 = &MultiSignatureTransaction{
			TransactionObject: &model.Transaction{
				ID:                   1390544043583530800,
				SenderAccountAddress: senderAddress1,
				Fee:                  1,
			},
			Body: multisigTxBody,
		}
		multisigTx1BodyBytes, _ = tx1.GetBodyBytes()
	)

	type fields struct {
		TransactionObject        *model.Transaction
		Body                     *model.MultiSignatureTransactionBody
		NormalFee                fee.FeeModelInterface
		TransactionUtil          UtilInterface
		TypeSwitcher             TypeActionSwitcher
		Signature                crypto.SignatureInterface
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		TransactionHelper        TransactionHelperInterface
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
			name: "parseBodyBytes - complete",
			fields: fields{
				TransactionObject: &model.Transaction{
					ID:                   1390544043583530800,
					SenderAccountAddress: senderAddress1,
					Fee:                  1,
					BlockID:              int64(111),
					Height:               uint32(10),
				},
				Body: multisigTxBody,
			},
			args: args{
				txBodyBytes: multisigTx1BodyBytes,
			},
			want:    multisigTxBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				TransactionObject:        tt.fields.TransactionObject,
				Body:                     tt.fields.Body,
				NormalFee:                tt.fields.NormalFee,
				TransactionUtil:          tt.fields.TransactionUtil,
				TypeSwitcher:             tt.fields.TypeSwitcher,
				Signature:                tt.fields.Signature,
				MultisigUtil:             tt.fields.MultisigUtil,
				SignatureInfoHelper:      tt.fields.SignatureInfoHelper,
				MultisignatureInfoHelper: tt.fields.MultisignatureInfoHelper,
				PendingTransactionHelper: tt.fields.PendingTransactionHelper,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				TransactionHelper:        tt.fields.TransactionHelper,
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

func TestMultiSignatureTransaction_GetSize(t *testing.T) {
	type fields struct {
		TransactionObject        *model.Transaction
		Body                     *model.MultiSignatureTransactionBody
		NormalFee                fee.FeeModelInterface
		TransactionUtil          UtilInterface
		TypeSwitcher             TypeActionSwitcher
		Signature                crypto.SignatureInterface
		MultisigUtil             MultisigTransactionUtilInterface
		SignatureInfoHelper      SignatureInfoHelperInterface
		MultisignatureInfoHelper MultisignatureInfoHelperInterface
		PendingTransactionHelper PendingTransactionHelperInterface
		AccountBalanceHelper     AccountBalanceHelperInterface
		TransactionHelper        TransactionHelperInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSizeComplete",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 2,
						Nonce:             1,
						Addresses: [][]byte{
							senderAddress1,
							senderAddress2,
							senderAddress3,
						},
					},
					UnsignedTransactionBytes: make([]byte, 64),
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures: map[string][]byte{
							"00000000b2df80b333966806b585b979a38b3378f60ffa38769fa6616228468223a468b6": make([]byte, 32),
						},
					},
				},
			},
			want: 360,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				TransactionObject:        tt.fields.TransactionObject,
				Body:                     tt.fields.Body,
				NormalFee:                tt.fields.NormalFee,
				TransactionUtil:          tt.fields.TransactionUtil,
				TypeSwitcher:             tt.fields.TypeSwitcher,
				Signature:                tt.fields.Signature,
				MultisigUtil:             tt.fields.MultisigUtil,
				SignatureInfoHelper:      tt.fields.SignatureInfoHelper,
				MultisignatureInfoHelper: tt.fields.MultisignatureInfoHelper,
				PendingTransactionHelper: tt.fields.PendingTransactionHelper,
				AccountBalanceHelper:     tt.fields.AccountBalanceHelper,
				TransactionHelper:        tt.fields.TransactionHelper,
			}
			if got, _ := tx.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
