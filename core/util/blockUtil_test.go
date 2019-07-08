// util package contain basic utilities commonly used across the core package
package util

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetBlockSeed(t *testing.T) {
	resultOne, _ := new(big.Int).SetString("17294717645224993457", 10)
	resulTwo, _ := new(big.Int).SetString("12040414258978844097", 10)
	type args struct {
		publicKey []byte
		block     model.Block
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name: "GetBlockSeed:one",
			args: args{
				publicKey: []byte{1, 2, 3, 4},
				block: model.Block{
					BlockSeed: []byte{1, 3, 4, 5},
				},
			},
			want:    resultOne,
			wantErr: false,
		},
		{
			name: "GetBlockSeed:two",
			args: args{
				publicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				block: model.Block{
					BlockSeed: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229},
				},
			},
			want:    resulTwo,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockSeed(tt.args.publicKey, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSmithTime(t *testing.T) {
	type args struct {
		balance *big.Int
		seed    *big.Int
		block   model.Block
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "GetSmithTime:0",
			args: args{
				balance: big.NewInt(0),
				seed:    big.NewInt(0),
			},
			want: 0,
		},
		{
			name: "GetSmithTime:!0",
			args: args{
				balance: big.NewInt(10000),
				seed:    big.NewInt(120000000),
				block: model.Block{
					SmithScale: 100,
					Timestamp:  120000,
				},
			},
			want: 120120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSmithTime(tt.args.balance, tt.args.seed, tt.args.block); got != tt.want {
				t.Errorf("GetSmithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateSmithScale(t *testing.T) {
	type args struct {
		previousBlock     model.Block
		block             model.Block
		smithingDelayTime int64
	}
	tests := []struct {
		name string
		args args
		want model.Block
	}{
		{
			name: "CalculateSmithScale",
			args: args{
				previousBlock: model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         []byte{},
					Timestamp:            15875392,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "100000",
					SmithScale:           108080,
				},
				block: model.Block{
					Version:           1,
					PreviousBlockHash: []byte{},
					BlockSeed:         []byte{},
					BlocksmithID:      []byte{},
					Timestamp:         15875392,
					TotalAmount:       0,
					TotalFee:          0,
					TotalCoinBase:     0,
					Transactions:      []*model.Transaction{},
					PayloadHash:       []byte{},
				},
				smithingDelayTime: 10,
			},
			want: model.Block{
				Version:              1,
				PreviousBlockHash:    []byte{},
				BlockSeed:            []byte{},
				BlocksmithID:         []byte{},
				Timestamp:            15875392,
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Transactions:         []*model.Transaction{},
				PayloadHash:          []byte{},
				CumulativeDifficulty: "341353517378119",
				SmithScale:           54040,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateSmithScale(tt.args.previousBlock, tt.args.block, tt.args.smithingDelayTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateSmithScale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBlockID(t *testing.T) {
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "GetBlockID:one",
			args: args{
				block: &model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithID:         []byte{},
					Timestamp:            15875392,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "341353517378119",
					SmithScale:           54040,
				},
			},
			want: -3024391177923659831,
		},
		{
			name: "GetBlockID:two",
			args: args{
				block: &model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithID:         []byte{12, 43, 65, 32, 56},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					SmithScale:           48985,
				},
			},
			want: 3300349166301930278,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBlockID(tt.args.block); got != tt.want {
				t.Errorf("GetBlockID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBlockByte(t *testing.T) {
	type args struct {
		block model.Block
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "GetBlockByte:one",
			args: args{
				block: model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithID:         []byte{12, 43, 65, 32, 56},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					SmithScale:           48985,
				},
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 43, 65, 32, 56, 2, 65, 76, 32, 76, 12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
		},
		{
			name: "GetBlockByte:withSignature",
			args: args{
				block: model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithID:         []byte{12, 43, 65, 32, 56},
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
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 43, 65, 32, 56, 2, 65, 76, 32, 76, 12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22, 1, 3, 4, 54, 65, 76, 3, 3, 54, 12, 5, 64, 23, 12, 21},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBlockByte(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateBlock(t *testing.T) { //todo:update test after applying signature related functionalities
	type args struct {
		block             *model.Block
		previousLastBlock model.Block
		curTime           int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateBlock(tt.args.block, tt.args.previousLastBlock, tt.args.curTime); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
