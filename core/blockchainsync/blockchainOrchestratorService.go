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
