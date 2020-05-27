package transaction

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/fee"

	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

func TestTypeSwitcher_GetTransactionType(t *testing.T) {
	_, _, nodeRegistrationBody, nodeRegistrationBodyBytes := GetFixturesForNoderegistration(query.NewNodeRegistrationQuery())
	_, _, updateNodeRegistrationBody, updateNodeRegistrationBodyBytes := GetFixturesForUpdateNoderegistration(query.NewNodeRegistrationQuery())
	removeNodeRegistrationBody, removeNodeRegistrationBodyBytes := GetFixturesForRemoveNoderegistration()
	_, claimNodeRegistrationBody, claimNodeRegistrationBodyBytes := GetFixturesForClaimNoderegistration()

	mockSetupAccountDatasetBody, mockBytesSetupAccountDataset := GetFixturesForSetupAccountDataset()
	mockRemoveAccountDatasetBody, mockBytesRemoveAccountDataset := GetFixturesForRemoveAccountDataset()

	approvalEscrowBody, approvalEscrowBytes := GetFixturesForApprovalEscrowTransaction()

	liquidPaymentBody, liquidPaymentBytes := GetFixturesForLiquidPaymentTransaction()
	liquidPaymentStopBody, liquidPaymentStopBytes := GetFixturesForLiquidPaymentStopTransaction()

	type fields struct {
		Executor query.ExecutorInterface
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
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountAddress: "",
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
				SenderAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress: "",
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				QueryExecutor:       &query.Executor{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				EscrowQuery:         query.NewEscrowTransactionQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				EscrowFee: fee.NewBlockLifeTimeFeeModel(
					10, constant.OneZBC/100,
				),
				NormalFee:           fee.NewConstantFeeModel(constant.OneZBC / 100),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
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
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountAddress: "",
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
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountAddress: "",
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
			name: "wantNil",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountAddress: "",
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
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
						NodeRegistrationTransactionBody: nodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 0, 0, 0}),
					TransactionBodyBytes: nodeRegistrationBodyBytes,
				},
			},
			want: &NodeRegistration{
				Height:                  0,
				SenderAddress:           "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Body:                    nodeRegistrationBody,
				QueryExecutor:           &query.Executor{},
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:               &auth.ProofOfOwnershipValidation{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				AccountLedgerQuery:      query.NewAccountLedgerQuery(),
			},
		},
		{
			name: "wantUpdateNodeRegistration",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountAddress: "",
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
				SenderAddress:         "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				QueryExecutor:         &query.Executor{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.ProofOfOwnershipValidation{},
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
			},
		},
		{
			name: "wantRemoveNodeRegistration",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_RemoveNodeRegistrationTransactionBody{
						RemoveNodeRegistrationTransactionBody: removeNodeRegistrationBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{2, 2, 0, 0}),
					TransactionBodyBytes: removeNodeRegistrationBodyBytes,
				},
			},
			want: &RemoveNodeRegistration{
				Body:                  removeNodeRegistrationBody,
				Height:                0,
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				QueryExecutor:         &query.Executor{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
			},
		},
		{
			name: "wantClaimNodeRegistration",
			fields: fields{
				Executor: &query.Executor{},
			},
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountAddress:    "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
					RecipientAccountAddress: "",
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
				SenderAddress:         "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
				QueryExecutor:         &query.Executor{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				AuthPoown:             &auth.ProofOfOwnershipValidation{},
				AccountLedgerQuery:    query.NewAccountLedgerQuery(),
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
					SenderAccountAddress:    mockSetupAccountDatasetBody.GetSetterAccountAddress(),
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_SetupAccountDatasetTransactionBody{
						SetupAccountDatasetTransactionBody: mockSetupAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 0, 0, 0}),
					TransactionBodyBytes: mockBytesSetupAccountDataset,
				},
			},
			want: &SetupAccountDataset{
				Body:                mockSetupAccountDatasetBody,
				Height:              5,
				SenderAddress:       mockSetupAccountDatasetBody.GetSetterAccountAddress(),
				QueryExecutor:       &query.Executor{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
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
					SenderAccountAddress:    mockRemoveAccountDatasetBody.GetSetterAccountAddress(),
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_RemoveAccountDatasetTransactionBody{
						RemoveAccountDatasetTransactionBody: mockRemoveAccountDatasetBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{3, 1, 0, 0}),
					TransactionBodyBytes: mockBytesRemoveAccountDataset,
				},
			},
			want: &RemoveAccountDataset{
				Body:                mockRemoveAccountDatasetBody,
				Height:              5,
				SenderAddress:       mockRemoveAccountDatasetBody.GetSetterAccountAddress(),
				QueryExecutor:       &query.Executor{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountDatasetQuery: query.NewAccountDatasetsQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
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
					SenderAccountAddress:    mockRemoveAccountDatasetBody.GetSetterAccountAddress(),
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_ApprovalEscrowTransactionBody{
						ApprovalEscrowTransactionBody: approvalEscrowBody,
					},
					TransactionType:      binary.LittleEndian.Uint32([]byte{4, 0, 0, 0}),
					TransactionBodyBytes: approvalEscrowBytes,
				},
			},
			want: &ApprovalEscrowTransaction{
				ID:                  0,
				SenderAddress:       mockRemoveAccountDatasetBody.GetSetterAccountAddress(),
				Body:                approvalEscrowBody,
				Height:              5,
				QueryExecutor:       &query.Executor{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
				EscrowQuery:         query.NewEscrowTransactionQuery(),
				BlockQuery:          query.NewBlockQuery(&chaintype.MainChain{}),
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &query.Executor{},
				},
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
				AccountBalanceQuery:           query.NewAccountBalanceQuery(),
				AccountLedgerQuery:            query.NewAccountLedgerQuery(),
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				NormalFee:                     fee.NewConstantFeeModel(constant.OneZBC / 100),
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
				AccountBalanceQuery:           query.NewAccountBalanceQuery(),
				AccountLedgerQuery:            query.NewAccountLedgerQuery(),
				LiquidPaymentTransactionQuery: query.NewLiquidPaymentTransactionQuery(),
				TransactionQuery:              query.NewTransactionQuery(&chaintype.MainChain{}),
				NormalFee:                     fee.NewConstantFeeModel(constant.OneZBC / 100),
				TypeActionSwitcher: &TypeSwitcher{
					Executor: &query.Executor{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TypeSwitcher{
				Executor: tt.fields.Executor,
			}
			if got, _ := ts.GetTransactionType(tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypeSwitcher.GetTransactionType() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
