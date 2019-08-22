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

<<<<<<< HEAD:core/blockchainsync/blockchainSync.go
	ChainType    chaintype.ChainType
	BlockService service.BlockServiceInterface
	P2pService   p2p.ServiceInterface
=======
	ChainType          contract.ChainType
	BlockService       service.BlockServiceInterface
	P2pService         p2p.P2pServiceInterface
	LastBlock          model.Block
	TransactionService service.TransactionServiceInterface
	TransactionQuery   query.TransactionQueryInterface
	ForkingProcess     ForkingProcess
>>>>>>> 8a8946e... applying rollback from spinechain template:core/blockchainSync/blockchainSync.go
}

func NewBlockchainSyncService(blockService service.BlockServiceInterface, p2pService p2p.ServiceInterface) *Service {
	return &Service{
		NeedGetMoreBlocks: true,
		ChainType:         blockService.GetChainType(),
		BlockService:      blockService,
		P2pService:        p2pService,
	}
}
