// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package service

import (
	"errors"
	"reflect"
	"testing"

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

type (
	mockScrambleNodeServiceGetScrambleNodeByHeightError struct {
		coreService.ScrambleNodeService
	}
	mockScrambleNodeServiceGetScrambleNodeByHeightSuccess struct {
		coreService.ScrambleNodeService
	}
)

func (*mockScrambleNodeServiceGetScrambleNodeByHeightError) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return nil, errors.New("mockedError")
}

func (*mockScrambleNodeServiceGetScrambleNodeByHeightSuccess) GetScrambleNodesByHeight(
	blockHeight uint32,
) (*model.ScrambledNodes, error) {
	return mockGoodScrambledNodes, nil
}

var (
	indexScramble = []int{
		0: 0,
		1: 1,
	}

	mockGoodScrambledNodes = &model.ScrambledNodes{
		AddressNodes: []*model.Peer{
			0: {
				Info: &model.Node{
					ID:      int64(111),
					Address: "127.0.0.1",
					Port:    8000,
				},
			},
			1: {
				Info: &model.Node{
					ID:      int64(222),
					Address: "127.0.0.1",
					Port:    3001,
				},
			},
		},
		IndexNodes: map[string]*int{
			"111": &indexScramble[0],
			"222": &indexScramble[1],
		},
	}
)

func TestHostService_GetHostInfo(t *testing.T) {
	var (
		mockBlockService       = make(map[int32]coreService.BlockServiceInterface)
		mockBlockStateStorages = make(map[int32]storage.CacheStorageInterface)
		hostToReturn           = &model.Host{}
		priorityPeersToReturn  = make(map[string]*model.Peer)
	)
	mockBlockService[int32(0)] = &MockBlockService{}
	mockBlockStateStorages[int32(0)] = storage.NewBlockStateStorage()
	_ = mockBlockStateStorages[int32(0)].SetItem(nil, model.Block{BlockHash: []byte{1}})

	type fields struct {
		Query                   query.ExecutorInterface
		P2pService              p2p.Peer2PeerServiceInterface
		BlockServices           map[int32]coreService.BlockServiceInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		BlockStateStorages      map[int32]storage.CacheStorageInterface
		ScrambleNodeService     coreService.ScrambleNodeServiceInterface
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
				BlockServices:      make(map[int32]coreService.BlockServiceInterface),
				BlockStateStorages: mockBlockStateStorages,
			},
			wantErr: true,
		},
		{
			name: "GetHostInfo:error-GetScrambleNodesByHeight",
			fields: fields{
				BlockServices:           mockBlockService,
				NodeRegistrationService: &MockNodeRegistrationServiceError{},
				BlockStateStorages:      mockBlockStateStorages,
				ScrambleNodeService:     &mockScrambleNodeServiceGetScrambleNodeByHeightError{},
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
				BlockStateStorages:  mockBlockStateStorages,
				ScrambleNodeService: &mockScrambleNodeServiceGetScrambleNodeByHeightSuccess{},
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
				ScrambledNodes:       mockGoodScrambledNodes.AddressNodes,
				ScrambledNodesHeight: mockGoodScrambledNodes.BlockHeight,
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
				BlockStateStorages:      tt.fields.BlockStateStorages,
				ScrambleNodeService:     tt.fields.ScrambleNodeService,
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
