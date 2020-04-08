package p2p

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
)

func TestNewFileDownloader(t *testing.T) {
	type args struct {
		p2pService              Peer2PeerServiceInterface
		fileService             service.FileServiceInterface
		logger                  *log.Logger
		blockchainStatusService service.BlockchainStatusServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *FileDownloader
	}{
		{
			name: "NewFileDownloader:success",
			args: args{
				p2pService:              &Peer2PeerService{},
				blockchainStatusService: &service.BlockchainStatusService{},
				logger:                  &log.Logger{},
				fileService:             &service.FileService{},
			},
			want: &FileDownloader{
				FileService:             &service.FileService{},
				Logger:                  &log.Logger{},
				BlockchainStatusService: &service.BlockchainStatusService{},
				P2pService:              &Peer2PeerService{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileDownloader(tt.args.p2pService, tt.args.fileService, tt.args.logger,
				tt.args.blockchainStatusService); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileDownloader() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockFileService struct {
		service.FileService
		successParseFileChunkHashes bool
		emptyRes                    bool
	}
	mockP2pService struct {
		Peer2PeerService
		success bool
	}
)

var (
	fdChunk1Hash = []byte{
		1, 1, 1, 249, 145, 71, 241, 88, 208, 4, 80, 132, 88, 43, 189, 93, 19, 104, 255, 61, 177, 177, 223,
		188, 144, 9, 73, 75, 6, 1, 1, 1,
	}
	fdChunk2Hash = []byte{
		2, 2, 2, 249, 145, 71, 241, 88, 208, 4, 80, 132, 88, 43, 189, 93, 19, 104, 255, 61, 177, 177, 223,
		188, 144, 9, 73, 75, 6, 2, 2, 2,
	}
)

func (mfs *mockFileService) ParseFileChunkHashes(fileHashes []byte, hashLength int) (fileHashesAry [][]byte, err error) {
	if mfs.emptyRes {
		return nil, nil
	}
	if mfs.successParseFileChunkHashes {
		return [][]byte{
			fdChunk1Hash,
			fdChunk2Hash,
		}, nil
	}
	return nil, errors.New("ParseFileChunkHashesFailed")
}

func (mfs *mockFileService) GetFileNameFromHash(fileHash []byte) string {
	return "testFileName"
}

func (mp2p *mockP2pService) DownloadFilesFromPeer(fileChunksNames []string, retryCount uint32) (failed []string, err error) {
	failed = make([]string, 0)
	if mp2p.success {
		return
	}
	return []string{"testFailedFile1"}, errors.New("DownloadFilesFromPeerFailed")
}

func TestFileDownloader_DownloadSnapshot(t *testing.T) {
	type fields struct {
		FileService             service.FileServiceInterface
		P2pService              Peer2PeerServiceInterface
		BlockchainStatusService service.BlockchainStatusServiceInterface
		Logger                  *log.Logger
	}
	type args struct {
		ct                 chaintype.ChainType
		spineBlockManifest *model.SpineBlockManifest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "DownloadSnapshot:success",
			args: args{
				ct:                 &chaintype.MainChain{},
				spineBlockManifest: &model.SpineBlockManifest{},
			},
			fields: fields{
				FileService: &mockFileService{
					successParseFileChunkHashes: true,
				},
				P2pService: &mockP2pService{
					success: true,
				},
				BlockchainStatusService: service.NewBlockchainStatusService(false, log.New()),
			},
		},
		{
			name: "DownloadSnapshot:fail-{ParseFileChunkHashesErr}",
			args: args{
				ct:                 &chaintype.MainChain{},
				spineBlockManifest: &model.SpineBlockManifest{},
			},
			fields: fields{
				FileService: &mockFileService{
					successParseFileChunkHashes: false,
				},
				P2pService: &mockP2pService{
					success: true,
				},
				BlockchainStatusService: service.NewBlockchainStatusService(false, log.New()),
			},
			wantErr: true,
		},
		{
			name: "DownloadSnapshot:fail-{ParseFileChunkHashesEmptyResult}",
			args: args{
				ct:                 &chaintype.MainChain{},
				spineBlockManifest: &model.SpineBlockManifest{},
			},
			fields: fields{
				FileService: &mockFileService{
					successParseFileChunkHashes: true,
					emptyRes:                    true,
				},
				P2pService: &mockP2pService{
					success: true,
				},
				BlockchainStatusService: service.NewBlockchainStatusService(false, log.New()),
			},
			wantErr: true,
		},
		{
			name: "DownloadSnapshot:fail-{DownloadFilesFromPeer}",
			args: args{
				ct:                 &chaintype.MainChain{},
				spineBlockManifest: &model.SpineBlockManifest{},
			},
			fields: fields{
				FileService: &mockFileService{
					successParseFileChunkHashes: true,
				},
				P2pService: &mockP2pService{
					success: false,
				},
				Logger:                  log.New(),
				BlockchainStatusService: service.NewBlockchainStatusService(false, log.New()),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &FileDownloader{
				FileService:             tt.fields.FileService,
				P2pService:              tt.fields.P2pService,
				BlockchainStatusService: tt.fields.BlockchainStatusService,
				Logger:                  tt.fields.Logger,
			}
			if _, err := ss.DownloadSnapshot(tt.args.ct, tt.args.spineBlockManifest); (err != nil) != tt.wantErr {
				t.Errorf("FileDownloader.DownloadSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
