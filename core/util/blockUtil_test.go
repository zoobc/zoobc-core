// util package contain basic utilities commonly used across the core package
package util

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var mockBlockValidSignature = &model.Block{
	ID:        1,
	Height:    0,
	BlockSeed: []byte{1, 2, 3},
	BlockSignature: []byte{
		0, 0, 0, 0, 131, 245, 139, 208, 104, 104, 139, 152, 250, 103, 59, 182, 252, 147, 93, 12, 9, 4, 161, 17, 28, 70, 176, 87, 32,
		154, 148, 182, 65, 107, 4, 51, 129, 163, 16, 224, 214, 56, 72, 228, 204, 250, 61, 134, 253, 233, 131, 2, 41, 59, 205, 5, 22,
		118, 219, 204, 119, 111, 65, 173, 121, 219, 77, 14,
	},
	BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	CumulativeDifficulty: "0",
	PayloadHash:          []byte{},
	PayloadLength:        1,
	PreviousBlockHash: []byte{
		0, 0, 0, 0, 0, 0, 0, 0,
	},
	SmithScale:    0,
	Timestamp:     1000,
	TotalAmount:   1000,
	TotalCoinBase: 0,
	TotalFee:      1,
	Transactions:  nil,
	Version:       1,
}

var mockBlockZeroID = &model.Block{
	ID:        0,
	Height:    0,
	BlockSeed: []byte{1, 2, 3},
	BlockSignature: []byte{
		0, 0, 0, 0, 186, 94, 104, 188, 20, 86, 116, 3, 115, 173, 143, 37, 41, 248, 134, 75, 167, 131, 189, 249, 17, 150,
		240, 103, 44, 27, 239, 66, 176, 71, 71, 254, 42, 248, 246, 15, 220, 80, 209, 242, 146, 163, 95, 40, 62, 120, 28,
		226, 158, 234, 253, 46, 174, 223, 98, 53, 160, 186, 124, 154, 72, 61, 64, 1,
	},
	BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	CumulativeDifficulty: "0",
	PayloadHash:          []byte{},
	PayloadLength:        1,
	PreviousBlockHash: []byte{
		0, 0, 0, 0, 0, 0, 0, 0,
	},
	SmithScale:    0,
	Timestamp:     1000,
	TotalAmount:   1000,
	TotalCoinBase: 0,
	TotalFee:      1,
	Transactions:  nil,
	Version:       1,
}

var mockBlockPrevious = &model.Block{
	ID:        0,
	Height:    0,
	BlockSeed: []byte{1, 2, 3},
	BlockSignature: []byte{
		0, 0, 0, 0, 143, 169, 160, 83, 125, 53, 38, 84, 54, 90, 232, 190, 87, 217, 90, 227, 249, 241, 3, 200, 204, 47,
		221, 191, 151, 232, 111, 241, 248, 69, 82, 203, 75, 128, 164, 45, 162, 105, 166, 1, 81, 113, 65, 8, 71, 19, 26,
		20, 3, 49, 192, 244, 20, 155, 96, 242, 195, 19, 108, 251, 100, 55, 153, 12,
	},
	BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	CumulativeDifficulty: "0",
	PayloadHash:          []byte{},
	PayloadLength:        1,
	PreviousBlockHash:    []byte{},
	SmithScale:           0,
	Timestamp:            1000,
	TotalAmount:          1000,
	TotalCoinBase:        0,
	TotalFee:             1,
	Transactions:         nil,
	Version:              1,
}

var mockBlockInvalidPrevious = &model.Block{
	ID:        100,
	Height:    0,
	BlockSeed: []byte{1, 2, 3},
	BlockSignature: []byte{
		0, 0, 0, 0, 143, 169, 160, 83, 125, 53, 38, 84, 54, 90, 232, 190, 87, 217, 90, 227, 249, 241, 3, 200, 204, 47,
		221, 191, 151, 232, 111, 241, 248, 69, 82, 203, 75, 128, 164, 45, 162, 105, 166, 1, 81, 113, 65, 8, 71, 19, 26,
		20, 3, 49, 192, 244, 20, 155, 96, 242, 195, 19, 108, 251, 100, 55, 153, 12,
	},
	BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	CumulativeDifficulty: "0",
	PayloadHash:          []byte{},
	PayloadLength:        1,
	PreviousBlockHash:    []byte{},
	SmithScale:           0,
	Timestamp:            1000,
	TotalAmount:          1000,
	TotalCoinBase:        0,
	TotalFee:             1,
	Transactions:         nil,
	Version:              1,
}

