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
	"crypto/sha256"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	mockMainBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
)

func TestTypeSwitcher_GetTransactionType(t *testing.T) {
	_, _, nodeRegistrationBody, nodeRegistrationBodyBytes := GetFixturesForNoderegistration(query.NewNodeRegistrationQuery())
	_, _, updateNodeRegistrationBody, updateNodeRegistrationBodyBytes := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	removeNodeRegistrationBody, removeNodeRegistrationBodyBytes := GetFixturesForRemoveNoderegistration()
	_, claimNodeRegistrationBody, claimNodeRegistrationBodyBytes := GetFixturesForClaimNoderegistration()

	mockSetupAccountDatasetBody, mockBytesSetupAccountDataset := GetFixturesForSetupAccountDataset()
	mockRemoveAccountDatasetBody, mockBytesRemoveAccountDataset := GetFixturesForRemoveAccountDataset()

	approvalEscrowBody, approvalEscrowBytes := GetFixturesForApprovalEscrowTransaction()
	feeVoteCommitTransactionBody, feeVoteCommitTransactionBodyBytes := GetFixtureForFeeVoteCommitTransaction(&model.FeeVoteInfo{
		RecentBlockHash:   []byte{},
		RecentBlockHeight: 100,
		FeeVote:           10,
	}, "ZOOBC")
	feeVoteRevealBody := GetFixtureForFeeVoteRevealTransaction(&model.FeeVoteInfo{
		RecentBlockHash:   sha256.New().Sum([]byte{}),
		RecentBlockHeight: 100,
		FeeVote:           10,
	}, "ZOOBC")
	liquidPaymentBody, liquidPaymentBytes := GetFixturesForLiquidPaymentTransaction()
	liquidPaymentStopBody, liquidPaymentStopBytes := GetFixturesForLiquidPaymentStopTransaction()
	accountBalanceHelper := NewAccountBalanceHelper(&query.Executor{}, query.NewAccountBalanceQuery(), query.NewAccountLedgerQuery())
	// // cache mock
	fixtureTransactionalCache := func(cache interface{}) storage.TransactionalCache {
		return cache.(storage.TransactionalCache)
	}
	txActiveNodeRegistryCache := fixtureTransactionalCache(&mockNodeRegistryCacheSuccess{})
	txPendingNodeRegistryCache := fixtureTransactionalCache(&mockNodeRegistryCacheSuccess{})
	type fields struct {
		Executor                   query.ExecutorInterface
		NodeAuthValidation         auth.NodeAuthValidationInterface
		MempoolCacheStorage        storage.CacheStorageInterface
		PendingNodeRegistryStorage storage.TransactionalCache
		ActiveNodeRegistryStorage  storage.TransactionalCache
		NodeAddressInfoStorage     storage.TransactionalCache
		FeeScaleService            fee.FeeScaleServiceInterface
	}
	type args struct {
		tx *model.Transaction
	}

	txBodyBytes1, _ := (&FeeVoteRevealTransaction{
		Body: feeVoteRevealBody,
	}).GetBodyBytes()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   TypeAction
	}{
		{
			name: "wantSendZBC",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_SendZBCTransactionBody{
						SendZBCTransactionBody: &model.SendZBCTransactionBody{
							Amount: 10,
						},
					},
					TransactionBodyBytes: util.ConvertUint64ToBytes(10),
					TransactionType:      binary.LittleEndian.Uint32([]byte{1, 0, 0, 0}),
				},
			},
			want: &SendZBC{
				TransactionObject: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBodyBytes:    util.ConvertUint64ToBytes(10),
					TransactionType:         binary.LittleEndian.Uint32([]byte{1, 0, 0, 0}),
					TransactionBody: &model.Transaction_SendZBCTransactionBody{
						SendZBCTransactionBody: &model.SendZBCTransactionBody{
							Amount: 10,
						},
					},
				},
				Body: &model.SendZBCTransactionBody{
					Amount: 10,
				},
				QueryExecutor:        &query.Executor{},
				EscrowQuery:          query.NewEscrowTransactionQuery(),
				BlockQuery:           query.NewBlockQuery(&chaintype.MainChain{}),
				AccountBalanceHelper: accountBalanceHelper,
			},
		},
		{
			name: "wantEmpty",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_SendZBCTransactionBody{
						SendZBCTransactionBody: &model.SendZBCTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{0, 0, 0, 0}),
				},
			},
			want: &TXEmpty{},
		},
		{
			name: "wantNil",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_SendZBCTransactionBody{
						SendZBCTransactionBody: &model.SendZBCTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
				},
			},
			want: nil,
		},
		{
			name: "wantNodeRegistration",
			fields: fields{
				Executor:                   &query.Executor{},
				NodeAuthValidation:         &auth.NodeAuthValidation{},
				PendingNodeRegistryStorage: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryStorage:  &mockNodeRegistryCacheSuccess{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
						NodeRegistrationTransactionBody: nodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 0, 0, 0}),
					TransactionBodyBytes: nodeRegistrationBodyBytes,
				},
			},
			want: &NodeRegistration{
				TransactionObject: &model.Transaction{
					Height:               0,
					SenderAccountAddress: senderAddress1,
					TransactionBodyBytes: nodeRegistrationBodyBytes,
					TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
						NodeRegistrationTransactionBody: nodeRegistrationBody,
					},
					TransactionType: 2,
				},
				Body:                     nodeRegistrationBody,
				QueryExecutor:            &query.Executor{},
				BlockQuery:               query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:                &auth.NodeAuthValidation{},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery:  query.NewParticipationScoreQuery(),
				EscrowQuery:              query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:     accountBalanceHelper,
				PendingNodeRegistryCache: txPendingNodeRegistryCache,
			},
		},
		{
			name: "wantUpdateNodeRegistration",
			fields: fields{
				Executor:                   &query.Executor{},
				NodeAuthValidation:         &auth.NodeAuthValidation{},
				PendingNodeRegistryStorage: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryStorage:  &mockNodeRegistryCacheSuccess{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_UpdateNodeRegistrationTransactionBody{
						UpdateNodeRegistrationTransactionBody: updateNodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 1, 0, 0}),
					TransactionBodyBytes: updateNodeRegistrationBodyBytes,
				},
			},
			want: &UpdateNodeRegistration{
				Body: updateNodeRegistrationBody,
				TransactionObject: &model.Transaction{
					Height:               0,
					SenderAccountAddress: senderAddress1,
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 1, 0, 0}),
					TransactionBodyBytes: updateNodeRegistrationBodyBytes,
					TransactionBody: &model.Transaction_UpdateNodeRegistrationTransactionBody{
						UpdateNodeRegistrationTransactionBody: updateNodeRegistrationBody,
					},
				},
				QueryExecutor:                &query.Executor{},
				NodeRegistrationQuery:        query.NewNodeRegistrationQuery(),
				BlockQuery:                   query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:                    &auth.NodeAuthValidation{},
				EscrowQuery:                  query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:         accountBalanceHelper,
				ActiveNodeRegistrationCache:  txActiveNodeRegistryCache,
				PendingNodeRegistrationCache: txPendingNodeRegistryCache,
			},
		},
		{
			name: "wantRemoveNodeRegistration",
			fields: fields{
				Executor:                   &query.Executor{},
				NodeAuthValidation:         &auth.NodeAuthValidation{},
				PendingNodeRegistryStorage: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryStorage:  &mockNodeRegistryCacheSuccess{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_RemoveNodeRegistrationTransactionBody{
						RemoveNodeRegistrationTransactionBody: removeNodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 2, 0, 0}),
					TransactionBodyBytes: removeNodeRegistrationBodyBytes,
				},
			},
			want: &RemoveNodeRegistration{
				Body: removeNodeRegistrationBody,
				TransactionObject: &model.Transaction{
					SenderAccountAddress: senderAddress1,
					TransactionBody: &model.Transaction_RemoveNodeRegistrationTransactionBody{
						RemoveNodeRegistrationTransactionBody: removeNodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 2, 0, 0}),
					TransactionBodyBytes: removeNodeRegistrationBodyBytes,
				},
				QueryExecutor:            &query.Executor{},
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				AccountBalanceHelper:     accountBalanceHelper,
				NodeAddressInfoQuery:     query.NewNodeAddressInfoQuery(),
				EscrowQuery:              query.NewEscrowTransactionQuery(),
				PendingNodeRegistryCache: txPendingNodeRegistryCache,
				ActiveNodeRegistryCache:  txActiveNodeRegistryCache,
			},
		},
		{
			name: "wantClaimNodeRegistration",
			fields: fields{
				Executor:                   &query.Executor{},
				NodeAuthValidation:         &auth.NodeAuthValidation{},
				PendingNodeRegistryStorage: &mockNodeRegistryCacheSuccess{},
				ActiveNodeRegistryStorage:  &mockNodeRegistryCacheSuccess{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_ClaimNodeRegistrationTransactionBody{
						ClaimNodeRegistrationTransactionBody: claimNodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 3, 0, 0}),
					TransactionBodyBytes: claimNodeRegistrationBodyBytes,
				},
			},
			want: &ClaimNodeRegistration{
				Body: claimNodeRegistrationBody,
				TransactionObject: &model.Transaction{
					Height:               0,
					SenderAccountAddress: senderAddress1,
					TransactionBody: &model.Transaction_ClaimNodeRegistrationTransactionBody{
						ClaimNodeRegistrationTransactionBody: claimNodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 3, 0, 0}),
					TransactionBodyBytes: claimNodeRegistrationBodyBytes,
				},
				QueryExecutor:           &query.Executor{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:               &auth.NodeAuthValidation{},
				AccountBalanceHelper:    accountBalanceHelper,
				EscrowQuery:             query.NewEscrowTransactionQuery(),
				ActiveNodeRegistryCache: txActiveNodeRegistryCache,
				NodeAddressInfoQuery:    query.NewNodeAddressInfoQuery(),
			},
		},
		{
			name: "wantSetupDataset",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    senderAddress2,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_SetupAccountDatasetTransactionBody{
						SetupAccountDatasetTransactionBody: mockSetupAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 0, 0, 0}),
					TransactionBodyBytes: mockBytesSetupAccountDataset,
				},
			},
			want: &SetupAccountDataset{
				Body: mockSetupAccountDatasetBody,
				TransactionObject: &model.Transaction{
					Height:               5,
					SenderAccountAddress: senderAddress2,
					TransactionBody: &model.Transaction_SetupAccountDatasetTransactionBody{
						SetupAccountDatasetTransactionBody: mockSetupAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 0, 0, 0}),
					TransactionBodyBytes: mockBytesSetupAccountDataset,
				},
				QueryExecutor:        &query.Executor{},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
			},
		},
		{
			name: "wantRemoveDataset",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_RemoveAccountDatasetTransactionBody{
						RemoveAccountDatasetTransactionBody: mockRemoveAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 1, 0, 0}),
					TransactionBodyBytes: mockBytesRemoveAccountDataset,
				},
			},
			want: &RemoveAccountDataset{
				Body: mockRemoveAccountDatasetBody,
				TransactionObject: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_RemoveAccountDatasetTransactionBody{
						RemoveAccountDatasetTransactionBody: mockRemoveAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 1, 0, 0}),
					TransactionBodyBytes: mockBytesRemoveAccountDataset,
				},
				QueryExecutor:        &query.Executor{},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowQuery:          query.NewEscrowTransactionQuery(),
			},
		},
		{
			name: "wantEscrowApproval",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_ApprovalEscrowTransactionBody{
						ApprovalEscrowTransactionBody: approvalEscrowBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{4, 0, 0, 0}),
					TransactionBodyBytes: approvalEscrowBytes,
				},
			},
			want: &ApprovalEscrowTransaction{
				TransactionObject: &model.Transaction{
					ID:                   0,
					SenderAccountAddress: senderAddress1,
					Height:               5,
					TransactionBody: &model.Transaction_ApprovalEscrowTransactionBody{
						ApprovalEscrowTransactionBody: approvalEscrowBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{4, 0, 0, 0}),
					TransactionBodyBytes: approvalEscrowBytes,
				},
				Body:             approvalEscrowBody,
				QueryExecutor:    &query.Executor{},
				EscrowQuery:      query.NewEscrowTransactionQuery(),
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &query.Executor{},
				},
				AccountBalanceHelper: accountBalanceHelper,
			},
		},
		{
			name: "wantFeeVoteCommitTransaction",
			fields: fields{
				Executor: &query.Executor{},
				FeeScaleService: fee.NewFeeScaleService(
					query.NewFeeScaleQuery(),
					&mockMainBlockStateStorageSuccess{},
					&query.Executor{},
				),
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_FeeVoteCommitTransactionBody{
						FeeVoteCommitTransactionBody: feeVoteCommitTransactionBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{7, 0, 0, 0}),
					TransactionBodyBytes: feeVoteCommitTransactionBodyBytes,
				},
			},
			want: &FeeVoteCommitTransaction{
				Body: feeVoteCommitTransactionBody,
				TransactionObject: &model.Transaction{
					Height:               5,
					SenderAccountAddress: senderAddress1,
					TransactionBody: &model.Transaction_FeeVoteCommitTransactionBody{
						FeeVoteCommitTransactionBody: feeVoteCommitTransactionBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{7, 0, 0, 0}),
					TransactionBodyBytes: feeVoteCommitTransactionBodyBytes,
				},
				QueryExecutor:              &query.Executor{},
				AccountBalanceHelper:       accountBalanceHelper,
				BlockQuery:                 query.NewBlockQuery(&chaintype.MainChain{}),
				NodeRegistrationQuery:      query.NewNodeRegistrationQuery(),
				FeeVoteCommitmentVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeScaleService: fee.NewFeeScaleService(
					query.NewFeeScaleQuery(),
					&mockMainBlockStateStorageSuccess{},
					&query.Executor{},
				),
				EscrowQuery: query.NewEscrowTransactionQuery(),
			},
		},
		{
			name: "wantFeeVoteRevealTransaction",
			fields: fields{
				Executor: &query.Executor{},
				FeeScaleService: fee.NewFeeScaleService(
					query.NewFeeScaleQuery(),
					&mockMainBlockStateStorageSuccess{},
					&query.Executor{},
				),
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_FeeVoteRevealTransactionBody{
						FeeVoteRevealTransactionBody: feeVoteRevealBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{7, 1, 0, 0}),
					TransactionBodyBytes: txBodyBytes1,
				},
			},
			want: &FeeVoteRevealTransaction{
				Body: feeVoteRevealBody,
				TransactionObject: &model.Transaction{
					Height:               5,
					SenderAccountAddress: senderAddress1,
					TransactionBody: &model.Transaction_FeeVoteRevealTransactionBody{
						FeeVoteRevealTransactionBody: feeVoteRevealBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{7, 1, 0, 0}),
					TransactionBodyBytes: txBodyBytes1,
				},
				QueryExecutor:          &query.Executor{},
				AccountBalanceHelper:   accountBalanceHelper,
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				BlockQuery:             query.NewBlockQuery(&chaintype.MainChain{}),
				SignatureInterface:     crypto.NewSignature(),
				FeeVoteCommitVoteQuery: query.NewFeeVoteCommitmentVoteQuery(),
				FeeVoteRevealVoteQuery: query.NewFeeVoteRevealVoteQuery(),
				FeeScaleService: fee.NewFeeScaleService(
					query.NewFeeScaleQuery(),
					&mockMainBlockStateStorageSuccess{},
					&query.Executor{},
				),
				EscrowQuery: query.NewEscrowTransactionQuery(),
			},
		},
		{
			name: "wantLiquidPayment",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					TransactionBody:         liquidPaymentBody,
					TransactionType:         binary.LittleEndian.Uint32([]byte{6, 0, 0, 0}),
					TransactionBodyBytes:    liquidPaymentBytes,
				},
			},
			want: &LiquidPaymentTransaction{
				TransactionObject: &model.Transaction{
					ID:                      0,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					Height:                  5,
					TransactionBody:         liquidPaymentBody,
					TransactionType:         binary.LittleEndian.Uint32([]byte{6, 0, 0, 0}),
					TransactionBodyBytes:    liquidPaymentBytes,
				},
				Body:                          liquidPaymentBody,
				QueryExecutor:                 &query.Executor{},
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				EscrowQuery:                   query.NewEscrowTransactionQuery(),
			},
		},
		{
			name: "wantLiquidPaymentStop",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  5,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					TransactionBody:         liquidPaymentStopBody,
					TransactionType:         binary.LittleEndian.Uint32([]byte{6, 1, 0, 0}),
					TransactionBodyBytes:    liquidPaymentStopBytes,
				},
			},
			want: &LiquidPaymentStopTransaction{
				TransactionObject: &model.Transaction{
					ID:                      0,
					SenderAccountAddress:    mockTxSenderAccountAddress,
					RecipientAccountAddress: mockTxRecipientAccountAddress,
					Height:                  5,
					TransactionBody:         liquidPaymentStopBody,
					TransactionType:         binary.LittleEndian.Uint32([]byte{6, 1, 0, 0}),
					TransactionBodyBytes:    liquidPaymentStopBytes,
				},
				Body:                          liquidPaymentStopBody,
				QueryExecutor:                 &query.Executor{},
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &query.Executor{},
				},
				EscrowQuery: query.NewEscrowTransactionQuery(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TypeSwitcher{
				Executor:                   tt.fields.Executor,
				MempoolCacheStorage:        tt.fields.MempoolCacheStorage,
				NodeAuthValidation:         tt.fields.NodeAuthValidation,
				PendingNodeRegistryStorage: tt.fields.PendingNodeRegistryStorage,
				ActiveNodeRegistryStorage:  tt.fields.ActiveNodeRegistryStorage,
				NodeAddressInfoStorage:     tt.fields.NodeAddressInfoStorage,
				FeeScaleService:            tt.fields.FeeScaleService,
			}
			if got, _ := ts.GetTransactionType(tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypeSwitcher.GetTransactionType() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
