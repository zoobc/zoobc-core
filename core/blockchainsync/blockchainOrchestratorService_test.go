package blockchainsync

import (
	"errors"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type (
	MockSpineBlockManifestServiceError struct {
		service.SpineBlockManifestServiceInterface
	}

	MockSpineBlockManifestServiceSuccessNoSpineBlockManifest struct {
		service.SpineBlockManifestServiceInterface
	}

	MockSpineBlockManifestServiceSuccessWithSpineBlockManifest struct {
		service.SpineBlockManifestServiceInterface
	}

	MockBlockchainSyncServiceError struct {
		BlockchainSyncServiceInterface
	}

	MockBlockchainSyncServiceSuccess struct {
		BlockchainSyncServiceInterface
	}

	MockBlockServiceError struct {
		service.BlockServiceInterface
		service.BlockServiceSpineInterface
	}

	MockBlockServiceSuccess struct {
		service.BlockServiceInterface
		service.BlockServiceSpineInterface
	}

	MockFileDownloaderError struct {
		p2p.FileDownloaderInterface
	}

	MockFileDownloaderSuccess struct {
		p2p.FileDownloaderInterface
	}

	MockMainchainSnapshotBlockServicesError struct {
		service.SnapshotBlockServiceInterface
	}

	MockMainchainSnapshotBlockServicesSuccess struct {
		service.SnapshotBlockServiceInterface
	}

	MockBlockchainStatusServiceNotFinished struct {
		service.BlockchainStatusServiceInterface
	}

	MockBlockchainStatusServiceFinished struct {
		service.BlockchainStatusServiceInterface
	}
)

func (*MockSpineBlockManifestServiceError) GetLastSpineBlockManifest(ct chaintype.ChainType,
	mbType model.SpineBlockManifestType) (*model.SpineBlockManifest, error) {
	return nil, errors.New("GetLastSpineBlockManifest error")
}

func (*MockSpineBlockManifestServiceSuccessNoSpineBlockManifest) GetLastSpineBlockManifest(ct chaintype.ChainType,
	mbType model.SpineBlockManifestType) (*model.SpineBlockManifest, error) {
	return nil, nil
}

func (*MockSpineBlockManifestServiceSuccessWithSpineBlockManifest) GetLastSpineBlockManifest(ct chaintype.ChainType,
	mbType model.SpineBlockManifestType) (*model.SpineBlockManifest, error) {
	return &model.SpineBlockManifest{}, nil
}

func (*MockBlockchainSyncServiceError) Start() {}

func (*MockBlockchainSyncServiceError) GetBlockService() service.BlockServiceInterface {
	return &MockBlockServiceError{}
}

func (*MockBlockServiceError) ValidateSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) error {
	return errors.New("ValidateSpineBlockManifest error")
}

func (*MockBlockServiceError) GetChainType() chaintype.ChainType {
	return &chaintype.SpineChain{}
}

func (*MockBlockServiceError) GetLastBlock() (*model.Block, error) {
	return nil, errors.New("GetLastBlock error")
}

func (*MockBlockchainSyncServiceSuccess) Start() {}

func (*MockBlockchainSyncServiceSuccess) GetBlockService() service.BlockServiceInterface {
	return &MockBlockServiceSuccess{}
}

func (*MockBlockServiceSuccess) ValidateSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) error {
	return nil
}

func (*MockBlockServiceSuccess) GetChainType() chaintype.ChainType {
	return &chaintype.SpineChain{}
}

func (*MockBlockServiceSuccess) GetLastBlock() (*model.Block, error) {
	return &model.Block{}, nil
}

func (*MockFileDownloaderError) DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) (*model.
	SnapshotFileInfo, error) {
	return nil, errors.New("DownloadSnapshot error")
}

func (*MockFileDownloaderSuccess) DownloadSnapshot(ct chaintype.ChainType, spineBlockManifest *model.SpineBlockManifest) (*model.
	SnapshotFileInfo, error) {
	return &model.SnapshotFileInfo{}, nil
}

func (*MockMainchainSnapshotBlockServicesError) ImportSnapshotFile(snapshotFileInfo *model.SnapshotFileInfo) error {
	return errors.New("ImportSnapshotFile error")
}

func (*MockMainchainSnapshotBlockServicesSuccess) ImportSnapshotFile(snapshotFileInfo *model.SnapshotFileInfo) error {
	return nil
}

func (*MockBlockchainStatusServiceNotFinished) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return false
}

func (*MockBlockchainStatusServiceFinished) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return true
}

func (*MockBlockchainStatusServiceFinished) SetIsSmithingLocked(isSmithingLocked bool) {}

