package blockchainsync

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
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

func NewBlockchainOrchestratorService(
	spinechainSyncService BlockchainSyncServiceInterface,
	mainchainSyncService BlockchainSyncServiceInterface,
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

func (bos *BlockchainOrchestratorService) SyncChain(chainSyncService BlockchainSyncServiceInterface, timeout time.Duration) {
	chainType := chainSyncService.GetBlockService().GetChainType()

	bos.Logger.Infof("downloading %s blocks...\n", chainType.GetName())
	log.Infof("downloading %s blocks...\n", chainType.GetName())
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
				bos.Logger.Fatalf("%s blocks sync timed out...\n", chainType.GetName())
			}
		}
	}

	lastChainBlock, err := chainSyncService.GetBlockService().GetLastBlock()
	if err != nil {
		bos.Logger.Errorf("cannot get last %s block\n", chainType.GetName())
		os.Exit(1)
	}
	bos.Logger.Infof("finished downloading %s blocks. last height is %d\n", chainType.GetName(), lastChainBlock.Height)
}

func (bos *BlockchainOrchestratorService) DownloadSnapshot(ct chaintype.ChainType) {
	bos.Logger.Info("dowloading snapshots...")
	log.Info("dowloading snapshots...")
	lastSpineBlockManifest, err := bos.SpineBlockManifestService.GetLastSpineBlockManifest(ct,
		model.SpineBlockManifestType_Snapshot)
	if err != nil {
		bos.Logger.Errorf("db error: cannot get last spineBlockManifest for chaintype %s: %s\n",
			ct.GetName(), err.Error())
		return
	}
	if lastSpineBlockManifest == nil {
		bos.Logger.Info("no lastSpineBlockManifest is found")
		log.Info("no lastSpineBlockManifest is found")
	} else {
		spinechainBlockService := (bos.SpinechainSyncService.GetBlockService()).(service.BlockServiceSpineInterface)
		err := spinechainBlockService.ValidateSpineBlockManifest(lastSpineBlockManifest)
		if err != nil {
			bos.Logger.Errorf("Invalid spineBlockManifest for chaintype %s Snapshot won't be downloaded. %s\n",
				ct.GetName(), err)
		} else {
			bos.Logger.Infof("found a Snapshot Spine Block Manifest for chaintype %s, "+
				"at height is %d. Start downloading...\n", ct.GetName(),
				lastSpineBlockManifest.SpineBlockManifestHeight)
			snapshotFileInfo, err := bos.FileDownloader.DownloadSnapshot(ct, lastSpineBlockManifest)
			if err != nil {
				bos.Logger.Warning(err)
			} else {
				log.Info("applying snapshots...")
				if err := bos.MainchainSnapshotBlockServices.ImportSnapshotFile(snapshotFileInfo); err != nil {
					bos.Logger.Warningf("error importing snapshot file for chaintype %s at height %d\n", ct.GetName(),
						lastSpineBlockManifest.SpineBlockManifestHeight)
				}
			}
		}
	}
}

func (bos *BlockchainOrchestratorService) Start() {
	// downloading spinechain and wait until the first download is complete
	// wait downloading snapshot and main blocks until node has finished downloading spine blocks
	bos.SyncChain(bos.SpinechainSyncService, constant.BlockchainsyncSpineTimeout)

	lastMainBlock, err := bos.MainchainSyncService.GetBlockService().GetLastBlock()
	if err != nil {
		bos.Logger.Fatal("cannot get last main block")
	}
	if lastMainBlock.Height == 0 &&
		bos.MainchainSyncService.GetBlockService().GetChainType().HasSnapshots() {
		bos.DownloadSnapshot(bos.MainchainSyncService.GetBlockService().GetChainType())
	}

	// start downloading mainchain
	bos.SyncChain(bos.MainchainSyncService, 0)

	bos.Logger.Info("blockchain sync completed. unlocking smithing process...")
	log.Info("blockchain sync completed. unlocking smithing process...")
	log.Info("now the blockchain operates normally")
	bos.BlockchainStatusService.SetIsSmithingLocked(false)
}
