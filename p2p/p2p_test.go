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
package p2p

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

type (
	p2pMockPeerExplorer struct {
		strategy.PeerExplorerStrategyInterface
		noResolvedPeers bool
		oneResolvedPeer bool
	}
	p2pMockPeerServiceClient struct {
		client.PeerServiceClient
		noFailedDownloads bool
		downloadErr       bool
		returnInvalidData bool
	}
	p2pMockFileService struct {
		coreService.FileService
		saveFileFailed bool
		retFileName    string
	}
)

var (
	p2pP1 = &model.Peer{
		Info: &model.Node{
			ID:      1111,
			Port:    8080,
			Address: "127.0.0.1",
		},
	}
	p2pP2 = &model.Peer{
		Info: &model.Node{
			ID:      2222,
			Port:    9090,
			Address: "127.0.0.2",
		},
	}
	p2pChunk1Bytes = []byte{
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	}
	p2pChunk2Bytes = []byte{
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	}
	p2pChunk2InvalidBytes = []byte{
		2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0,
	}
)

func (p2pMpe *p2pMockPeerExplorer) GetResolvedPeers() map[string]*model.Peer {
	if p2pMpe.noResolvedPeers {
		return nil
	}

	peers := make(map[string]*model.Peer)
	peers[p2pP1.Info.Address] = p2pP1
	if !p2pMpe.oneResolvedPeer {
		peers[p2pP2.Info.Address] = p2pP2
	}
	return peers
}

func (p2pMpsc *p2pMockPeerServiceClient) RequestDownloadFile(*model.Peer, []byte, []string) (*model.FileDownloadResponse, error) {
	var (
		failed           []string
		downloadedChunks [][]byte
	)
	if p2pMpsc.downloadErr {
		return nil, errors.New("RequestDownloadFileFailed")
	}
	if p2pMpsc.returnInvalidData {
		downloadedChunks = [][]byte{
			p2pChunk1Bytes,
			p2pChunk2InvalidBytes,
		}
	} else {
		downloadedChunks = [][]byte{
			p2pChunk1Bytes,
			p2pChunk2Bytes,
		}
	}
	if !p2pMpsc.noFailedDownloads {
		failed = []string{
			"testChunkFailed1",
		}
	}
	return &model.FileDownloadResponse{
		FileChunks: downloadedChunks,
		Failed:     failed,
	}, nil
}

func (p2pMfs *p2pMockFileService) GetFileNameFromBytes(fileBytes []byte) string {
	if bytes.Equal(fileBytes, p2pChunk1Bytes) {
		return "testChunk1"
	}
	if bytes.Equal(fileBytes, p2pChunk2Bytes) {
		return "testChunk2"
	}
	if bytes.Equal(fileBytes, p2pChunk2InvalidBytes) {
		return "testChunk2Invalid"
	}
	return p2pMfs.retFileName
}
func (p2pMfs *p2pMockFileService) GetFileNameFromHash(fileBytes []byte) string {
	if bytes.Equal(fileBytes, p2pChunk1Bytes) {
		return "testChunk1"
	}
	if bytes.Equal(fileBytes, p2pChunk2Bytes) {
		return "testChunk2"
	}
	if bytes.Equal(fileBytes, p2pChunk2InvalidBytes) {
		return "testChunk2Invalid"
	}
	return p2pMfs.retFileName
}

func (p2pMfs *p2pMockFileService) SaveSnapshotChunks(dir string, chunks [][]byte) (fileHashes [][]byte, err error) {
	if p2pMfs.saveFileFailed {
		return nil, errors.New("SaveBytesToFileFailed")
	}
	return nil, nil

}

