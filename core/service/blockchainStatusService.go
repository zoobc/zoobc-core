package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"sync"
)

type (
	BlockchainStatusServiceInterface interface {
		SetFirstDownloadFinished(ct chaintype.ChainType, isSpineBlocksDownloadFinished bool)
		IsFirstDownloadFinished(ct chaintype.ChainType) bool
		SetIsDownloading(ct chaintype.ChainType, newValue bool)
		IsDownloading(ct chaintype.ChainType) bool
		SetIsSmithingLocked(isSmithingLocked bool)
		IsSmithingLocked() bool
		SetIsSmithing(ct chaintype.ChainType, smithing bool)
		IsSmithing(ct chaintype.ChainType) bool
		SetIsDownloadingSnapshot(ct chaintype.ChainType, isDownloadingSnapshot bool)
		IsDownloadingSnapshot(ct chaintype.ChainType) bool
		SetIsBlocksmith(isBlocksmith bool)
		IsBlocksmith() bool
		SetLastBlock(block *model.Block, ct chaintype.ChainType)
		GetLastBlock(ct chaintype.ChainType) *model.Block
	}
)

type (
	BlockchainStatusService struct {
		Logger *log.Logger
	}
)

var (
	isFirstDownloadFinished = model.NewMapIntBool()
	isDownloading           = model.NewMapIntBool()
	isDownloadingSnapshot   = model.NewMapIntBool()
	isSmithing              = model.NewMapIntBool()
	lastBlock               = make(map[int32]*model.Block)
	isSmithingLocked        bool
	isBlocksmith            bool
	lastBlockMux            sync.RWMutex
)

func NewBlockchainStatusService(
	lockSmithing bool,
	logger *log.Logger,
) *BlockchainStatusService {
	// init variables for all block types
	var btss = &BlockchainStatusService{
		Logger: logger,
	}
	btss.SetIsSmithingLocked(lockSmithing)
	return btss
}

// SetLastBlock set 'cached' last block (updated every time a block is pushed)
func (btss *BlockchainStatusService) SetLastBlock(block *model.Block, ct chaintype.ChainType) {
	lastBlockMux.Lock()
	lastBlock[ct.GetTypeInt()] = block
	lastBlockMux.Unlock()
}

// GetLastBlock get 'cached' last block (updated every time a block is pushed)
func (btss *BlockchainStatusService) GetLastBlock(ct chaintype.ChainType) *model.Block {
	lastBlockMux.Lock()
	defer lastBlockMux.Unlock()
	if bl, ok := lastBlock[ct.GetTypeInt()]; ok {
		return bl
	}
	return nil
}

func (btss *BlockchainStatusService) SetFirstDownloadFinished(ct chaintype.ChainType, finished bool) {
	// set it only once, when the node starts
	if res, ok := isFirstDownloadFinished.Load(ct.GetTypeInt()); ok && res {
		return
	}
	isFirstDownloadFinished.Store(ct.GetTypeInt(), finished)
	if finished {
		btss.Logger.Infof("%s first download finished", ct.GetName())
	}
}

func (btss *BlockchainStatusService) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	if res, ok := isFirstDownloadFinished.Load(ct.GetTypeInt()); ok {
		return res
	}
	return false
}

func (btss *BlockchainStatusService) SetIsDownloading(ct chaintype.ChainType, downloading bool) {
	isDownloading.Store(ct.GetTypeInt(), downloading)
}

func (btss *BlockchainStatusService) IsDownloading(ct chaintype.ChainType) bool {
	if res, ok := isDownloading.Load(ct.GetTypeInt()); ok {
		return res
	}
	return false
}

func (btss *BlockchainStatusService) SetIsSmithingLocked(smithingLocked bool) {
	var (
		lockedStr string
	)
	isSmithingLocked = smithingLocked
	if isSmithingLocked {
		lockedStr = "locked"
	} else {
		lockedStr = "unlocked"
	}
	btss.Logger.Infof("smithing process %s...", lockedStr)
}

func (btss *BlockchainStatusService) IsSmithingLocked() bool {
	return isSmithingLocked
}

func (btss *BlockchainStatusService) SetIsSmithing(ct chaintype.ChainType, smithing bool) {
	isSmithing.Store(ct.GetTypeInt(), smithing)
}

func (btss *BlockchainStatusService) IsSmithing(ct chaintype.ChainType) bool {
	if res, ok := isSmithing.Load(ct.GetTypeInt()); ok {
		return res
	}
	return false
}

func (btss *BlockchainStatusService) SetIsDownloadingSnapshot(ct chaintype.ChainType, downloadingSnapshot bool) {
	isDownloadingSnapshot.Store(ct.GetTypeInt(), downloadingSnapshot)
	if downloadingSnapshot {
		btss.Logger.Infof("Downloading snapshot for %s...", ct.GetName())
	} else {
		btss.Logger.Infof("Finished Downloading snapshot for %s...", ct.GetName())
	}

}

func (btss *BlockchainStatusService) IsDownloadingSnapshot(ct chaintype.ChainType) bool {
	if !ct.HasSnapshots() {
		return false
	}
	if res, ok := isDownloadingSnapshot.Load(ct.GetTypeInt()); ok {
		return res
	}
	return false
}

func (btss *BlockchainStatusService) SetIsBlocksmith(blocksmith bool) {
	isBlocksmith = blocksmith
}

func (btss *BlockchainStatusService) IsBlocksmith() bool {
	return isBlocksmith
}
