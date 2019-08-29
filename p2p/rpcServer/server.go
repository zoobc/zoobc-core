package rpcServer

import (
	"context"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	service2 "github.com/zoobc/zoobc-core/p2p/service"
)

// P2PServerHandler represent data service node as server
type P2PServerHandler struct {
	Service service2.P2PServerServiceInterface
}

func NewServerService(
	p2pServerService service2.P2PServerServiceInterface,
) *P2PServerHandler {
	return &P2PServerHandler{
		Service: p2pServerService,
	}
}

// GetPeerInfo to return info of this host
func (ss *P2PServerHandler) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	return ss.Service.GetPeerInfo(req)
}

// GetMorePeers contains info other peers
func (ss *P2PServerHandler) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	var nodes []*model.Node
	nodes, err := ss.Service.GetMorePeers(req)
	if err != nil {
		return nil, err
	}
	return &model.GetMorePeersResponse{
		Peers: nodes,
	}, nil
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (ss *P2PServerHandler) SendPeers(ctx context.Context, req *model.SendPeersRequest) (*model.Empty, error) {
	// TODO: only accept nodes that are already registered in the node registration
	if req.Peers == nil {
		return nil, blocker.NewBlocker(
			blocker.RequestParameterErr,
			"sendPeers node cannot be nil",
		)
	}
	return ss.Service.SendPeers(req.Peers)
}

// GetCumulativeDifficulty responds to the request of the cummulative difficulty status of a node
func (ss *P2PServerHandler) GetCumulativeDifficulty(ctx context.Context,
	req *model.GetCumulativeDifficultyRequest,
) (*model.GetCumulativeDifficultyResponse, error) {
	return ss.Service.GetCumulativeDifficulty(chaintype.GetChainType(req.ChainType))
}

func (ss *P2PServerHandler) GetCommonMilestoneBlockIDs(ctx context.Context,
	req *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	// if `lastBlockID` is supplied
	// check it the last `lastBlockID` got matches with the host's lastBlock then return the response as is
	chainType := chaintype.GetChainType(req.ChainType)
	if req.LastBlockID == 0 && req.LastMilestoneBlockID == 0 {
		return nil, blocker.NewBlocker(
			blocker.RequestParameterErr,
			"either LastBlockID or LastMilestoneBlockID has to be supplied",
		)
	}
	return ss.Service.GetCommonMilestoneBlockIDs(
		chainType, req.LastBlockID, req.LastMilestoneBlockID,
	)
}

func (ss *P2PServerHandler) GetNextBlockIDs(ctx context.Context, req *model.GetNextBlockIdsRequest) (*model.BlockIdsResponse, error) {
	chainType := chaintype.GetChainType(req.ChainType)
	blockIds, err := ss.Service.GetNextBlockIDs(chainType, req.Limit, req.BlockId)
	if err != nil {
		return nil, err
	}
	return &model.BlockIdsResponse{
		BlockIds: blockIds,
	}, nil
}

func (ss *P2PServerHandler) GetNextBlocks(ctx context.Context, req *model.GetNextBlocksRequest) (*model.BlocksData, error) {
	// TODO: getting data from cache
	chainType := chaintype.GetChainType(req.ChainType)
	return ss.Service.GetNextBlocks(
		chainType,
		req.BlockId,
		req.BlockIds,
	)
}

// SendBlock receive block from other node and calling BlockReceived Event
func (ss *P2PServerHandler) SendBlock(ctx context.Context, req *model.SendBlockRequest) (*model.Receipt, error) {
	return ss.Service.SendBlock(chaintype.GetChainType(req.ChainType), req.Block)
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ss *P2PServerHandler) SendTransaction(ctx context.Context, req *model.SendTransactionRequest) (*model.Receipt, error) {
	//ss.Observer.Notify(observer.TransactionReceived, req.GetTransactionBytes(), nil)
	return &model.Receipt{}, nil
}
