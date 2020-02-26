package blockchainsync

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
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

// TODO: rename into something more specific, such as SyncService
type Service struct {
	// isScanningBlockchain       bool
	ChainType              chaintype.ChainType
	PeerServiceClient      client.PeerServiceClientInterface
	PeerExplorer           strategy.PeerExplorerStrategyInterface
	BlockService           service.BlockServiceInterface
	BlockchainDownloader   BlockchainDownloadInterface
	ForkingProcessor       ForkingProcessorInterface
	Logger                 *log.Logger
	TransactionUtil        transaction.UtilInterface
	BlockTypeStatusService service.BlockTypeStatusServiceInterface
}

func NewBlockchainSyncService(
	blockService service.BlockServiceInterface,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	queryExecutor query.ExecutorInterface,
	mempoolService service.MempoolServiceInterface,
	txActionSwitcher transaction.TypeActionSwitcher,
	logger *log.Logger,
	kvdb kvdb.KVExecutorInterface,
	transactionUtil transaction.UtilInterface,
	transactionCoreService service.TransactionCoreServiceInterface,
	blockTypeStatusService service.BlockTypeStatusServiceInterface,
) *Service {
	return &Service{
		ChainType:         blockService.GetChainType(),
		BlockService:      blockService,
		PeerServiceClient: peerServiceClient,
		PeerExplorer:      peerExplorer,
		BlockchainDownloader: NewBlockchainDownloader(
			blockService,
			peerServiceClient,
			peerExplorer,
			logger,
			blockTypeStatusService,
		),
		ForkingProcessor: &ForkingProcessor{
			ChainType:             blockService.GetChainType(),
			BlockService:          blockService,
			QueryExecutor:         queryExecutor,
			ActionTypeSwitcher:    txActionSwitcher,
			MempoolService:        mempoolService,
			KVExecutor:            kvdb,
			PeerExplorer:          peerExplorer,
			Logger:                logger,
			TransactionUtil:       transactionUtil,
			TransactionCorService: transactionCoreService,
		},
		Logger:                 logger,
		BlockTypeStatusService: blockTypeStatusService,
	}
}

func (bss *Service) Start() {
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

func (bss *Service) GetMoreBlocksThread() {
	defer func() {
		bss.Logger.Info("getMoreBlocksThread stopped")
	}()

	for {
		bss.getMoreBlocks()
		time.Sleep(constant.GetMoreBlocksDelay * time.Second)
	}
}

func (bss *Service) getMoreBlocks() {
	// Pausing another process when they are using blockService.ChainWriteLock()
	bss.BlockService.ChainWriteLock(constant.BlockchainStatusSyncingBlock)
	defer bss.BlockService.ChainWriteUnlock(constant.BlockchainStatusSyncingBlock)
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
		// break
		needDownloadBlock := true
		peerBlockchainInfo, err = bss.BlockchainDownloader.GetPeerBlockchainInfo()
		if err != nil {
			needDownloadBlock = false
			errCasted := err.(blocker.Blocker)
			switch errCasted.Type {
			case blocker.P2PNetworkConnectionErr:
				// this will allow the node to start smithing if it fails to connect to the p2p network,
				// eg. he is the first node. if later on he can connect, it will try resolve the fork normally
				bss.BlockTypeStatusService.SetFirstDownloadFinished(bss.ChainType, true)
				bss.Logger.Info(errCasted.Message)
			case blocker.ChainValidationErr:
				bss.Logger.Infof("peer %s:%d: %s",
					peerBlockchainInfo.Peer.GetInfo().Address,
					peerBlockchainInfo.Peer.GetInfo().Port,
					errCasted.Message)
			default:
				bss.Logger.Infof("failed to getPeerBlockchainInfo: %v", err)

			}
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
			bss.BlockTypeStatusService.SetIsDownloading(bss.ChainType, false)
			bss.Logger.Infof("Finished %s blockchain download: %d blocks pulled", bss.ChainType.GetName(), lastBlock.Height-initialHeight)
			break
		}

		if newLastBlock == nil {
			break
		}

		lastBlock = newLastBlock
	}
}