var mockBlockInvalidSignature = &model.Block{
	ID:        1,
	Height:    0,
	BlockSeed: []byte{1, 2, 3},
	BlockSignature: []byte{
		186, 94, 104, 188, 20, 86, 116, 3, 115, 173, 143, 37, 41, 248, 134, 75, 167, 131, 189, 249, 17, 150,
		240, 103, 44, 27, 239, 66, 176, 71, 71, 254, 42, 248, 246, 15, 220, 80, 209, 242, 146, 163, 95, 40, 62, 120, 28,
		226, 158, 234, 253, 46, 174, 223, 98, 53, 160, 186, 124, 154, 72, 61, 64, 1,
	},
	BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	CumulativeDifficulty: "0",
	PayloadHash:          []byte{},
	PayloadLength:        1,
	PreviousBlockHash: []byte{
		0, 0, 0, 0, 0, 0, 0, 0,
	},
	SmithScale:    0,
	Timestamp:     1000,
	TotalAmount:   1000,
	TotalCoinBase: 0,
	TotalFee:      1,
	Transactions:  nil,
	Version:       1,
}

func TestGetBlockSeed(t *testing.T) {
	resultOne, _ := new(big.Int).SetString("17294717645224993457", 10)
	resulTwo, _ := new(big.Int).SetString("12040414258978844097", 10)
	type args struct {
		publicKey []byte
		block     *model.Block
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
				block: &model.Block{
					BlockSeed: []byte{1, 3, 4, 5},
				},
			},
			want:    resultOne,
			wantErr: false,
		},
		{
			name: "GetBlockSeed:two",
			args: args{
				publicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139,
					255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				block: &model.Block{
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
		block   *model.Block
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
				block: &model.Block{
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
		previousBlock     *model.Block
		block             *model.Block
		smithingDelayTime int64
	}
	tests := []struct {
		name string
		args args
		want *model.Block
	}{
		{
			name: "CalculateSmithScale",
			args: args{
				previousBlock: &model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{},
					BlockSeed:            []byte{},
					BlocksmithPublicKey:  []byte{},
					Timestamp:            15875392,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "100000",
					SmithScale:           108080,
				},
				block: &model.Block{
					Version:             1,
					PreviousBlockHash:   []byte{},
					BlockSeed:           []byte{},
					BlocksmithPublicKey: []byte{},
					Timestamp:           15875392,
					TotalAmount:         0,
					TotalFee:            0,
					TotalCoinBase:       0,
					Transactions:        []*model.Transaction{},
					PayloadHash:         []byte{},
				},
				smithingDelayTime: 10,
			},
			want: &model.Block{
				Version:              1,
				PreviousBlockHash:    []byte{},
				BlockSeed:            []byte{},
				BlocksmithPublicKey:  []byte{},
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
					BlocksmithPublicKey:  []byte{},
					Timestamp:            15875392,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "341353517378119",
					BlockSignature:       []byte{},
					SmithScale:           54040,
				},
			},
			want: 2302495433703223211,
		},
		{
			name: "GetBlockID:two",
			args: args{
				block: &model.Block{
					Version:              1,
					PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey:  []byte{},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					BlockSignature:       []byte{},
					SmithScale:           48985,
				},
			},
			want: -3939633329194296199,
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

func TestIsBlockIDExist(t *testing.T) { //todo:update test after applying signature related functionalities
	type args struct {
		blockIds        []int64
		expectedBlockID int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "wantSuccess:BlockIDFound",
			args: args{
				blockIds:        []int64{1, 2, 3, 4},
				expectedBlockID: int64(1),
			},
			want: true,
		},
		{
			name: "wantSuccess:InvalidTimestamp",
			args: args{
				blockIds:        []int64{1, 2, 3, 4},
				expectedBlockID: int64(5),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlockIDExist(tt.args.blockIds, tt.args.expectedBlockID); got != tt.want {
				t.Errorf("TestIsBlockIDExist() got = %v, want %v", got, tt.want)
			}
		})
	}
}
