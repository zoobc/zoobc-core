package blockchainsync

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type Service struct {
	// isScanningBlockchain       bool
	ChainType chaintype.ChainType

	BlockService         service.BlockServiceInterface
	BlockchainDownloader BlockchainDownloadInterface
	ForkingProcessor     ForkingProcessorInterface
}

func NewBlockchainSyncService(blockService service.BlockServiceInterface, p2pService p2p.ServiceInterface, queryExecutor query.ExecutorInterface) *Service {
	return &Service{
		ChainType:    blockService.GetChainType(),
		BlockService: blockService,
		BlockchainDownloader: &BlockchainDownloader{
			ChainType:    blockService.GetChainType(),
			BlockService: blockService,
			P2pService:   p2pService,
		},
		ForkingProcessor: &ForkingProcessor{
			ChainType:    blockService.GetChainType(),
			BlockService: blockService,
			BlockPopper: &BlockPopper{
				ChainType:     blockService.GetChainType(),
				BlockService:  blockService,
				QueryExecutor: queryExecutor,
			},
		},
	}
}

func (bss *Service) Start(runNext chan bool) {
	// // TODO: restore this checking
	// if bss.ChainType == nil {
	// 	log.Fatal("no chaintype")
	// }
	// if bss.P2pService == nil {
	// 	log.Fatal("no p2p service defined")
	// }
	bss.GetMoreBlocksThread(runNext)
}

func (bss *Service) GetMoreBlocksThread(runNext chan bool) {
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
	log.Info("Get more blocks...")
	// notify observer about start of blockchain download of this specific chain

	lastBlock, blockErr := bss.BlockService.GetLastBlock()
	if blockErr != nil {
		log.Warn(fmt.Sprintf("failed to start getMoreBlocks go routine: %v", blockErr))
	}
	if lastBlock == nil {
		log.Warn("There is no genesis block found")
	}
	initialHeight := lastBlock.Height

	for {
		bss.BlockService.ChainWriteLock()
		defer bss.BlockService.ChainWriteUnlock()
		peerBlockchainInfo, err := bss.BlockchainDownloader.GetPeerBlockchainInfo()
		if err != nil {
			log.Warnf("\nfailed to getPeerBlockchainInfo: %v\n\n", err)
			break
		}

		peerForkInfo, errDownload := bss.BlockchainDownloader.DownloadFromPeer(peerBlockchainInfo.Peer, peerBlockchainInfo.ChainBlockIds, peerBlockchainInfo.CommonBlock)
		if errDownload != nil {
			log.Warnf("\nfailed to DownloadFromPeer: %v\n\n", err)
			break
		}

		if len(peerForkInfo.ForkBlocks) > 0 {
			err := bss.ForkingProcessor.ProcessFork(peerForkInfo.ForkBlocks, peerBlockchainInfo.CommonBlock, peerForkInfo.FeederPeer)
			if err != nil {
				log.Warnf("\nfailed to ProcessFork: %v\n\n", err)
				break
			}
		}

		// TODO: fix ConfirmBlockchainState
		// confirmBlockchainError := bss.BlockchainDownloader.ConfirmBlockchainState(peerBlockchainInfo.Peer, peerBlockchainInfo.CommonMilestoneBlockID, peerBlockchainInfo.CommonBlock)
		// if confirmBlockchainError != nil {
		// 	log.Warnf("\nfailed to ConfirmBlockchainState: %v\n\n", err)
		// }
		newLastBlock, err := bss.BlockService.GetLastBlock()
		if err != nil {
			log.Warnf("\nfailed to getMoreBlocks: %v\n\n", err)
			break
		}

		if lastBlock.ID == newLastBlock.ID {
			log.Println("Did not accept peers's blocks, back to our own fork")
			break
		}

		if bss.BlockchainDownloader.IsDownloadFinish(lastBlock) {
			bss.BlockchainDownloader.SetIsDownloading(false)
			log.Infof("Finished %s blockchain download: %d blocks pulled", bss.ChainType.GetName(), lastBlock.Height-initialHeight)
			break
		}
		lastBlock = newLastBlock
	}

	// TODO: Handle interruption and other exceptions
	time.Sleep(constant.GetMoreBlocksDelay * time.Second)
	runNext <- true
}
