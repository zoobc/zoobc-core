package blockchainsync

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

type Service struct {
	// isScanningBlockchain       bool
	ChainType            chaintype.ChainType
	PeerServiceClient    client.PeerServiceClientInterface
	PeerExplorer         strategy.PeerExplorerStrategyInterface
	BlockService         service.BlockServiceInterface
	BlockchainDownloader BlockchainDownloadInterface
	ForkingProcessor     ForkingProcessorInterface
	Logger               *log.Logger
}

func NewBlockchainSyncService(blockService service.BlockServiceInterface,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	queryExecutor query.ExecutorInterface,
	mempoolService service.MempoolServiceInterface,
	txActionSwitcher transaction.TypeActionSwitcher,
	logger *log.Logger,
	kvdb kvdb.KVExecutorInterface,
) *Service {
	return &Service{
		ChainType:         blockService.GetChainType(),
		BlockService:      blockService,
		PeerServiceClient: peerServiceClient,
		PeerExplorer:      peerExplorer,
		BlockchainDownloader: &BlockchainDownloader{
			ChainType:         blockService.GetChainType(),
			BlockService:      blockService,
			PeerServiceClient: peerServiceClient,
			PeerExplorer:      peerExplorer,
			Logger:            logger,
		},
		ForkingProcessor: &ForkingProcessor{
			ChainType:          blockService.GetChainType(),
			BlockService:       blockService,
			QueryExecutor:      queryExecutor,
			ActionTypeSwitcher: txActionSwitcher,
			MempoolService:     mempoolService,
			Logger:             logger,
			BlockPopper: &BlockPopper{
				ChainType:          blockService.GetChainType(),
				BlockService:       blockService,
				MempoolService:     mempoolService,
				QueryExecutor:      queryExecutor,
				ActionTypeSwitcher: txActionSwitcher,
				KVDB:               kvdb,
				Logger:             logger,
			},
		},
		Logger: logger,
	}
}

func (bss *Service) Start(runNext chan bool) {
	if bss.ChainType == nil {
		bss.Logger.Fatal("no chaintype")
	}
	if bss.PeerServiceClient == nil || bss.PeerExplorer == nil {
		bss.Logger.Fatal("no p2p service defined")
	}
	bss.GetMoreBlocksThread(runNext)
}

func (bss *Service) GetMoreBlocksThread(runNext chan bool) {

	defer func() {
		bss.Logger.Info("getMoreBlocksThread stopped")
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case download := <-runNext:
			if download {
				go bss.getMoreBlocks(runNext)
			}
		case <-sigs:
			return
		}
	}
}

func (bss *Service) getMoreBlocks(runNext chan bool) {
	// Pausing another process when they are using blockService.ChainWriteLock()
	bss.BlockService.ChainWriteLock(constant.BlockchainStatusSyncingBlock)
	defer bss.BlockService.ChainWriteUnlock()
	bss.Logger.Info("Get more blocks...")

	var (
		peerBlockchainInfo     *PeerBlockchainInfo
		otherPeerChainBlockIds []int64
		newLastBlock           *model.Block
		peerForkInfo           *PeerForkInfo
		lastBlock, err         = bss.BlockService.GetLastBlock()
	)
	// notify observer about start of blockchain download of this specific chain
	if err != nil {
		bss.Logger.Warn(fmt.Sprintf("failed to start getMoreBlocks go routine: %v", err))
	}
	if lastBlock == nil {
		bss.Logger.Warn("There is no genesis block found")
	}
	initialHeight := lastBlock.Height

	// Blockchain download
	for {
		needDownloadBlock := true
		peerBlockchainInfo, err = bss.BlockchainDownloader.GetPeerBlockchainInfo()
		if err != nil {
			bss.Logger.Infof("\nfailed to getPeerBlockchainInfo: %v\n\n", err)
			needDownloadBlock = false
		}

		newLastBlock = nil
		if needDownloadBlock {
			peerForkInfo, err = bss.BlockchainDownloader.DownloadFromPeer(peerBlockchainInfo.Peer, peerBlockchainInfo.ChainBlockIds,
				peerBlockchainInfo.CommonBlock)
			if err != nil {
				bss.Logger.Warnf("\nfailed to DownloadFromPeer: %v\n\n", err)
				break
			}

			if len(peerForkInfo.ForkBlocks) > 0 {

				err := bss.ForkingProcessor.ProcessFork(peerForkInfo.ForkBlocks, peerBlockchainInfo.CommonBlock, peerForkInfo.FeederPeer)
				if err != nil {
					bss.Logger.Warnf("\nfailed to ProcessFork: %v\n\n", err)
					break
				}
			}

			// confirming the node's blockchain state with other nodes
			var confirmations int32
			// counting the confirmations of the common block received with other peers he knows
			for _, peerToCheck := range bss.PeerExplorer.GetResolvedPeers() {
				if confirmations >= constant.DefaultNumberOfForkConfirmations {
					break
				}

				otherPeerChainBlockIds, err = bss.BlockchainDownloader.ConfirmWithPeer(peerToCheck, peerBlockchainInfo.CommonMilestoneBlockID)
				switch {
				case err != nil:
					bss.Logger.Warn(err)
				case len(otherPeerChainBlockIds) == 0:
					_, errDownload := bss.BlockchainDownloader.DownloadFromPeer(peerToCheck, otherPeerChainBlockIds, peerBlockchainInfo.CommonBlock)
					if errDownload != nil {
						bss.Logger.Warn(errDownload)
					}
				default:
					confirmations++
				}
			}

			newLastBlock, err = bss.BlockService.GetLastBlock()
			if err != nil {
				bss.Logger.Warnf("\nfailed to getMoreBlocks: %v\n\n", err)
				break
			}

			if lastBlock.ID == newLastBlock.ID {
				bss.Logger.Info("Did not accept peers's blocks, back to our own fork")
				break
			}
		}

		if bss.BlockchainDownloader.IsDownloadFinish(lastBlock) {
			bss.BlockchainDownloader.SetIsDownloading(false)
			bss.Logger.Infof("Finished %s blockchain download: %d blocks pulled", bss.ChainType.GetName(), lastBlock.Height-initialHeight)
			break
		}

		if newLastBlock == nil {
			break
		}

		lastBlock = newLastBlock
	}

	// TODO: Handle interruption and other exceptions
	time.Sleep(constant.GetMoreBlocksDelay * time.Second)
	runNext <- true
}
