package strategy

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
)

func TestNewBlocksmithStrategy(t *testing.T) {
	type args struct {
		logger                  *log.Logger
		activeNodeRegistryCache storage.CacheStorageInterface
		chaintype               chaintype.ChainType
		rng                     *crypto.RandomNumberGenerator
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithStrategyMain
	}{
		{
			name: "Success",
			args: args{
				logger: nil,
			},
			want: NewBlocksmithStrategyMain(
				nil, nil, nil, nil, nil,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithStrategyMain(
				tt.args.logger, nil, tt.args.activeNodeRegistryCache, tt.args.rng,
				tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithStrategyMain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategyMain_convertRandomNumberToIndex(t *testing.T) {
	type fields struct {
		Chaintype                      chaintype.ChainType
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		Logger                         *log.Logger
		CurrentNodePublicKey           []byte
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
	type args struct {
		randNumber              int64
		activeNodeRegistryCount int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "convertRandomNumberToIndex:Success",
			fields: fields{},
			args: args{
				randNumber:              1002,
				activeNodeRegistryCount: 100,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				Chaintype:                      tt.fields.Chaintype,
				ActiveNodeRegistryCacheStorage: tt.fields.ActiveNodeRegistryCacheStorage,
				Logger:                         tt.fields.Logger,
				CurrentNodePublicKey:           tt.fields.CurrentNodePublicKey,
				candidates:                     tt.fields.candidates,
				me:                             tt.fields.me,
				lastBlockHash:                  tt.fields.lastBlockHash,
				rng:                            tt.fields.rng,
			}
			if got := bss.convertRandomNumberToIndex(tt.args.randNumber, tt.args.activeNodeRegistryCount); got != tt.want {
				t.Errorf("convertRandomNumberToIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockNodeRegistries = []storage.NodeRegistry{
		{
			Node: model.NodeRegistration{
				NodeID:        1,
				NodePublicKey: append([]byte{1}, make([]byte, 31)...),
			},
			ParticipationScore: 100,
		},
		{
			Node: model.NodeRegistration{
				NodeID:        2,
				NodePublicKey: append([]byte{2}, make([]byte, 31)...),
			},
			ParticipationScore: 200,
		},
		{
			Node: model.NodeRegistration{
				NodeID:        3,
				NodePublicKey: append([]byte{3}, make([]byte, 31)...),
			},
			ParticipationScore: 300,
		},
		{
			Node: model.NodeRegistration{
				NodeID:        4,
				NodePublicKey: append([]byte{4}, make([]byte, 31)...),
			},
			ParticipationScore: 400,
		},
		{
			Node: model.NodeRegistration{
				NodeID:        5,
				NodePublicKey: append([]byte{5}, make([]byte, 31)...),
			},
			ParticipationScore: 500,
		},
	}
)

type (
	mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates) GetAllItems(items interface{}) error {
	copyItems, _ := items.(*[]storage.NodeRegistry)
	*copyItems = mockNodeRegistries
	return nil
}

func TestBlocksmithStrategyMain_GetBlocksBlocksmiths(t *testing.T) {
	type fields struct {
		Chaintype                      chaintype.ChainType
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		Logger                         *log.Logger
		CurrentNodePublicKey           []byte
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
	type args struct {
		previousBlock *model.Block
		block         *model.Block
	}
	mainchain := &chaintype.MainChain{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Blocksmith
		wantErr bool
	}{
		{
			name: "Success:OneSmithingCandidate",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					BlockSeed: util.ConvertUint64ToBytes(12345),
					Timestamp: 0,
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod(),
					BlocksmithPublicKey: mockNodeRegistries[3].Node.GetNodePublicKey(),
				},
			},
			want: []*model.Blocksmith{
				{
					NodeID:        mockNodeRegistries[3].Node.NodeID,
					NodePublicKey: mockNodeRegistries[3].Node.NodePublicKey,
					Score:         big.NewInt(mockNodeRegistries[3].ParticipationScore),
				},
			},
			wantErr: false,
		},
		{
			name: "Fail:FiveSmithingCandidate-BlocksmithNotInListCandidate",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					BlockSeed: util.ConvertUint64ToBytes(12345),
					Timestamp: 0,
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod() + 4*mainchain.GetBlocksmithTimeGap(),
					BlocksmithPublicKey: mockNodeRegistries[4].Node.GetNodePublicKey(),
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Success:FiveSmithingCandidate - FirstOne",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					BlockSeed: util.ConvertUint64ToBytes(12345),
					Timestamp: 0,
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod() + 4*mainchain.GetBlocksmithTimeGap(),
					BlocksmithPublicKey: mockNodeRegistries[3].Node.GetNodePublicKey(),
				},
			},
			want: []*model.Blocksmith{
				{
					NodeID:        mockNodeRegistries[3].Node.NodeID,
					NodePublicKey: mockNodeRegistries[3].Node.NodePublicKey,
					Score:         big.NewInt(mockNodeRegistries[3].ParticipationScore),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				Chaintype:                      tt.fields.Chaintype,
				ActiveNodeRegistryCacheStorage: tt.fields.ActiveNodeRegistryCacheStorage,
				Logger:                         tt.fields.Logger,
				CurrentNodePublicKey:           tt.fields.CurrentNodePublicKey,
				candidates:                     tt.fields.candidates,
				me:                             tt.fields.me,
				lastBlockHash:                  tt.fields.lastBlockHash,
				rng:                            tt.fields.rng,
			}
			got, err := bss.GetBlocksBlocksmiths(tt.args.previousBlock, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlocksBlocksmiths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlocksBlocksmiths() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategyMain_GetSmithingRound(t *testing.T) {
	type fields struct {
		Chaintype                      chaintype.ChainType
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		Logger                         *log.Logger
		CurrentNodePublicKey           []byte
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
	type args struct {
		previousBlock *model.Block
		block         *model.Block
	}
	mainchain := &chaintype.MainChain{}
	spinechain := &chaintype.SpineChain{}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "OneRound",
			fields: fields{
				Chaintype: mainchain,
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				block: &model.Block{
					Timestamp: mainchain.GetSmithingPeriod(),
				},
			},
			want: 1,
		},
		{
			name: "MultipleRound",
			fields: fields{
				Chaintype: mainchain,
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				block: &model.Block{
					Timestamp: mainchain.GetSmithingPeriod() + 4*mainchain.GetBlocksmithTimeGap(),
				},
			},
			want: 5,
		},
		{
			name: "MultipleRound-Spinechain",
			fields: fields{
				Chaintype: spinechain,
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
				},
				block: &model.Block{
					Timestamp: spinechain.GetSmithingPeriod() + 4*spinechain.GetBlocksmithTimeGap(),
				},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				Chaintype:                      tt.fields.Chaintype,
				ActiveNodeRegistryCacheStorage: tt.fields.ActiveNodeRegistryCacheStorage,
				Logger:                         tt.fields.Logger,
				CurrentNodePublicKey:           tt.fields.CurrentNodePublicKey,
				candidates:                     tt.fields.candidates,
				me:                             tt.fields.me,
				lastBlockHash:                  tt.fields.lastBlockHash,
				rng:                            tt.fields.rng,
			}
			if got := bss.GetSmithingRound(tt.args.previousBlock, tt.args.block); got != tt.want {
				t.Errorf("GetSmithingRound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlocksmithStrategyMain_CanPersistBlock(t *testing.T) {
	type fields struct {
		Chaintype                      chaintype.ChainType
		ActiveNodeRegistryCacheStorage storage.CacheStorageInterface
		Logger                         *log.Logger
		CurrentNodePublicKey           []byte
		candidates                     []Candidate
		me                             Candidate
		lastBlockHash                  []byte
		rng                            *crypto.RandomNumberGenerator
	}
	mainchain := &chaintype.MainChain{}
	type args struct {
		previousBlock *model.Block
		block         *model.Block
		timestamp     int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "CanPersistBlock:True - FirstBlocksmith",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
					BlockSeed: util.ConvertUint64ToBytes(12345),
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod(),
					BlocksmithPublicKey: mockNodeRegistries[3].Node.GetNodePublicKey(),
				},
				timestamp: 25,
			},
			wantErr: false,
		},
		{
			name: "CanPersistBlock:CanPersistWithinBlockCreationTime - FirstBlocksmith",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
					BlockSeed: util.ConvertUint64ToBytes(12345),
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod(),
					BlocksmithPublicKey: mockNodeRegistries[3].Node.GetNodePublicKey(), // first blocksmith with provided blockseed
				},
				timestamp: mainchain.GetSmithingPeriod() + mainchain.GetBlocksmithBlockCreationTime(),
			},
			wantErr: false,
		},
		{
			name: "CanPersistBlock:Expired - Fourth Blocksmith",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
					BlockSeed: util.ConvertUint64ToBytes(12345),
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod() + mainchain.GetBlocksmithBlockCreationTime() + mainchain.GetBlocksmithNetworkTolerance(),
					BlocksmithPublicKey: mockNodeRegistries[3].Node.GetNodePublicKey(), // first blocksmith with provided blockseed
				},
				timestamp: mainchain.GetSmithingPeriod() + (4 * mainchain.GetBlocksmithTimeGap()) + mainchain.GetBlocksmithBlockCreationTime() +
					mainchain.GetBlocksmithNetworkTolerance() + 1,
			},
			wantErr: true,
		},
		{
			name: "CanPersistBlock:CanPersist - SecondBlocksmith",
			fields: fields{
				Chaintype:                      mainchain,
				ActiveNodeRegistryCacheStorage: &mockGetBlockBlocksmithsActiveNodeRegistryCache5Candidates{},
				Logger:                         nil,
				CurrentNodePublicKey:           nil,
				candidates:                     nil,
				me:                             Candidate{},
				lastBlockHash:                  nil,
				rng:                            crypto.NewRandomNumberGenerator(),
			},
			args: args{
				previousBlock: &model.Block{
					Timestamp: 0,
					BlockSeed: util.ConvertUint64ToBytes(12345),
				},
				block: &model.Block{
					Timestamp:           mainchain.GetSmithingPeriod() + mainchain.GetBlocksmithBlockCreationTime() + mainchain.GetBlocksmithNetworkTolerance(),
					BlocksmithPublicKey: mockNodeRegistries[4].Node.GetNodePublicKey(), // first blocksmith with provided blockseed
				},
				timestamp: mainchain.GetSmithingPeriod() + mainchain.GetBlocksmithBlockCreationTime() + mainchain.GetBlocksmithNetworkTolerance() + 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bss := &BlocksmithStrategyMain{
				Chaintype:                      tt.fields.Chaintype,
				ActiveNodeRegistryCacheStorage: tt.fields.ActiveNodeRegistryCacheStorage,
				Logger:                         tt.fields.Logger,
				CurrentNodePublicKey:           tt.fields.CurrentNodePublicKey,
				candidates:                     tt.fields.candidates,
				me:                             tt.fields.me,
				lastBlockHash:                  tt.fields.lastBlockHash,
				rng:                            tt.fields.rng,
			}
			if err := bss.CanPersistBlock(tt.args.previousBlock, tt.args.block, tt.args.timestamp); (err != nil) != tt.wantErr {
				t.Errorf("CanPersistBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
