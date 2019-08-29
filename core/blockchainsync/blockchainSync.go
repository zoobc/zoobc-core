package blockchainsync

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
	"github.com/zoobc/zoobc-core/p2p/rpcClient"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

type DownloadBlockchainService struct {
	NeedGetMoreBlocks          bool
	IsDownloading              bool // only for status
	LastBlockchainFeeder       *model.Peer
	LastBlockchainFeederHeight uint32

	PeerHasMore bool

	ChainType         chaintype.ChainType
	BlockService      service.BlockServiceInterface
	P2pService        p2p.Peer2PeerServiceInterface
	PeerServiceClient rpcClient.PeerServiceClientInterface
	PeerExplorer      strategy.PeerExplorerStrategyInterface
}

func NewBlockchainSyncService(
	blockService service.BlockServiceInterface,
	p2pService p2p.Peer2PeerServiceInterface,
	peerServiceClient rpcClient.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
) *DownloadBlockchainService {
	return &DownloadBlockchainService{
		NeedGetMoreBlocks: true,
		ChainType:         blockService.GetChainType(),
		BlockService:      blockService,
		P2pService:        p2pService,
		PeerServiceClient: peerServiceClient,
		PeerExplorer:      peerExplorer,
	}
}
