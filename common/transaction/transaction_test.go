package transaction

import (
	"crypto/sha256"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
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
	// cache mock
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
	tests := []struct {
		name   string
		fields fields
		args   args
		want   TypeAction
	}{
		{
			name: "wantSendMoney",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10,
						},
					},
					TransactionBodyBytes: util.ConvertUint64ToBytes(10),
					TransactionType:      binary.LittleEndian.Uint32([]byte{1, 0, 0, 0}),
				},
			},
			want: &SendMoney{
				Height:           0,
				SenderAddress:    senderAddress1,
				RecipientAddress: nil,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				QueryExecutor: &query.Executor{},
				EscrowQuery:   query.NewEscrowTransactionQuery(),
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, constant.OneZBC/100,
				),
				NormalFee:            fee.NewConstantFeeModel(constant.OneZBC / 100),
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
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
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
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
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
				Height:                  0,
				SenderAddress:           senderAddress1,
				Body:                    nodeRegistrationBody,
				QueryExecutor:           &query.Executor{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:               &auth.NodeAuthValidation{},
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				EscrowQuery:             query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:    accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:                fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				Body:                  updateNodeRegistrationBody,
				Height:                0,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &query.Executor{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.NodeAuthValidation{},
				EscrowQuery:           query.NewEscrowTransactionQuery(),
				AccountBalanceHelper:  accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:                    fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				Body:                  removeNodeRegistrationBody,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &query.Executor{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountBalanceHelper:  accountBalanceHelper,
				NodeAddressInfoQuery:  query.NewNodeAddressInfoQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:                fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				Body:                  claimNodeRegistrationBody,
				Height:                0,
				SenderAddress:         senderAddress1,
				QueryExecutor:         &query.Executor{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.NodeAuthValidation{},
				AccountBalanceHelper:  accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:               fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				EscrowQuery:             query.NewEscrowTransactionQuery(),
				ActiveNodeRegistryCache: txActiveNodeRegistryCache,
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
					SenderAccountAddress:    senderAddress1,
					RecipientAccountAddress: nil,
					TransactionBody: &model.Transaction_SetupAccountDatasetTransactionBody{
						SetupAccountDatasetTransactionBody: mockSetupAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 0, 0, 0}),
					TransactionBodyBytes: mockBytesSetupAccountDataset,
				},
			},
			want: &SetupAccountDataset{
				Body:                 mockSetupAccountDatasetBody,
				Height:               5,
				SenderAddress:        senderAddress2,
				QueryExecutor:        &query.Executor{},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:   fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				EscrowQuery: query.NewEscrowTransactionQuery(),
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
				Body:                 mockRemoveAccountDatasetBody,
				Height:               5,
				SenderAddress:        senderAddress1,
				RecipientAddress:     recipientAddress1,
				QueryExecutor:        &query.Executor{},
				AccountDatasetQuery:  query.NewAccountDatasetsQuery(),
				AccountBalanceHelper: accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:   fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				EscrowQuery: query.NewEscrowTransactionQuery(),
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
				ID:               0,
				SenderAddress:    senderAddress1,
				Body:             approvalEscrowBody,
				Height:           5,
				QueryExecutor:    &query.Executor{},
				EscrowQuery:      query.NewEscrowTransactionQuery(),
				BlockQuery:       query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery: query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &query.Executor{},
				},
				AccountBalanceHelper: accountBalanceHelper,
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee: fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				Body:                       feeVoteCommitTransactionBody,
				Height:                     5,
				SenderAddress:              senderAddress1,
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
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:   fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
					TransactionType: binary.LittleEndian.Uint32([]byte{7, 1, 0, 0}),
					TransactionBodyBytes: (&FeeVoteRevealTransaction{
						Body: feeVoteRevealBody,
					}).GetBodyBytes(),
				},
			},
			want: &FeeVoteRevealTransaction{
				Body:                   feeVoteRevealBody,
				Height:                 5,
				SenderAddress:          senderAddress1,
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
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:   fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
				ID:                            0,
				SenderAddress:                 mockTxSenderAccountAddress,
				RecipientAddress:              mockTxRecipientAccountAddress,
				Body:                          liquidPaymentBody,
				Height:                        5,
				QueryExecutor:                 &query.Executor{},
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:   fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
				EscrowQuery: query.NewEscrowTransactionQuery(),
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
				ID:                            0,
				SenderAddress:                 mockTxSenderAccountAddress,
				RecipientAddress:              mockTxRecipientAccountAddress,
				Body:                          liquidPaymentStopBody,
				Height:                        5,
				QueryExecutor:                 &query.Executor{},
				AccountBalanceHelper:          accountBalanceHelper,
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &query.Executor{},
				},
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, fee.SendMoneyFeeConstant,
				),
				NormalFee:   fee.NewConstantFeeModel(fee.SendMoneyFeeConstant),
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
