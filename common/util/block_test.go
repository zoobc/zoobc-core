package util

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetBlockByte(t *testing.T) {
	type args struct {
		block *model.Block
		sign  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetBlockByte:one",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					SmithScale:           48985,
				},
				sign: false,
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21,
				99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 2, 65, 76, 32, 76,
				12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
			wantErr: false,
		},
		{
			name: "GetBlockByte:withSignature",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					BlockSignature:       []byte{1, 3, 4, 54, 65, 76, 3, 3, 54, 12, 5, 64, 23, 12, 21},
					SmithScale:           48985,
				},
				sign: true,
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199,
				145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 2, 65,
				76, 32, 76, 12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22, 1, 3, 4, 54, 65, 76, 3, 3, 54,
				12, 5, 64, 23, 12, 21},
			wantErr: false,
		},
		{
			name: "GetBlockByte:error-{sign true without signature}",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					SmithScale:           48985,
				},
				sign: true,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockByte(tt.args.block, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockByte() = %v, want %v", got, tt.want)
			}
		})
	}
}
