package blockchainSync

import (
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type BlockchainSyncService struct {
	NeedGetMoreBlocks          bool
	IsDownloading              bool // only for status
	LastBlockchainFeeder       *model.Peer
	LastBlockchainFeederHeight uint32

	PeerHasMore bool

	ChainType    contract.ChainType
	BlockService service.BlockServiceInterface
	P2pService   p2p.P2pServiceInterface
}

func NewBlockchainSyncService(blockService service.BlockServiceInterface, p2pService p2p.P2pServiceInterface) *BlockchainSyncService {
	return &BlockchainSyncService{
		NeedGetMoreBlocks: true,
		ChainType:         blockService.GetChainType(),
		BlockService:      blockService,
		P2pService:        p2pService,
	}
}
