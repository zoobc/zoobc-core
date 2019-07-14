package transaction

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetTransactionType(t *testing.T) {
	type args struct {
		tx *model.Transaction
	}
	tests := []struct {
		name string
		args args
		want TypeAction
	}{
		{
			name: "wantSendMoney",
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountType:       0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountType:    0,
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{1, 0, 0, 0}),
				},
			},
			want: TypeAction(&SendMoney{
				Height:               0,
				SenderAccountType:    0,
				SenderAddress:        "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				RecipientAddress:     "",
				RecipientAccountType: 0,
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
			}),
		},
		{
			name: "wantEmpty",
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountType:       0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountType:    0,
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{0, 0, 0, 0}),
				},
			},
			want: TypeAction(&TXEmpty{}),
		},
		{
			name: "wantNil",
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountType:       0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountType:    0,
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{0, 1, 0, 0}),
				},
			},
		},
		{
			name: "wantNil",
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountType:       0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountType:    0,
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{1, 1, 0, 0}),
				},
			},
		},
		{
			name: "wantNil",
			args: args{
				tx: &model.Transaction{
					Height:                  0,
					SenderAccountType:       0,
					SenderAccountAddress:    "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					RecipientAccountType:    0,
					RecipientAccountAddress: "",
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10,
						},
					},
					TransactionType: binary.LittleEndian.Uint32([]byte{2, 1, 0, 0}),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTransactionType(tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactionType() = %v, want %v", got, tt.want)
			}
		})
	}
}
