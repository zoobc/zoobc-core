package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

func TestGetGenesisTransactions(t *testing.T) {
	type args struct {
		chainType contract.ChainType
	}
	tests := []struct {
		name string
		args args
		want []*model.Transaction
	}{
		{
			name: "wantGenesisTX",
			args: args{
				chainType: &chaintype.MainChain{},
			},
			want: []*model.Transaction{
				{
					Version:                 1,
					TransactionType:         util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
					Height:                  0,
					Timestamp:               1562806389280,
					SenderAccountType:       0,
					SenderAccountAddress:    genesisSender,
					RecipientAccountType:    0,
					RecipientAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					Fee:                     0,
					TransactionBodyLength:   8,
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10000000,
						},
					},
					TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(10000000)),
					Signature:            genesisSignature,
				},
			},
		},
		{
			name: "wantNilTX",
			args: args{
				chainType: nil,
			},
			want: []*model.Transaction{
				{
					Version:                 1,
					TransactionType:         util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
					Height:                  0,
					Timestamp:               1562806389280,
					SenderAccountType:       0,
					SenderAccountAddress:    genesisSender,
					RecipientAccountType:    0,
					RecipientAccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					Fee:                     0,
					TransactionBodyLength:   8,
					TransactionBody: &model.Transaction_SendMoneyTransactionBody{
						SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
							Amount: 10000000,
						},
					},
					TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(10000000)),
					Signature:            genesisSignature,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGenesisTransactions(tt.args.chainType); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetGenesisTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
