// util package contain basic utilities commonly used across the core package
package util

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestGetBlockSeed(t *testing.T) {
	resultOne, _ := new(big.Int).SetString("7942030951238827391", 10)
	resulTwo, _ := new(big.Int).SetString("5467201322837561108", 10)
	type args struct {
		publicKey    []byte
		nodeID       int64
		block        *model.Block
		secretPhrase string
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
				nodeID:    10,
				block: &model.Block{
					BlockSeed: []byte{1, 3, 4, 5},
				},
				secretPhrase: "randomsecretphrase",
			},
			want:    resultOne,
			wantErr: false,
		},
		{
			name: "GetBlockSeed:two",
			args: args{
				nodeID: 20,
				publicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139,
					255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				block: &model.Block{
					BlockSeed: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229},
				},
				secretPhrase: "randomsecretphrase",
			},
			want:    resulTwo,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockSeed(tt.args.nodeID, tt.args.block)
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
				block: &model.Block{
					SmithScale: 1,
					Timestamp:  0,
				},
			},
			want: 0,
		},
		{
			name: "GetSmithTime:!0",
			args: args{
				seed: big.NewInt(120000000),
				block: &model.Block{
					SmithScale: 100,
					Timestamp:  120000,
				},
			},
			want: 120000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSmithTime(tt.args.seed, tt.args.block); got != tt.want {
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
		blockQuery        query.BlockQueryInterface
		executor          query.ExecutorInterface
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
			if got, _ := CalculateSmithScale(
				tt.args.previousBlock,
				tt.args.block,
				tt.args.smithingDelayTime,
				tt.args.blockQuery,
				tt.args.executor,
			); !reflect.DeepEqual(got, tt.want) {
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
			want: 4891391764897612667,
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
			want: 5677934310196121651,
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

func TestIsGenesis(t *testing.T) {
	type args struct {
		previousBlockID int64
		block           *model.Block
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not genesis",
			args: args{
				previousBlockID: 1,
				block: &model.Block{
					ID:                   2,
					PreviousBlockHash:    nil,
					Height:               2,
					Timestamp:            0,
					BlockSeed:            nil,
					BlockSignature:       nil,
					CumulativeDifficulty: "",
					SmithScale:           0,
				},
			},
			want: false,
		},
		{
			name: "genesis",
			args: args{
				previousBlockID: -1,
				block: &model.Block{
					ID:                   1,
					PreviousBlockHash:    nil,
					Height:               2,
					Timestamp:            0,
					BlockSeed:            nil,
					BlockSignature:       nil,
					CumulativeDifficulty: "11111",
					SmithScale:           11110,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGenesis(tt.args.previousBlockID, tt.args.block); got != tt.want {
				t.Errorf("IsGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateSmithOrder(t *testing.T) {
	mockBlock := &model.Block{
		BlockSeed:            new(big.Int).SetInt64(1000).Bytes(),
		BlockSignature:       nil,
		CumulativeDifficulty: "",
		SmithScale:           1000,
		BlocksmithPublicKey:  nil,
	}
	type args struct {
		nodeID int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "CalculateSmithOrder:success",
			args: args{
				nodeID: int64(1000),
			},
			want: "17465045",
		},
		{
			name: "CalculateSmithOrder:success2",
			args: args{
				nodeID: int64(10),
			},
			want: "5535207",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CalculateSmithOrder(tt.args.nodeID, mockBlock); !reflect.DeepEqual(fmt.Sprintf("%d", got), tt.want) {
				t.Errorf("CalculateSmithOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCalculateSmithOrderDistribution simulation of pseudo-random smith distribution
func TestCalculateSmithOrderDistribution(t *testing.T) {
	type args struct {
		initialDistribution map[int64]int
		score               *big.Int
		blockSeeds          []*model.Block
	}
	tests := []struct {
		name string
		args args
		want map[int64]int
	}{
		{
			name: "CalculateSmithOrderDistribution:success",
			args: args{
				initialDistribution: map[int64]int{
					7357219233906154824:  0,
					9145301605531286876:  0,
					-8566484392056561567: 0,
					7937868735467829764:  0,
					-243345789637905342:  0,
				},
				score:      new(big.Int).SetInt64(1000),
				blockSeeds: getBlockSeedsFromFile("blockSeeds1000.csv"),
			},
			want: map[int64]int{
				-8566484392056561567: 192,
				-243345789637905342:  190,
				7357219233906154824:  213,
				7937868735467829764:  195,
				9145301605531286876:  210,
			},
		},
		{
			name: "CalculateSmithOrderDistribution:success2",
			args: args{
				initialDistribution: map[int64]int{
					7357219233906154824:  0,
					9145301605531286876:  0,
					-8566484392056561567: 0,
					7937868735467829764:  0,
					-243345789637905342:  0,
				},
				score:      new(big.Int).SetInt64(1000),
				blockSeeds: getBlockSeedsFromFile("blockSeeds10000.csv"),
			},
			want: map[int64]int{
				-8566484392056561567: 1963,
				-243345789637905342:  2024,
				7357219233906154824:  2049,
				7937868735467829764:  2003,
				9145301605531286876:  1961,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[int64]int
			if got = testCalculateSmithDistribution(tt.args.initialDistribution, tt.args.blockSeeds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateSmithOrderDistribution() = %v, want %v", got, tt.want)
			}
			// verify that we get the same distribution if we re-calculate using same input
			if gotVerify := testCalculateSmithDistribution(tt.args.initialDistribution, tt.args.blockSeeds); !reflect.DeepEqual(gotVerify, got) {
				t.Errorf("CalculateSmithOrderDistribution():veriry = %v, want %v", gotVerify, got)
			}
		})
	}
}

func getBlockSeedsFromFile(fileName string) (blocks []*model.Block) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", fileName))
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(strings.NewReader(string(b)))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	for _, rec := range records {
		seed, _ := new(big.Int).SetString(rec[0], 10)
		blocks = append(blocks, &model.Block{
			BlockSeed:  seed.Bytes(),
			SmithScale: 1000,
		})
	}
	return blocks
}

func testCalculateSmithDistribution(distribution map[int64]int, blocks []*model.Block) map[int64]int {
	for i := 0; i < len(blocks); i++ {
		var (
			maxSmithRndNum *big.Int
			selectedID     int64
		)
		j := 0
		for nodeID := range distribution {
			smithRndNum, _ := CalculateSmithOrder(nodeID, blocks[i])
			if j == 0 {
				maxSmithRndNum = smithRndNum
				selectedID = nodeID
				j++
				continue
			} else if smithRndNum.Cmp(maxSmithRndNum) > 0 {
				maxSmithRndNum = smithRndNum
				selectedID = nodeID
			}
		}
		distribution[selectedID]++
	}
	return distribution
}

func TestCalculateNodeOrder(t *testing.T) {
	type args struct {
		score     *big.Int
		blockSeed *big.Int
		nodeID    int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "CalculateNodeOrder:success",
			args: args{
				score:     new(big.Int).SetInt64(1000),
				blockSeed: new(big.Int).SetInt64(1000),
				nodeID:    int64(1000),
			},
			want: "7357219233906154",
		},
		{
			name: "CalculateNodeOrder:success2",
			args: args{
				score:     new(big.Int).SetInt64(2000),
				blockSeed: new(big.Int).SetInt64(3000),
				nodeID:    int64(10),
			},
			want: "2461325401382821",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateNodeOrder(tt.args.score, tt.args.blockSeed,
				tt.args.nodeID); !reflect.DeepEqual(fmt.Sprintf("%d", got), tt.want) {
				t.Errorf("CalculateNodeOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
