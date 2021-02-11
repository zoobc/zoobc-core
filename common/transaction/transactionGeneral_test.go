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
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
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
			want: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149,
				127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 0, 0, 0, 0, 229, 176, 168, 71,
				174, 217, 223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232,
				110, 14, 64, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 2, 0, 0, 0, 0, 0, 0, 0,
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
			want: []byte{2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149,
				127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 2, 0, 0, 0, 64, 66, 15, 0, 0,
				0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 2, 0, 0, 0, 0, 0, 0, 0,
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
				2, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214,
				82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 0, 0, 0, 0, 229, 176, 168, 71, 174, 217,
				223, 62, 98, 47, 207, 16, 210, 190, 79, 28, 126, 202, 25, 79, 137, 40, 243, 132, 77, 206, 170, 27, 124, 232, 110, 14, 64,
				66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 0, 2, 178, 0, 53, 239, 224, 110, 3, 190, 249, 254, 250,
				58, 2, 83, 75, 213, 137, 66, 236, 188, 43, 59, 241, 146, 243, 147, 58, 161, 35, 229, 54, 24, 0, 0, 0, 0, 0, 0, 0, 100, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
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
			want: []byte{4, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149,
				127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 2, 0, 0, 0, 1, 0, 0, 0, 0, 0,
				0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
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
			want: []byte{4, 0, 0, 0, 1, 32, 10, 133, 222, 107, 1, 0, 0, 0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149,
				127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169, 2, 0, 0, 0, 1, 0, 0, 0, 0, 0,
				0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
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
				fmt.Println(byteStrArr)
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
		ID:                      9040547499122759451,
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
			4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174, 239, 46, 190, 78,
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
		-8289164386094074251,
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
			ApproverAddress: mockTxSenderAccountAddress,
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
				ID:                      499264076282620792,
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
					ApproverAddress: mockTxSenderAccountAddress,
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
	signatureTXValidateNoRecipient, _ := (&crypto.Signature{}).Sign(txBytesHash[:], model.AccountType_ZbcAccountType,
		senderAddress1PassPhrase)
	txValidateNoRecipient.Signature = signatureTXValidateNoRecipient

	txValidateMustEscrow := GetFixturesForTransaction(
		1562893303,
		senderAddress1,
		recipientAddress1,
		false,
	)
	txBytesMustEscrow, _ := transactionUtil.GetTransactionBytes(txValidateMustEscrow, false)
	txBytesMustEscrowHash := sha3.Sum256(txBytesMustEscrow)
	signatureTXValidateMustEscrow, _ := (&crypto.Signature{}).Sign(txBytesMustEscrowHash[:], model.AccountType_ZbcAccountType,
		senderAddress1PassPhrase)
	txValidateMustEscrow.Signature = signatureTXValidateMustEscrow

	txValidateEscrow := GetFixturesForTransaction(
		1562893303,
		senderAddress1,
		recipientAddress1,
		true,
	)
	txValidateEscrow.Escrow.ApproverAddress = recipientAddress1
	txBytesEscrow, _ := transactionUtil.GetTransactionBytes(txValidateEscrow, false)
	txBytesEscrowHash := sha3.Sum256(txBytesEscrow)
	signatureTXValidateEscrow, _ := (&crypto.Signature{}).Sign(txBytesEscrowHash[:], model.AccountType_ZbcAccountType,
		senderAddress1PassPhrase)
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
			want: []byte{0, 0, 0, 0, 156, 245, 22, 64, 141, 106, 136, 228, 125, 30, 62, 62, 38, 92, 203, 116, 9, 51, 188, 100, 158, 147, 219, 171, 75,
				7, 219, 56, 28, 223, 180, 47},
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
	_, _, _, _, validAddress, _ := sig.GenerateAccountFromSeed(&accounttype.ZbcAccountType{}, "a")
	validAddressHex := hex.EncodeToString(validAddress)
	validSignature, _ := sig.Sign(txHash, model.AccountType_ZbcAccountType, "a")
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
						"00000000db112e8c4cd6ae57bde47eeb563e58d813f6f1e4c574c5c9253be91643ce6ed5": make([]byte, 68),
					},
				},
				multiSignatureInfoAddresses: map[string]bool{
					"0000000004264418e6f758dc777c33957fd652e048ef388bff51e5b84d505027fead1ca9": true,
					"000000004dfa35867733ca4ed2c68acc5a41a65996b9b06721c54619afc6b53f314ebf20": true,
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
						"0000000004264418e6f758dc777c33957fd652e048ef388bff51e5b84d505027fead1ca9": make([]byte, 68),
					},
				},
				multiSignatureInfoAddresses: map[string]bool{
					"0000000004264418e6f758dc777c33957fd652e048ef388bff51e5b84d505027fead1ca9": true,
					"000000004dfa35867733ca4ed2c68acc5a41a65996b9b06721c54619afc6b53f314ebf20": true,
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
						validAddressHex: validSignature,
					},
				},
				multiSignatureInfoAddresses: map[string]bool{
					validAddressHex: true,
					"000000004dfa35867733ca4ed2c68acc5a41a65996b9b06721c54619afc6b53f314ebf20": true,
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
