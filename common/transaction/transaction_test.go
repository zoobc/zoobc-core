package transaction

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestTypeSwitcher_GetTransactionType(t *testing.T) {
	_, _, nodeRegistrationBody, nodeRegistrationBodyBytes := GetFixturesForNoderegistration()
	_, _, updateNodeRegistrationBody, updateNodeRegistrationBodyBytes := GetFixturesForUpdateNoderegistration()
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
				Height:                0,
				SenderAddress:         "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				Body:                  nodeRegistrationBody,
				QueryExecutor:         &query.Executor{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TypeSwitcher{
				Executor: tt.fields.Executor,
			}
			if got := ts.GetTransactionType(tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypeSwitcher.GetTransactionType() = %v, want %v", got, tt.want)
			}
		})
	}
}
