package transaction

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"golang.org/x/crypto/sha3"
)

var (
	mockTxID                   int64 = 1390544043583530800
	mockTxTimestamp            int64 = 1581301507
	mockTxSenderAccountAddress       = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224,
		72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	mockTxRecipientAccountAddress = []byte{0, 0, 0, 0, 229, 176, 168, 71, 174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126,
		202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14}
	mockTxApproverAccountAddress = []byte{0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250, 58, 2, 83, 75, 213, 137, 66, 236, 188, 43,
		59, 241, 146, 243, 147, 58, 161, 35, 229, 54}
	mockTxBodyLength uint32 = 8
)

func TestGetTransactionBytes(t *testing.T) {
	var (
		mockTxSignedSuccess, mockTxSignedSuccessBytes = GetFixtureForSpecificTransaction(
			mockTxID,
			mockTxTimestamp,
			mockTxSenderAccountAddress,
			mockTxRecipientAccountAddress,
			mockTxBodyLength,
			model.TransactionType_SendMoneyTransaction,
			&model.SendMoneyTransactionBody{
				Amount: 10,
			},
			false,
			true,
		)
		mockTxSignedEscrowSuccess, mockTxSignedEscrowSuccessBytes = GetFixtureForSpecificTransaction(
			mockTxID,
			mockTxTimestamp,
			mockTxSenderAccountAddress,
			mockTxRecipientAccountAddress,
			mockTxBodyLength,
			model.TransactionType_SendMoneyTransaction,
			&model.SendMoneyTransactionBody{
				Amount: 10,
			},
			true,
			true,
		)
	)
	type args struct {
		transaction *model.Transaction
		sign        bool
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetTransactionBytes:success",
			args: args{
				transaction: mockTxSignedSuccess,
				sign:        true,
			},
			want:    mockTxSignedSuccessBytes,
			wantErr: false,
		},
		{
			name: "GetTransactionBytes:success-{without-signature}",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					TransactionType:         2,
					Timestamp:               1562806389280,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					Fee:                     1000000,
					TransactionBodyLength:   8,
					TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
				sign: false,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79,
				50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114,
				110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85,
				89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57,
				106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0,
				0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0,
			},
			wantErr: false,
		},
		{
			name: "GetTransactionBytes:fail-{sign:true, no signature}",
			args: args{
				transaction: &model.Transaction{
					TransactionType:         2,
					Version:                 1,
					Timestamp:               1562806389280,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					Fee:                     1000000,
					TransactionBodyLength:   8,
					TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
				sign: true,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetTransactionBytes:success-{without recipient}",
			args: args{
				transaction: &model.Transaction{
					Version:               1,
					TransactionType:       2,
					Timestamp:             1562806389280,
					SenderAccountAddress:  mockTxSenderAccountAddress,
					Fee:                   1000000,
					TransactionBodyLength: 8,
					TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				},
				sign: false,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102,
				79, 50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80,
				79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 0, 0, 0, 0, 64, 66, 15, 0, 0, 0, 0,
				0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
			wantErr: false,
		},
		{
			name: "Success:WithEscrow",
			args: args{
				transaction: mockTxSignedEscrowSuccess,
				sign:        true,
			},
			want: mockTxSignedEscrowSuccessBytes,
		},
		{
			name: "SuccessNoSigned:WithEscrow",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					TransactionType:         2,
					Timestamp:               1562806389280,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					Fee:                     1000000,
					TransactionBodyLength:   8,
					TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
					Escrow: &model.Escrow{
						ApproverAddress: mockTxApproverAccountAddress,
						Commission:      24,
						Timeout:         100,
					},
					Signature: []byte{0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174,
						239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
						139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				},
				sign: false,
			},
			want: []byte{
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51,
				99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80, 57, 56, 71, 69, 65, 85, 67, 55, 44, 0, 0, 0, 66, 67, 90,
				75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106,
				102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 44, 0, 0, 0, 66, 67, 90, 68,
				95, 86, 120, 102, 79, 50, 83, 57, 97, 122, 105, 73, 76, 51, 99, 110, 95, 99, 88, 87, 55, 117, 80, 68, 86, 80, 79, 114, 110, 88, 117, 80,
				57, 56, 71, 69, 65, 85, 67, 55, 24, 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name: "EscrowApproval",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					ID:                      1,
					BlockID:                 1,
					Height:                  1,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: nil,
					TransactionType:         4,
					Fee:                     1,
					Timestamp:               1562806389280,
					TransactionHash:         nil,
					TransactionBodyLength:   12,
					TransactionBodyBytes:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					TransactionIndex:        0,
					Signature:               nil,
					Escrow:                  nil,
				},
			},
			want: []byte{4, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 3, 0, 0, 0, 71, 72, 73, 0, 0, 0, 0, 1, 0, 0,
				0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "EscrowApproval:Signed",
			args: args{
				transaction: &model.Transaction{
					Version:                 1,
					ID:                      1,
					BlockID:                 1,
					Height:                  1,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: nil,
					TransactionType:         4,
					Fee:                     1,
					Timestamp:               1562806389280,
					TransactionHash:         nil,
					TransactionBodyLength:   12,
					TransactionBodyBytes:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					TransactionIndex:        0,
					Signature:               nil,
					Escrow:                  nil,
				},
			},
			want: []byte{4, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 3, 0, 0, 0, 71, 72, 73, 0, 0, 0, 0, 1, 0,
				0, 0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&Util{}).GetTransactionBytes(tt.args.transaction, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionBytes() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				var byteStrArr []string
				for _, bt := range got {
					byteStrArr = append(byteStrArr, fmt.Sprintf("%v", bt))
				}
				t.Logf(strings.Join(byteStrArr, ", "))
				t.Errorf("GetTransactionBytes() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockMempoolCacheStorageSuccessGet struct {
		storage.MempoolCacheStorage
	}
)

func (*mockMempoolCacheStorageSuccessGet) GetItem(key, item interface{}) error { return nil }

func TestParseTransactionBytes(t *testing.T) {
	var mockTransactionWithEscrow = &model.Transaction{
		ID:                      4870989829983641364,
		Version:                 1,
		TransactionType:         2,
		BlockID:                 0,
		Height:                  0,
		Timestamp:               1562806389280,
		SenderAccountAddress:    mockTxSenderAccountAddress,
		RecipientAccountAddress: mockTxRecipientAccountAddress,
		Fee:                     1000000,
		TransactionHash: []byte{
			59, 106, 191, 6, 145, 54, 181, 186, 75, 93, 234, 139, 131, 96, 153, 252, 40, 245, 235, 132,
			187, 45, 245, 113, 210, 87, 23, 67, 157, 117, 41, 143,
		},
		TransactionBodyLength: 8,
		TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Signature: []byte{
			0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
			68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
			77, 80, 80, 39, 254, 173, 28, 169,
		},
		Escrow: &model.Escrow{
			ApproverAddress: approverAddress1,
			Commission:      24,
			Timeout:         100,
		},
	}
	transactionWithEscrowBytes, transactionWithEscrowHashed := GetFixturesForTransactionBytes(mockTransactionWithEscrow, true)
	mockTransactionWithEscrow.TransactionHash = transactionWithEscrowHashed[:]

	approvalTX, approvalTXBytes := GetFixtureForSpecificTransaction(
		-5081269314054617420,
		12345678,
		senderAddress1,
		nil,
		constant.EscrowApprovalBytesLength,
		model.TransactionType_ApprovalEscrowTransaction,
		&model.ApprovalEscrowTransactionBody{
			Approval:      model.EscrowApproval_Approve,
			TransactionID: 0,
		},
		false,
		true,
	)

	successWithoutSig, successWithoutSigHashed := GetFixturesForTransactionBytes(&model.Transaction{
		ID:                      670925173877174625,
		Version:                 1,
		TransactionType:         2,
		BlockID:                 0,
		Height:                  0,
		Timestamp:               1562806389280,
		SenderAccountAddress:    mockTxSenderAccountAddress,
		RecipientAccountAddress: mockTxRecipientAccountAddress,
		Fee:                     1000000,
		TransactionHash: []byte{
			59, 106, 191, 6, 145, 54, 181, 186, 75, 93, 234, 139, 131, 96, 153, 252, 40, 245, 235, 132,
			187, 45, 245, 113, 210, 87, 23, 67, 157, 117, 41, 143,
		},
		TransactionBodyLength: 8,
		TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Signature: []byte{
			0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
			68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184,
			77, 80, 80, 39, 254, 173, 28, 169,
		},
		Escrow: &model.Escrow{
			ApproverAddress: approverAddress1,
			Commission:      24,
			Timeout:         100,
		},
	}, false)
	type args struct {
		transactionBytes []byte
		sign             bool
	}
	type fields struct {
		mempoolCacheStorage storage.CacheStorageInterface
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *model.Transaction
		wantErr bool
	}{
		{
			name: "ParseTransactionBytes:withEscrow",
			args: args{
				transactionBytes: transactionWithEscrowBytes,
				sign:             true,
			},
			fields: fields{
				mempoolCacheStorage: &mockMempoolCacheStorageSuccessGet{},
			},
			want:    mockTransactionWithEscrow,
			wantErr: false,
		},
		{
			name: "ParseTransactionBytes:success-{without-signature}",
			args: args{
				transactionBytes: successWithoutSig,
				sign:             false,
			},
			fields: fields{
				mempoolCacheStorage: &mockMempoolCacheStorageSuccessGet{},
			},
			want: &model.Transaction{
				ID:                      4956766951297472907,
				Version:                 1,
				TransactionType:         2,
				BlockID:                 0,
				Height:                  0,
				Timestamp:               1562806389280,
				SenderAccountAddress:    mockTxSenderAccountAddress,
				RecipientAccountAddress: mockTxRecipientAccountAddress,
				Fee:                     1000000,
				TransactionHash:         successWithoutSigHashed[:],
				TransactionBodyLength:   8,
				TransactionBodyBytes:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Escrow: &model.Escrow{
					ApproverAddress: approverAddress1,
					Commission:      24,
					Timeout:         100,
				},
			},
			wantErr: false,
		},
		{
			name: "Ups",
			args: args{
				transactionBytes: approvalTXBytes,
				sign:             true,
			},
			fields: fields{
				mempoolCacheStorage: &mockMempoolCacheStorageSuccessGet{},
			},
			want: approvalTX,
		},
		{
			name: "ParseTransactionBytes:fail",
			args: args{
				transactionBytes: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 44, 0, 0, 0, 66, 67, 90, 68, 95, 86, 120, 102, 79, 50, 83,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5,
					6, 7, 8, 93, 3},
				sign: true,
			},
			fields: fields{
				mempoolCacheStorage: &mockMempoolCacheStorageSuccessGet{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&Util{
				MempoolCacheStorage: tt.fields.mempoolCacheStorage,
			}).ParseTransactionBytes(tt.args.transactionBytes, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTransactionBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTransactionBytes() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestGetTransactionID(t *testing.T) {
	type args struct {
		tx *model.Transaction
		ct chaintype.ChainType
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "GetTransactionID:success",
			args: args{
				tx: &model.Transaction{
					TransactionHash: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					Signature:       make([]byte, 64),
				},
				ct: &chaintype.MainChain{},
			},
			wantErr: false,
			want:    72340172838076673,
		},
		{
			name: "GetTransactionID:fail",
			args: args{
				tx: &model.Transaction{
					TransactionHash: []byte{},
				},
				ct: &chaintype.MainChain{},
			},
			wantErr: true,
			want:    -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&Util{}).GetTransactionID(tt.args.tx.TransactionHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTransactionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockValidateTransactionFeeScaleServiceCache struct {
		fee.FeeScaleServiceInterface
	}
)

func (*mockValidateTransactionFeeScaleServiceCache) GetLatestFeeScale(feeScale *model.FeeScale) error {
	*feeScale = model.FeeScale{
		FeeScale:    constant.OneZBC,
		BlockHeight: 0,
		Latest:      true,
	}
	return nil
}

type (
	mockTypeActionValidateTransactionSuccess struct {
		TypeAction
	}
)

func (mockTypeActionValidateTransactionSuccess) GetMinimumFee() (int64, error) {
	return 0, nil
}

type (
	mockAccountDatasetQueryValidateTransaction struct {
		query.AccountDatasetQuery
		wantNoRow bool
	}
)

func (*mockAccountDatasetQueryValidateTransaction) GetAccountDatasetEscrowApproval(recipientAddress []byte) (qry string, args []interface{}) {
	return
}
func (m *mockAccountDatasetQueryValidateTransaction) Scan(dataset *model.AccountDataset, _ *sql.Row) error {
	if m.wantNoRow {
		return sql.ErrNoRows
	}
	*dataset = model.AccountDataset{
		SetterAccountAddress:    mockTxSenderAccountAddress,
		RecipientAccountAddress: mockTxRecipientAccountAddress,
		Property:                "Admin",
		Value:                   "You're Welcome",
		IsActive:                true,
		Latest:                  true,
		Height:                  5,
	}

	return nil
}

type mockQueryExecutorQueryValidateTransaction struct {
	query.Executor
	wantErr     bool
	wantErrType error
}

func (m *mockQueryExecutorQueryValidateTransaction) ExecuteSelectRow(qu string, tx bool, args ...interface{}) (*sql.Row, error) {
	if m.wantErr {
		if m.wantErrType == sql.ErrNoRows {
			db, mock, _ := sqlmock.New()
			mock.ExpectQuery(regexp.QuoteMeta(qu)).WillReturnError(sql.ErrNoRows)
			return db.QueryRow(qu), nil
		}
		return nil, m.wantErrType
	}

	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qu)).WillReturnRows(sqlmock.NewRows([]string{"column"}))
	return db.QueryRow(qu), nil
}

func TestUtil_ValidateTransaction(t *testing.T) {
	transactionUtil := &Util{
		FeeScaleService: &mockValidateTransactionFeeScaleServiceCache{},
	}
	txValidateNoRecipient := GetFixturesForTransaction(
		1562893303,
		senderAddress1,
		nil,
		true,
	)
	txBytesNoRecipient, _ := transactionUtil.GetTransactionBytes(txValidateNoRecipient, false)
	txBytesHash := sha3.Sum256(txBytesNoRecipient)
	signatureTXValidateNoRecipient, _ := (&crypto.Signature{}).Sign(txBytesHash[:], model.SignatureType_DefaultSignature,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	txValidateNoRecipient.Signature = signatureTXValidateNoRecipient

	txValidateMustEscrow := GetFixturesForTransaction(
		1562893303,
		senderAddress1,
		recipientAddress1,
		false,
	)
	txBytesMustEscrow, _ := transactionUtil.GetTransactionBytes(txValidateMustEscrow, false)
	txBytesMustEscrowHash := sha3.Sum256(txBytesMustEscrow)
	signatureTXValidateMustEscrow, _ := (&crypto.Signature{}).Sign(txBytesMustEscrowHash[:], model.SignatureType_DefaultSignature,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	txValidateMustEscrow.Signature = signatureTXValidateMustEscrow

	txValidateEscrow := GetFixturesForTransaction(
		1562893303,
		senderAddress1,
		recipientAddress1,
		true,
	)
	txBytesEscrow, _ := transactionUtil.GetTransactionBytes(txValidateEscrow, false)
	txBytesEscrowHash := sha3.Sum256(txBytesEscrow)
	signatureTXValidateEscrow, _ := (&crypto.Signature{}).Sign(txBytesEscrowHash[:], model.SignatureType_DefaultSignature,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	txValidateEscrow.Signature = signatureTXValidateEscrow

	type fields struct {
		FeeScaleService     fee.FeeScaleServiceInterface
		MempoolCacheStorage storage.CacheStorageInterface
		QueryExecutor       query.ExecutorInterface
		AccountDatasetQuery query.AccountDatasetQueryInterface
	}
	type args struct {
		tx              *model.Transaction
		typeAction      TypeAction
		verifySignature bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess:NoRecipient",
			fields: fields{
				FeeScaleService: &mockValidateTransactionFeeScaleServiceCache{},
			},
			args: args{
				tx: GetFixturesForTransaction(
					1562893303,
					senderAddress1,
					nil,
					false,
				),
				typeAction:      &mockTypeActionValidateTransactionSuccess{},
				verifySignature: false,
			},
		},
		{
			name: "wantSuccess:NoRecipientSign",
			fields: fields{
				FeeScaleService: &mockValidateTransactionFeeScaleServiceCache{},
			},
			args: args{
				tx:              txValidateNoRecipient,
				typeAction:      &mockTypeActionValidateTransactionSuccess{},
				verifySignature: true,
			},
		},
		{
			name: "wantError:MustEscrow",
			fields: fields{
				FeeScaleService:     &mockValidateTransactionFeeScaleServiceCache{},
				AccountDatasetQuery: &mockAccountDatasetQueryValidateTransaction{},
				QueryExecutor:       &mockQueryExecutorQueryValidateTransaction{},
			},
			args: args{
				tx:              txValidateMustEscrow,
				typeAction:      &mockTypeActionValidateTransactionSuccess{},
				verifySignature: true,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess:Escrow",
			fields: fields{
				FeeScaleService:     &mockValidateTransactionFeeScaleServiceCache{},
				AccountDatasetQuery: &mockAccountDatasetQueryValidateTransaction{},
				QueryExecutor:       &mockQueryExecutorQueryValidateTransaction{},
			},
			args: args{
				tx:              txValidateEscrow,
				typeAction:      &mockTypeActionValidateTransactionSuccess{},
				verifySignature: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Util{
				FeeScaleService:     tt.fields.FeeScaleService,
				MempoolCacheStorage: tt.fields.MempoolCacheStorage,
				QueryExecutor:       tt.fields.QueryExecutor,
				AccountDatasetQuery: tt.fields.AccountDatasetQuery,
			}
			if err := u.ValidateTransaction(tt.args.tx, tt.args.typeAction, tt.args.verifySignature); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUtil_GenerateMultiSigAddress(t *testing.T) {
	type args struct {
		info *model.MultiSignatureInfo
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{info: &model.MultiSignatureInfo{
				MinimumSignatures: 2,
				Nonce:             12,
				Addresses: [][]byte{
					senderAddress1,
					recipientAddress1,
					approverAddress1,
				},
			}},
			want: senderAddress1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tu := &Util{}
			got, err := tu.GenerateMultiSigAddress(tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMultiSigAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("GenerateMultiSigAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultisigTransactionUtil_ValidateMultisignatureInfo(t *testing.T) {
	type args struct {
		multisigInfo *model.MultiSignatureInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Multsig.Participants < 2",
			args: args{
				&model.MultiSignatureInfo{
					MinimumSignatures: 0,
					Nonce:             0,
					Addresses:         make([][]byte, 1),
				},
			},
			wantErr: true,
		},
		{
			name: "Multsig.MinSigs < 1",
			args: args{
				&model.MultiSignatureInfo{
					MinimumSignatures: 0,
					Nonce:             0,
					Addresses:         make([][]byte, 2),
				},
			},
			wantErr: true,
		},
		{
			name: "Multsig.MinSigs < 1",
			args: args{
				&model.MultiSignatureInfo{
					MinimumSignatures: 1,
					Nonce:             0,
					Addresses:         make([][]byte, 2),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &MultisigTransactionUtil{}
			if err := mu.ValidateMultisignatureInfo(tt.args.multisigInfo); (err != nil) != tt.wantErr {
				t.Errorf("ValidateMultisignatureInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultisigTransactionUtil_ValidateSignatureInfo(t *testing.T) {
	type args struct {
		signature                   crypto.SignatureInterface
		signatureInfo               *model.SignatureInfo
		multiSignatureInfoAddresses map[string]bool
	}
	sig := &crypto.Signature{}
	txHash := make([]byte, 32)
	_, _, _, validAddress, _, _ := sig.GenerateAccountFromSeed(&accounttype.ZbcAccountType{}, "a")
	validSignature, _ := sig.Sign(txHash, model.SignatureType_DefaultSignature, "a")
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ValidateSignatureInfo - transaction Hash not exist",
			args: args{
				signature: nil,
				signatureInfo: &model.SignatureInfo{
					TransactionHash: nil,
					Signatures:      nil,
				},
				multiSignatureInfoAddresses: nil,
			},
			wantErr: true,
		},
		{
			name: "ValidateSignatureInfo - signatures not provided",
			args: args{
				signature: nil,
				signatureInfo: &model.SignatureInfo{
					TransactionHash: txHash,
					Signatures:      make(map[string][]byte),
				},
				multiSignatureInfoAddresses: nil,
			},
			wantErr: true,
		},
		{
			name: "ValidateSignatureInfo - one or more participants provide empty signature",
			args: args{
				signature: nil,
				signatureInfo: &model.SignatureInfo{
					TransactionHash: txHash,
					Signatures: map[string][]byte{
						"a": nil,
					},
				},
				multiSignatureInfoAddresses: nil,
			},
			wantErr: true,
		},
		{
			name: "ValidateSignatureInfo - one or more participants is not participant in multisigInfo provided",
			args: args{
				signature: nil,
				signatureInfo: &model.SignatureInfo{
					TransactionHash: txHash,
					Signatures: map[string][]byte{
						"c": make([]byte, 68),
					},
				},
				multiSignatureInfoAddresses: map[string]bool{
					"a": true, "b": true,
				},
			},
			wantErr: true,
		},
		{
			name: "ValidateSignatureInfo - normal account participant provide wrong signature",
			args: args{
				signature: sig,
				signatureInfo: &model.SignatureInfo{
					TransactionHash: make([]byte, 32),
					Signatures: map[string][]byte{
						"a": make([]byte, 68),
					},
				},
				multiSignatureInfoAddresses: map[string]bool{
					"a": true, "b": true,
				},
			},
			wantErr: true,
		},
		{
			name: "ValidateSignatureInfo - normal account participant provide valid signature",
			args: args{
				signature: sig,
				signatureInfo: &model.SignatureInfo{
					TransactionHash: make([]byte, 32),
					Signatures: map[string][]byte{
						validAddress: validSignature,
					},
				},
				multiSignatureInfoAddresses: map[string]bool{
					validAddress: true, "b": true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtu := &MultisigTransactionUtil{}
			if err := mtu.ValidateSignatureInfo(tt.args.signature, tt.args.signatureInfo, tt.args.multiSignatureInfoAddresses); (err != nil) != tt.wantErr {
				t.Errorf("ValidateSignatureInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
