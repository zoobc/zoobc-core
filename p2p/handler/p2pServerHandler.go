// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package handler

import (
	"context"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/feedbacksystem"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	service2 "github.com/zoobc/zoobc-core/p2p/service"
)

// P2PServerHandler represent data service node as server
type P2PServerHandler struct {
	Service          service2.P2PServerServiceInterface
	FeedbackStrategy feedbacksystem.FeedbackStrategyInterface
}

func NewP2PServerHandler(
	p2pServerService service2.P2PServerServiceInterface,
	feedbackStrategy feedbacksystem.FeedbackStrategyInterface,
) *P2PServerHandler {
	return &P2PServerHandler{
		Service:          p2pServerService,
		FeedbackStrategy: feedbackStrategy,
	}
}

// GetPeerInfo to return info of this host
func (ss *P2PServerHandler) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

	return ss.Service.GetPeerInfo(ctx, req)
}

// GetNodeAddressesInfo return content of node_address_info table to requesting peer
func (ss *P2PServerHandler) GetNodeAddressesInfo(ctx context.Context, req *model.GetNodeAddressesInfoRequest) (*model.
	GetNodeAddressesInfoResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

	return ss.Service.GetNodeAddressesInfo(ctx, req)
}

// SendNodeAddressInfo receive a NodeAddressInfo sent by a peer
func (ss *P2PServerHandler) SendNodeAddressInfo(ctx context.Context, req *model.SendNodeAddressInfoRequest) (*model.Empty, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetPeerInfoServer)
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

	// service method (P2PServerServiceInterface) to send a node address info message to a peer
	return ss.Service.SendNodeAddressInfo(ctx, req)
}

// GetMorePeers contains info other peers
func (ss *P2PServerHandler) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetMorePeersServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetMorePeersServer)
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

	return ss.Service.GetCumulativeDifficulty(ctx, chaintype.GetChainType(req.ChainType))
}

func (ss *P2PServerHandler) GetCommonMilestoneBlockIDs(ctx context.Context,
	req *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	monitoring.IncrementGoRoutineActivity(monitoring.P2pGetCommonMilestoneBlockIDsServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pGetCommonMilestoneBlockIDsServer)
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	monitoring.IncrementGoRoutineActivity(monitoring.P2pSendTransactionServer)
	defer monitoring.DecrementGoRoutineActivity(monitoring.P2pSendTransactionServer)
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

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
	ss.FeedbackStrategy.IncrementVarCount("P2PIncomingRequests")
	defer ss.FeedbackStrategy.DecrementVarCount("P2PIncomingRequests")

	res, err := ss.Service.GetNodeProofOfOrigin(ctx, req)
	return res, err
}
