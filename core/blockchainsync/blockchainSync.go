package blockchainsync

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
)

type Service struct {
	NeedGetMoreBlocks          bool
	IsDownloading              bool // only for status
	LastBlockchainFeeder       *model.Peer
	LastBlockchainFeederHeight uint32

	PeerHasMore bool

	ChainType          chaintype.ChainType
	BlockService       service.BlockServiceInterface
	P2pService         p2p.ServiceInterface
	LastBlock          model.Block
	TransactionService service.TransactionServiceInterface
	TransactionQuery   query.TransactionQueryInterface
	ForkingProcess     ForkingProcess
}

func NewBlockchainSyncService(blockService service.BlockServiceInterface, p2pService p2p.ServiceInterface) *Service {
	return &Service{
		NeedGetMoreBlocks: true,
		ChainType:         blockService.GetChainType(),
		BlockService:      blockService,
		P2pService:        p2pService,
	}
}
