package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type (
	MockNodeRegistrationServiceError struct {
		coreService.NodeRegistrationServiceInterface
	}
	MockNodeRegistrationServiceSuccess struct {
		coreService.NodeRegistrationServiceInterface
	}
	MockBlockService struct {
		coreService.BlockServiceInterface
	}
	MockP2pService struct {
		p2p.Peer2PeerServiceInterface
		HostToReturn          *model.Host
		PriorityPeersToReturn map[string]*model.Peer
	}
)

func (*MockNodeRegistrationServiceError) GetScrambleNodesByHeight(blockHeight uint32) (*model.ScrambledNodes, error) {
	return nil, errors.New("err")
}

func (*MockNodeRegistrationServiceSuccess) GetScrambleNodesByHeight(blockHeight uint32) (*model.ScrambledNodes, error) {
	return &model.ScrambledNodes{}, nil
}

func (*MockBlockService) GetLastBlock() (*model.Block, error) {
	return &model.Block{
		Height: 0,
	}, nil
}

func (m *MockP2pService) GetHostInfo() *model.Host {
	return m.HostToReturn
}

func (m *MockP2pService) GetPriorityPeers() map[string]*model.Peer {
	return m.PriorityPeersToReturn
}

func TestHostService_GetHostInfo(t *testing.T) {
	mockBlockService := make(map[int32]coreService.BlockServiceInterface)
	mockBlockService[int32(0)] = &MockBlockService{}

	hostToReturn := &model.Host{}
	priorityPeersToReturn := make(map[string]*model.Peer)

	type fields struct {
		Query                   query.ExecutorInterface
		P2pService              p2p.Peer2PeerServiceInterface
		BlockServices           map[int32]coreService.BlockServiceInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		BlockStateCache         storage.CacheStorageInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    *model.HostInfo
		wantErr bool
	}{
		{
			name: "GetHostInfo:error-lastBlockIsNil",
			fields: fields{
				BlockServices:   make(map[int32]coreService.BlockServiceInterface),
				BlockStateCache: storage.NewBlockStateStorage((&chaintype.MainChain{}).GetTypeInt(), model.Block{}),
			},
			wantErr: true,
		},
		{
			name: "GetHostInfo:error-GetScrambleNodesByHeight",
			fields: fields{
				BlockServices:           mockBlockService,
				NodeRegistrationService: &MockNodeRegistrationServiceError{},
				BlockStateCache:         storage.NewBlockStateStorage((&chaintype.MainChain{}).GetTypeInt(), model.Block{}),
			},
			wantErr: true,
		},
		{
			name: "GetHostInfo:success",
			fields: fields{
				BlockServices:           mockBlockService,
				NodeRegistrationService: &MockNodeRegistrationServiceSuccess{},
				P2pService: &MockP2pService{
					HostToReturn:          hostToReturn,
					PriorityPeersToReturn: priorityPeersToReturn,
				},
				BlockStateCache: storage.NewBlockStateStorage(
					(&chaintype.MainChain{}).GetTypeInt(),
					model.Block{
						BlockHash: []byte{1},
					},
				),
			},
			want: &model.HostInfo{
				Host: hostToReturn,
				ChainStatuses: []*model.ChainStatus{
					{
						ChainType: int32(0),
						Height:    0,
						LastBlock: &model.Block{
							BlockHash: []byte{1},
						},
					},
				},
				ScrambledNodes:       nil,
				ScrambledNodesHeight: 0,
				PriorityPeers:        priorityPeersToReturn,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := &HostService{
				Query:                   tt.fields.Query,
				P2pService:              tt.fields.P2pService,
				BlockServices:           tt.fields.BlockServices,
				NodeRegistrationService: tt.fields.NodeRegistrationService,
				BlockStateCache:         tt.fields.BlockStateCache,
			}
			got, err := hs.GetHostInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("HostService.GetHostInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HostService.GetHostInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
