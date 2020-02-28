package service

import "github.com/zoobc/zoobc-core/common/chaintype"

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
	}
)

type (
	BlockchainStatusService struct {
		isFirstDownloadFinished map[int32]bool
		isDownloading           map[int32]bool
		isDownloadingSnapshot   map[int32]bool
		isSmithing              map[int32]bool
		isSmithingLocked        bool
	}
)

func NewBlockchainStatusService(
	lockSmithing bool,
) *BlockchainStatusService {
	// init variables for all block types
	var btss = &BlockchainStatusService{
		isDownloading:           make(map[int32]bool),
		isDownloadingSnapshot:   make(map[int32]bool),
		isFirstDownloadFinished: make(map[int32]bool),
		isSmithing:              make(map[int32]bool),
	}
	for _, ct := range chaintype.GetChainTypes() {
		btss.isDownloading[ct.GetTypeInt()] = false
		btss.isFirstDownloadFinished[ct.GetTypeInt()] = false
	}
	btss.isSmithingLocked = lockSmithing
	return btss
}

func (btss *BlockchainStatusService) SetFirstDownloadFinished(ct chaintype.ChainType, finished bool) {
	btss.isFirstDownloadFinished[ct.GetTypeInt()] = finished
}

func (btss *BlockchainStatusService) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return btss.isFirstDownloadFinished[ct.GetTypeInt()]
}

func (btss *BlockchainStatusService) SetIsDownloading(ct chaintype.ChainType, newValue bool) {
	btss.isDownloading[ct.GetTypeInt()] = newValue
}

func (btss *BlockchainStatusService) IsDownloading(ct chaintype.ChainType) bool {
	return btss.isDownloading[ct.GetTypeInt()]
}

func (btss *BlockchainStatusService) SetIsSmithingLocked(isSmithingLocked bool) {
	btss.isSmithingLocked = isSmithingLocked
}

func (btss *BlockchainStatusService) IsSmithingLocked() bool {
	return btss.isSmithingLocked
}

func (btss *BlockchainStatusService) SetIsSmithing(ct chaintype.ChainType, isSmithing bool) {
	btss.isSmithing[ct.GetTypeInt()] = isSmithing
}

func (btss *BlockchainStatusService) IsSmithing(ct chaintype.ChainType) bool {
	return btss.isSmithing[ct.GetTypeInt()]
}

func (btss *BlockchainStatusService) SetIsDownloadingSnapshot(ct chaintype.ChainType, isDownloadingSnapshot bool) {
	btss.isDownloadingSnapshot[ct.GetTypeInt()] = isDownloadingSnapshot
}

func (btss *BlockchainStatusService) IsDownloadingSnapshot(ct chaintype.ChainType) bool {
	if !ct.HasSnapshots() {
		return false
	}
	return btss.isDownloadingSnapshot[ct.GetTypeInt()]
}
