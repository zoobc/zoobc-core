package service

import "github.com/zoobc/zoobc-core/common/chaintype"

type (
	BlockTypeStatusServiceInterface interface {
		SetFirstDownloadFinished(ct chaintype.ChainType, isSpineBlocksDownloadFinished bool)
		IsFirstDownloadFinished(ct chaintype.ChainType) bool
		SetIsDownloading(ct chaintype.ChainType, newValue bool)
		IsDownloading(ct chaintype.ChainType) bool
		SetIsSmithingLocked(isSmithingLocked bool)
		IsSmithingLocked() bool
		SetIsSmithing(ct chaintype.ChainType, smithing bool)
		IsSmithing(ct chaintype.ChainType) bool
	}
)

type (
	BlockTypeStatusService struct {
		isFirstDownloadFinished map[int32]bool
		isDownloading           map[int32]bool
		isSmithing              map[int32]bool
		isSmithingLocked        bool
	}
)

func NewBlockTypeStatusService(
	lockSmithing bool,
) *BlockTypeStatusService {
	// init variables for all block types
	var btss = &BlockTypeStatusService{
		isDownloading:           make(map[int32]bool),
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

func (btss *BlockTypeStatusService) SetFirstDownloadFinished(ct chaintype.ChainType, finished bool) {
	btss.isFirstDownloadFinished[ct.GetTypeInt()] = finished
}

func (btss *BlockTypeStatusService) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return btss.isFirstDownloadFinished[ct.GetTypeInt()]
}

func (btss *BlockTypeStatusService) SetIsDownloading(ct chaintype.ChainType, newValue bool) {
	btss.isDownloading[ct.GetTypeInt()] = newValue
}

func (btss *BlockTypeStatusService) IsDownloading(ct chaintype.ChainType) bool {
	return btss.isDownloading[ct.GetTypeInt()]
}

func (btss *BlockTypeStatusService) SetIsSmithingLocked(isSmithingLocked bool) {
	btss.isSmithingLocked = isSmithingLocked
}

func (btss *BlockTypeStatusService) IsSmithingLocked() bool {
	return btss.isSmithingLocked
}

func (btss *BlockTypeStatusService) SetIsSmithing(ct chaintype.ChainType, isSmithing bool) {
	btss.isSmithing[ct.GetTypeInt()] = isSmithing
}

func (btss *BlockTypeStatusService) IsSmithing(ct chaintype.ChainType) bool {
	return btss.isSmithing[ct.GetTypeInt()]
}
