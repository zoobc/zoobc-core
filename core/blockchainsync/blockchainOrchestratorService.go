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
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type (
	BlockchainOrchestratorService struct {
		SpinechainSyncService          BlockchainSyncServiceInterface
		MainchainSyncService           BlockchainSyncServiceInterface
		BlockchainStatusService        service.BlockchainStatusServiceInterface
		SpineBlockManifestService      service.SpineBlockManifestServiceInterface
		FileDownloader                 p2p.FileDownloaderInterface
		MainchainSnapshotBlockServices service.SnapshotBlockServiceInterface
		Logger                         *log.Logger
	}
)

// NewBlockchainOrchestratorService returns service instance for orchestrating the blockchains
// as multiple blockhains are implemented in the application, this service controls the behavior
// of the blockchains so that the expected behavior is consistent within the application.
// In the future, this service may also be expanded to orchestrate the smithing activity of the blockchains
func NewBlockchainOrchestratorService(
	spinechainSyncService, mainchainSyncService BlockchainSyncServiceInterface,
	blockchainStatusService service.BlockchainStatusServiceInterface,
	spineBlockManifestService service.SpineBlockManifestServiceInterface,
	fileDownloader p2p.FileDownloaderInterface,
	mainchainSnapshotBlockServices service.SnapshotBlockServiceInterface,
	logger *log.Logger) *BlockchainOrchestratorService {
	return &BlockchainOrchestratorService{
		SpinechainSyncService:          spinechainSyncService,
		MainchainSyncService:           mainchainSyncService,
		BlockchainStatusService:        blockchainStatusService,
		SpineBlockManifestService:      spineBlockManifestService,
		FileDownloader:                 fileDownloader,
		MainchainSnapshotBlockServices: mainchainSnapshotBlockServices,
		Logger:                         logger,
	}
}

func (bos *BlockchainOrchestratorService) StartSyncChain(chainSyncService BlockchainSyncServiceInterface, timeout time.Duration) error {
	chainType := chainSyncService.GetBlockService().GetChainType()

	bos.Logger.Infof("downloading %s blocks...\n", chainType.GetName())

	go chainSyncService.Start()

	ticker := time.NewTicker(constant.BlockchainsyncCheckInterval)
	timeoutChannel := time.After(timeout)
CheckLoop:
	for {
		select {
		case <-ticker.C:
			if bos.BlockchainStatusService.IsFirstDownloadFinished(chainType) {
				ticker.Stop()
				break CheckLoop
			}
		// spine blocks shouldn't take that long to be downloaded. shutdown the node
		// TODO: add push notification to node owner that the node has shutdown because of network issues
		case <-timeoutChannel:
			if timeout != 0 {
				return blocker.NewBlocker(blocker.TimeoutExceeded, fmt.Sprintf("%s blocks sync timed out...\n", chainType.GetName()))
			}
		}
	}

	lastChainBlock, err := chainSyncService.GetBlockService().GetLastBlock()
	if err != nil {
		bos.Logger.Errorf("cannot get last %s block\n", chainType.GetName())
		return err
	}
	bos.Logger.Infof("finished downloading %s blocks. last height is %d\n", chainType.GetName(), lastChainBlock.Height)
	return nil
}

func (bos *BlockchainOrchestratorService) DownloadSnapshot(ct chaintype.ChainType) error {
	bos.Logger.Info("downloading snapshots...")
	lastSpineBlockManifest, err := bos.SpineBlockManifestService.GetLastSpineBlockManifest(ct,
		model.SpineBlockManifestType_Snapshot)
	if err != nil {
		bos.Logger.Errorf("db error: cannot get last spineBlockManifest for chaintype %s: %s\n",
			ct.GetName(), err.Error())
		return err
	}
	if lastSpineBlockManifest == nil {
		bos.Logger.Info("no lastSpineBlockManifest is found")
	} else {
		spinechainBlockService := (bos.SpinechainSyncService.GetBlockService()).(service.BlockServiceSpineInterface)
		err := spinechainBlockService.ValidateSpineBlockManifest(lastSpineBlockManifest)
		if err != nil {
			bos.Logger.Errorf("Invalid spineBlockManifest for chaintype %s Snapshot won't be downloaded. %s\n",
				ct.GetName(), err)
			return err
		}
		bos.Logger.Infof("found a Snapshot Spine Block Manifest for chaintype %s, "+
			"at height is %d. Start downloading...\n", ct.GetName(),
			lastSpineBlockManifest.ManifestReferenceHeight)
		snapshotFileInfo, err := bos.FileDownloader.DownloadSnapshot(ct, lastSpineBlockManifest)
		if err != nil {
			bos.Logger.Warning(err)
			return err
		}

		err = bos.MainchainSnapshotBlockServices.ImportSnapshotFile(snapshotFileInfo)
		if err != nil {
			bos.Logger.Warningf("error importing snapshot file for chaintype %s at height %d: %s\n", ct.GetName(),
				lastSpineBlockManifest.ManifestReferenceHeight, err.Error())
			return err
		}

	}
	return nil
}

func (bos *BlockchainOrchestratorService) Start() error {
	var err error
	// downloading spinechain and wait until the first download is complete
	// wait downloading snapshot and main blocks until node has finished downloading spine blocks
	err = bos.StartSyncChain(bos.SpinechainSyncService, constant.BlockchainsyncSpineTimeout)
	if err != nil {
		return err
	}

	lastMainBlock, err := bos.MainchainSyncService.GetBlockService().GetLastBlock()
	if err != nil {
		return fmt.Errorf("cannot get last main block: %s", err.Error())
	}
	if lastMainBlock.Height == 0 &&
		bos.MainchainSyncService.GetBlockService().GetChainType().HasSnapshots() {
		_ = bos.DownloadSnapshot(bos.MainchainSyncService.GetBlockService().GetChainType())
	}

	// start downloading mainchain
	err = bos.StartSyncChain(bos.MainchainSyncService, 0)
	if err != nil {
		return err
	}

	bos.Logger.Info("blockchain sync completed. unlocking smithing process...")
	bos.BlockchainStatusService.SetIsSmithingLocked(false)
	return nil
}