func TestPeer2PeerService_DownloadFilesFromPeer(t *testing.T) {
	type fields struct {
		PeerExplorer      strategy.PeerExplorerStrategyInterface
		PeerServiceClient client.PeerServiceClientInterface
		Logger            *log.Logger
		TransactionUtil   transaction.UtilInterface
		FileService       coreService.FileServiceInterface
	}
	type args struct {
		fullHash        []byte
		fileChunksNames []string
		validNodeIDs    map[int64]bool
		maxRetryCount   uint32
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFailed []string
		wantErr    bool
	}{
		{
			name: "DownloadFilesFromPeer:success-{noRetry}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 0,
				validNodeIDs: map[int64]bool{
					1111: true,
					2222: true,
				},
			},
			fields: fields{
				Logger:            log.New(),
				PeerExplorer:      &p2pMockPeerExplorer{},
				FileService:       &p2pMockFileService{},
				PeerServiceClient: &p2pMockPeerServiceClient{},
			},
			wantFailed: []string{
				"testChunkFailed1",
			},
		},
		{
			name: "DownloadFilesFromPeer:success-{WithRetry}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 1,
				validNodeIDs: map[int64]bool{
					1111: true,
					2222: true,
				},
			},
			fields: fields{
				Logger:            log.New(),
				PeerExplorer:      &p2pMockPeerExplorer{},
				FileService:       &p2pMockFileService{},
				PeerServiceClient: &p2pMockPeerServiceClient{},
			},
			wantFailed: []string{
				"testChunkFailed1",
			},
		},
		{
			name: "DownloadFilesFromPeer:success-{WithRetryNoFailedDownloads}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 1,
				validNodeIDs: map[int64]bool{
					1111: true,
					2222: true,
				},
			},
			fields: fields{
				Logger:       log.New(),
				PeerExplorer: &p2pMockPeerExplorer{},
				FileService:  &p2pMockFileService{},
				PeerServiceClient: &p2pMockPeerServiceClient{
					noFailedDownloads: true,
				},
			},
		},
		{
			name: "DownloadFilesFromPeer:fail-{DownloadFailed - only one resolved peer}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 0,
				validNodeIDs: map[int64]bool{
					1111: true,
					2222: true,
				},
			},
			fields: fields{
				Logger: log.New(),
				PeerExplorer: &p2pMockPeerExplorer{
					oneResolvedPeer: true,
				},
				FileService: &p2pMockFileService{},
				PeerServiceClient: &p2pMockPeerServiceClient{
					downloadErr: true,
				},
			},
			wantErr: true,
		},
		{
			name: "DownloadFilesFromPeer:success-{DownloadedInvalidFileChunk}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 0,
				validNodeIDs: map[int64]bool{
					1111: true,
					2222: true,
				},
			},
			fields: fields{
				Logger:       log.New(),
				PeerExplorer: &p2pMockPeerExplorer{},
				FileService:  &p2pMockFileService{},
				PeerServiceClient: &p2pMockPeerServiceClient{
					returnInvalidData: true,
				},
			},
			wantFailed: []string{
				"testChunk1",
				"testChunk2",
				"testChunk3",
			},
		},
		{
			name: "DownloadFilesFromPeer:fail-{SaveFileFailed}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 0,
				validNodeIDs: map[int64]bool{
					1111: true,
					2222: true,
				},
			},
			fields: fields{
				Logger:       log.New(),
				PeerExplorer: &p2pMockPeerExplorer{},
				FileService: &p2pMockFileService{
					saveFileFailed: true,
				},
				PeerServiceClient: &p2pMockPeerServiceClient{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Peer2PeerService{
				PeerExplorer:      tt.fields.PeerExplorer,
				PeerServiceClient: tt.fields.PeerServiceClient,
				Logger:            tt.fields.Logger,
				TransactionUtil:   tt.fields.TransactionUtil,
				FileService:       tt.fields.FileService,
			}
			gotFailed, err := s.DownloadFilesFromPeer(tt.args.fullHash, tt.args.fileChunksNames, tt.args.validNodeIDs, tt.args.maxRetryCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Peer2PeerService.DownloadFilesFromPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFailed, tt.wantFailed) {
				t.Errorf("Peer2PeerService.DownloadFilesFromPeer() = %v, want %v", gotFailed, tt.wantFailed)
			}
		})
	}
}
