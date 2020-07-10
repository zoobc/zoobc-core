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
	peers[p2pP2.Info.Address] = p2pP2
	return peers
}

func (p2pMpsc *p2pMockPeerServiceClient) RequestDownloadFile(
	destPeer *model.Peer,
	fileChunkNames []string,
) (*model.FileDownloadResponse, error) {
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

func (p2pMfs *p2pMockFileService) SaveBytesToFile(fileBasePath, filename string, b []byte) error {
	if p2pMfs.saveFileFailed {
		return errors.New("SaveBytesToFileFailed")
	}
	return nil
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
		fileChunksNames []string
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
			name: "DownloadFilesFromPeer:fail-{DownloadFailed}",
			args: args{
				fileChunksNames: []string{
					"testChunk1",
					"testChunk2",
					"testChunk3",
				},
				maxRetryCount: 0,
			},
			fields: fields{
				Logger:       log.New(),
				PeerExplorer: &p2pMockPeerExplorer{},
				FileService:  &p2pMockFileService{},
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
			gotFailed, err := s.DownloadFilesFromPeer(tt.args.fileChunksNames, tt.args.maxRetryCount)
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
