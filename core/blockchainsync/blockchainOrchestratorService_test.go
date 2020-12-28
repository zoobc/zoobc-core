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
