package blockchainsync

import (
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/monitoring"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

// TODO: rename into something more specific, such as SyncService
type (
	BlockchainSyncServiceInterface interface {
		GetBlockService() service.BlockServiceInterface
		Start()
	}

	BlockchainSyncService struct {
		ChainType               chaintype.ChainType
		PeerServiceClient       client.PeerServiceClientInterface
		PeerExplorer            strategy.PeerExplorerStrategyInterface
		BlockService            service.BlockServiceInterface
		BlockchainDownloader    BlockchainDownloadInterface
		ForkingProcessor        ForkingProcessorInterface
		Logger                  *log.Logger
		TransactionUtil         transaction.UtilInterface
		BlockchainStatusService service.BlockchainStatusServiceInterface
	}
)

func NewBlockchainSyncService(
	blockService service.BlockServiceInterface,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	logger *log.Logger,
	blockchainStatusService service.BlockchainStatusServiceInterface,
	blockchainDownloader BlockchainDownloadInterface,
	forkingProcessor ForkingProcessorInterface,
) *BlockchainSyncService {
	return &BlockchainSyncService{
		ChainType:               blockService.GetChainType(),
		BlockService:            blockService,
		PeerServiceClient:       peerServiceClient,
		PeerExplorer:            peerExplorer,
		BlockchainDownloader:    blockchainDownloader,
		ForkingProcessor:        forkingProcessor,
		Logger:                  logger,
		BlockchainStatusService: blockchainStatusService,
	}
}

func (bss *BlockchainSyncService) GetBlockService() service.BlockServiceInterface {
	return bss.BlockService
}

func (bss *BlockchainSyncService) Start() {
	if bss.ChainType == nil {
		bss.Logger.Fatal("no chaintype")
	}
	if bss.PeerServiceClient == nil || bss.PeerExplorer == nil {
		bss.Logger.Fatal("no p2p service defined")
	}
	// Give node time to connect to some peers
	time.Sleep(constant.BlockchainsyncWaitingTime)
	bss.GetMoreBlocksThread()
}

func (bss *BlockchainSyncService) GetMoreBlocksThread() {
	defer func() {
		bss.Logger.Info("getMoreBlocksThread stopped")
	}()

	for {
		bss.getMoreBlocks()
		time.Sleep(constant.GetMoreBlocksDelay * time.Second)
	}
}

func (bss *BlockchainSyncService) getMoreBlocks() {
	// Pausing another process when they are using blockService.ChainWriteLock()
	bss.BlockService.ChainWriteLock(constant.BlockchainStatusSyncingBlock)
	defer bss.BlockService.ChainWriteUnlock(constant.BlockchainStatusSyncingBlock)
	bss.Logger.Info("Get more blocks...")
	monitoring.ResetMainchainDownloadCycleDebugger(bss.ChainType)

	var (
		peerBlockchainInfo     *PeerBlockchainInfo
		otherPeerChainBlockIds []int64
		newLastBlock           *model.Block
		peerForkInfo           *PeerForkInfo
		lastBlock, err         = bss.BlockService.GetLastBlock()
	)
	// notify observer about start of blockchain download of this specific chain
	if err != nil {
		bss.Logger.Fatalf("getMoreBlocks:GetLastBlock()-Fail: error: %v", err)
	}
	if lastBlock == nil {
		bss.Logger.Fatalf("getMoreBlocks:GetLastBlock()-NoError-LastBlockNil: error: %v", err)
	}
	initialHeight := lastBlock.Height

	// Blockchain download
	for {
		monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 1)
		needDownloadBlock := true
		peerBlockchainInfo, err = bss.BlockchainDownloader.GetPeerBlockchainInfo()
		monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 2)
		if err != nil {
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 3)
			needDownloadBlock = false
			errCasted, ok := err.(blocker.Blocker)
			if !ok {
				errCasted = blocker.NewBlocker(blocker.P2PNetworkConnectionErr, err.Error()).(blocker.Blocker)
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 4)
			switch errCasted.Type {
			case blocker.P2PPeerError:
				// this will allow the node to start smithing if it fails to connect to the p2p network,
				// eg. he is the first node. if later on he can connect, it will try resolve the fork normally
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 5)
				bss.Logger.Info(err)
				bss.BlockchainStatusService.SetIsSmithingLocked(false)
				bss.Logger.Info(errCasted.Message)
			case blocker.ChainValidationErr:
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 6)
				bss.Logger.Infof("peer %s:%d: %s",
					peerBlockchainInfo.Peer.GetInfo().Address,
					peerBlockchainInfo.Peer.GetInfo().Port,
					errCasted.Message)
			default:
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 7)
				bss.Logger.Infof("ChainSync: failed to getPeerBlockchainInfo: %v", err)
			}
		}

		monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 8)
		newLastBlock = nil
		if needDownloadBlock && len(peerBlockchainInfo.ChainBlockIds) > 0 {
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 9)
			peerForkInfo, err = bss.BlockchainDownloader.DownloadFromPeer(peerBlockchainInfo.Peer, peerBlockchainInfo.ChainBlockIds,
				peerBlockchainInfo.CommonBlock)
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 10)
			if err != nil {
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 11)
				bss.Logger.Warnf("ChainSync: failed to DownloadFromPeer: %v\n\n", err)
				break
			}

			if len(peerForkInfo.ForkBlocks) > 0 {
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 11)
				err := bss.ForkingProcessor.ProcessFork(peerForkInfo.ForkBlocks, peerBlockchainInfo.CommonBlock, peerForkInfo.FeederPeer)
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 12)
				if err != nil {
					monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 13)
					bss.Logger.Warnf("\nfailed to ProcessFork: %v\n\n", err)
					break
				}
			}

			// confirming the node's blockchain state with other nodes
			var confirmations int32
			// counting the confirmations of the common block received with other peers he knows
			for _, peerToCheck := range bss.PeerExplorer.GetResolvedPeers() {
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 14)
				if confirmations >= constant.DefaultNumberOfForkConfirmations {
					break
				}

				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 15)
				otherPeerChainBlockIds, err = bss.BlockchainDownloader.ConfirmWithPeer(peerToCheck, peerBlockchainInfo.CommonMilestoneBlockID)
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 16)
				switch {
				case err != nil:
					monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 17)
					bss.Logger.Warn(err)
				case len(otherPeerChainBlockIds) != 0:
					monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 17)
					_, errDownload := bss.BlockchainDownloader.DownloadFromPeer(peerToCheck, otherPeerChainBlockIds, peerBlockchainInfo.CommonBlock)
					monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 18)
					if errDownload != nil {
						bss.Logger.Warn(errDownload)
					}
				default:
					monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 19)
					confirmations++
				}
			}

			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 20)
			newLastBlock, err = bss.BlockService.GetLastBlock()
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 21)
			if err != nil {
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 22)
				bss.Logger.Warnf("\nfailed to getMoreBlocks: %v\n\n", err)
				break
			}

			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 23)
			if lastBlock.ID == newLastBlock.ID {
				monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 24)
				bss.Logger.Info("Did not accept peers's blocks, back to our own fork")
				break
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 25)
		}

		monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 26)
		if bss.BlockchainDownloader.IsDownloadFinish(lastBlock) {
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 27)
			bss.BlockchainStatusService.SetIsDownloading(bss.ChainType, false)
			// only set the first download finished = true once
			bss.BlockchainStatusService.SetFirstDownloadFinished(bss.ChainType, true)
			bss.Logger.Infof("Finished %s blocks download: %d blocks pulled", bss.ChainType.GetName(),
				lastBlock.Height-initialHeight)
			break
		}

		monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 28)
		if newLastBlock == nil {
			monitoring.IncrementMainchainDownloadCycleDebugger(bss.ChainType, 29)
			break
		}

		lastBlock = newLastBlock
	}
}