func TestBlockchainOrchestratorService_DownloadSnapshot(t *testing.T) {
	type fields struct {
		SpinechainSyncService          BlockchainSyncServiceInterface
		MainchainSyncService           BlockchainSyncServiceInterface
		BlockchainStatusService        service.BlockchainStatusServiceInterface
		SpineBlockManifestService      service.SpineBlockManifestServiceInterface
		FileDownloader                 p2p.FileDownloaderInterface
		MainchainSnapshotBlockServices service.SnapshotBlockServiceInterface
		Logger                         *log.Logger
	}
	type args struct {
		ct chaintype.ChainType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error:lastSpineBlockManifest",
			fields: fields{
				SpineBlockManifestService: &MockSpineBlockManifestServiceError{},
				Logger:                    log.New(),
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			wantErr: true,
		},
		{
			name: "error:ValidateSpineBlockManifest",
			fields: fields{
				SpineBlockManifestService: &MockSpineBlockManifestServiceSuccessWithSpineBlockManifest{},
				Logger:                    log.New(),
				SpinechainSyncService:     &MockBlockchainSyncServiceError{},
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			wantErr: true,
		},
		{
			name: "error:DownloadSnapshot",
			fields: fields{
				SpineBlockManifestService: &MockSpineBlockManifestServiceSuccessWithSpineBlockManifest{},
				Logger:                    log.New(),
				SpinechainSyncService:     &MockBlockchainSyncServiceSuccess{},
				FileDownloader:            &MockFileDownloaderError{},
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			wantErr: true,
		},
		{
			name: "error:ImportSnapshotFile",
			fields: fields{
				SpineBlockManifestService:      &MockSpineBlockManifestServiceSuccessWithSpineBlockManifest{},
				Logger:                         log.New(),
				SpinechainSyncService:          &MockBlockchainSyncServiceSuccess{},
				FileDownloader:                 &MockFileDownloaderSuccess{},
				MainchainSnapshotBlockServices: &MockMainchainSnapshotBlockServicesError{},
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			wantErr: true,
		},
		{
			name: "success:noLastSpineBlockManifest",
			fields: fields{
				SpineBlockManifestService: &MockSpineBlockManifestServiceSuccessNoSpineBlockManifest{},
				Logger:                    log.New(),
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			wantErr: false,
		},
		{
			name: "success:allProcessSuccessGracefully",
			fields: fields{
				SpineBlockManifestService:      &MockSpineBlockManifestServiceSuccessWithSpineBlockManifest{},
				Logger:                         log.New(),
				SpinechainSyncService:          &MockBlockchainSyncServiceSuccess{},
				FileDownloader:                 &MockFileDownloaderSuccess{},
				MainchainSnapshotBlockServices: &MockMainchainSnapshotBlockServicesSuccess{},
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bos := &BlockchainOrchestratorService{
				SpinechainSyncService:          tt.fields.SpinechainSyncService,
				MainchainSyncService:           tt.fields.MainchainSyncService,
				BlockchainStatusService:        tt.fields.BlockchainStatusService,
				SpineBlockManifestService:      tt.fields.SpineBlockManifestService,
				FileDownloader:                 tt.fields.FileDownloader,
				MainchainSnapshotBlockServices: tt.fields.MainchainSnapshotBlockServices,
				Logger:                         tt.fields.Logger,
			}
			_ = bos.DownloadSnapshot(tt.args.ct)
			if err := bos.DownloadSnapshot(tt.args.ct); (err != nil) != tt.wantErr {
				t.Errorf("BlockchainOrchestratorService.DownloadSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockchainOrchestratorService_StartSyncChain(t *testing.T) {
	type fields struct {
		SpinechainSyncService          BlockchainSyncServiceInterface
		MainchainSyncService           BlockchainSyncServiceInterface
		BlockchainStatusService        service.BlockchainStatusServiceInterface
		SpineBlockManifestService      service.SpineBlockManifestServiceInterface
		FileDownloader                 p2p.FileDownloaderInterface
		MainchainSnapshotBlockServices service.SnapshotBlockServiceInterface
		Logger                         *log.Logger
	}
	type args struct {
		chainSyncService BlockchainSyncServiceInterface
		timeout          time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error:timeoutExceeds",
			fields: fields{
				BlockchainStatusService: &MockBlockchainStatusServiceNotFinished{},
				Logger:                  log.New(),
			},
			args: args{
				chainSyncService: &MockBlockchainSyncServiceError{},
				timeout:          1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "error:cannotGetLastBlockOfTheChain",
			fields: fields{
				BlockchainStatusService: &MockBlockchainStatusServiceFinished{},
				Logger:                  log.New(),
			},
			args: args{
				chainSyncService: &MockBlockchainSyncServiceError{},
				timeout:          0,
			},
			wantErr: true,
		},
		{
			name: "success:theOperationRunsWell",
			fields: fields{
				BlockchainStatusService: &MockBlockchainStatusServiceFinished{},
				Logger:                  log.New(),
			},
			args: args{
				chainSyncService: &MockBlockchainSyncServiceError{},
				timeout:          0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bos := &BlockchainOrchestratorService{
				SpinechainSyncService:          tt.fields.SpinechainSyncService,
				MainchainSyncService:           tt.fields.MainchainSyncService,
				BlockchainStatusService:        tt.fields.BlockchainStatusService,
				SpineBlockManifestService:      tt.fields.SpineBlockManifestService,
				FileDownloader:                 tt.fields.FileDownloader,
				MainchainSnapshotBlockServices: tt.fields.MainchainSnapshotBlockServices,
				Logger:                         tt.fields.Logger,
			}
			if err := bos.StartSyncChain(tt.args.chainSyncService, tt.args.timeout); (err != nil) != tt.wantErr {
				t.Errorf("BlockchainOrchestratorService.StartSyncChain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
