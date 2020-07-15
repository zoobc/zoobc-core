package handler

import (
	"context"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	service2 "github.com/zoobc/zoobc-core/p2p/service"
)

// P2PServerHandler represent data service node as server
type P2PServerHandler struct {
	Service service2.P2PServerServiceInterface
}

func NewP2PServerHandler(
	p2pServerService service2.P2PServerServiceInterface,
) *P2PServerHandler {
	return &P2PServerHandler{
		Service: p2pServerService,
	}
}

// GetPeerInfo to return info of this host
func (ss *P2PServerHandler) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)

	return ss.Service.GetPeerInfo(ctx, req)
}

// GetNodeAddressesInfo return content of node_address_info table to requesting peer
func (ss *P2PServerHandler) GetNodeAddressesInfo(ctx context.Context, req *model.GetNodeAddressesInfoRequest) (*model.
	GetNodeAddressesInfoResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)

	return ss.Service.GetNodeAddressesInfo(ctx, req)
}

// SendNodeAddressInfo receive a NodeAddressInfo sent by a peer
func (ss *P2PServerHandler) SendNodeAddressInfo(ctx context.Context, req *model.SendNodeAddressInfoRequest) (*model.Empty, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)

	// service method (P2PServerServiceInterface) to send a node address info message to a peer
	return ss.Service.SendNodeAddressInfo(ctx, req)
}

// GetMorePeers contains info other peers
func (ss *P2PServerHandler) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetMorePeersServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetMorePeersServer)

	var nodes []*model.Node
	nodes, err := ss.Service.GetMorePeers(ctx, req)
	if err != nil {
		return nil, err
	}
	return &model.GetMorePeersResponse{
		Peers: nodes,
	}, nil
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (ss *P2PServerHandler) SendPeers(ctx context.Context, req *model.SendPeersRequest) (*model.Empty, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendPeersServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendPeersServer)

	// TODO: only accept nodes that are already registered in the node registration
	if req.Peers == nil {
		return nil, blocker.NewBlocker(
			blocker.RequestParameterErr,
			"sendPeers node cannot be nil",
		)
	}
	return ss.Service.SendPeers(ctx, req.Peers)
}

// GetCumulativeDifficulty responds to the request of the cumulative difficulty status of a node
func (ss *P2PServerHandler) GetCumulativeDifficulty(ctx context.Context,
	req *model.GetCumulativeDifficultyRequest,
) (*model.GetCumulativeDifficultyResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetCumulativeDifficultyServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetCumulativeDifficultyServer)

	return ss.Service.GetCumulativeDifficulty(ctx, chaintype.GetChainType(req.ChainType))
}

func (ss *P2PServerHandler) GetCommonMilestoneBlockIDs(ctx context.Context,
	req *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetCommonMilestoneBlockIDsServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetCommonMilestoneBlockIDsServer)

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
		ctx, chainType, req.LastBlockID, req.LastMilestoneBlockID,
	)
}

func (ss *P2PServerHandler) GetNextBlockIDs(ctx context.Context, req *model.GetNextBlockIdsRequest) (*model.BlockIdsResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetNextBlockIDsServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetNextBlockIDsServer)

	chainType := chaintype.GetChainType(req.ChainType)
	blockIds, err := ss.Service.GetNextBlockIDs(ctx, chainType, req.Limit, req.BlockId)
	if err != nil {
		return nil, err
	}
	return &model.BlockIdsResponse{
		BlockIds: blockIds,
	}, nil
}

func (ss *P2PServerHandler) GetNextBlocks(ctx context.Context, req *model.GetNextBlocksRequest) (*model.BlocksData, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetNextBlocksServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetNextBlocksServer)

	// TODO: getting data from cache
	chainType := chaintype.GetChainType(req.ChainType)
	return ss.Service.GetNextBlocks(
		ctx,
		chainType,
		req.BlockId,
		req.BlockIds,
	)
}

// SendBlock receive block from other node and calling BlockReceived Event
func (ss *P2PServerHandler) SendBlock(ctx context.Context, req *model.SendBlockRequest) (*model.SendBlockResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendBlockServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendBlockServer)

	// todo: validate request
	return ss.Service.SendBlock(
		ctx,
		chaintype.GetChainType(req.ChainType),
		req.Block,
		req.SenderPublicKey,
	)
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ss *P2PServerHandler) SendTransaction(
	ctx context.Context,
	req *model.SendTransactionRequest,
) (*model.SendTransactionResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendTransactionServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendTransactionServer)

	return ss.Service.SendTransaction(
		ctx,
		chaintype.GetChainType(req.ChainType),
		req.TransactionBytes,
		req.SenderPublicKey,
	)
}

// SendBlockTransactions receive transaction from other node and calling TransactionReceived Event
func (ss *P2PServerHandler) SendBlockTransactions(
	ctx context.Context,
	req *model.SendBlockTransactionsRequest,
) (*model.SendBlockTransactionsResponse, error) {
	return ss.Service.SendBlockTransactions(
		ctx,
		chaintype.GetChainType(req.ChainType),
		req.TransactionsBytes,
		req.SenderPublicKey,
	)
}

// RequestBlockTransactions receive requested transaction from another node
func (ss *P2PServerHandler) RequestBlockTransactions(
	ctx context.Context,
	req *model.RequestBlockTransactionsRequest,
) (*model.Empty, error) {
	return ss.Service.RequestBlockTransactions(
		ctx,
		chaintype.GetChainType(req.ChainType),
		req.GetBlockID(),
		req.GetTransactionIDs(),
	)
}

// RequestFileDownload receives an array of file names and return corresponding files.
func (ss *P2PServerHandler) RequestFileDownload(
	ctx context.Context,
	req *model.FileDownloadRequest,
) (*model.FileDownloadResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pRequestFileDownloadServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pRequestFileDownloadServer)
	if len(req.FileChunkNames) == 0 {
		return nil, blocker.NewBlocker(
			blocker.RequestParameterErr,
			"request does not contain any file name",
		)
	}
	res, err := ss.Service.RequestDownloadFile(ctx, req.GetSnapshotHash(), req.GetFileChunkNames())
	if res != nil {
		monitoring.IncrementSnapshotDownloadCounter(int32(len(res.FileChunks)), int32(len(res.Failed)))
	}
	return res, err
}

func (ss *P2PServerHandler) GetNodeProofOfOrigin(ctx context.Context, req *model.GetNodeProofOfOriginRequest) (*model.ProofOfOrigin, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetNodeProofOfOriginServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetNodeProofOfOriginServer)
	res, err := ss.Service.GetNodeProofOfOrigin(ctx, req)
	return res, err
}
