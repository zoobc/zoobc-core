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
package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/zoobc/zoobc-core/common/feedbacksystem"
	"github.com/zoobc/zoobc-core/common/monitoring"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	// P2PServerServiceInterface interface that contains registered methods of P2P Server Service
	P2PServerServiceInterface interface {
		GetNodeAddressesInfo(ctx context.Context, req *model.GetNodeAddressesInfoRequest) (*model.GetNodeAddressesInfoResponse, error)
		SendNodeAddressInfo(ctx context.Context, req *model.SendNodeAddressInfoRequest) (*model.Empty, error)
		GetNodeProofOfOrigin(ctx context.Context, req *model.GetNodeProofOfOriginRequest) (*model.ProofOfOrigin, error)
		GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error)
		GetMorePeers(ctx context.Context, req *model.Empty) ([]*model.Node, error)
		SendPeers(ctx context.Context, peers []*model.Node) (*model.Empty, error)
		GetCumulativeDifficulty(
			ctx context.Context,
			chainType chaintype.ChainType,
		) (*model.GetCumulativeDifficultyResponse, error)
		GetCommonMilestoneBlockIDs(
			ctx context.Context,
			chainType chaintype.ChainType,
			lastBlockID,
			lastMilestoneBlockID int64,
		) (*model.GetCommonMilestoneBlockIdsResponse, error)
		GetNextBlockIDs(
			ctx context.Context,
			chainType chaintype.ChainType,
			reqLimit uint32,
			reqBlockID int64,
		) ([]int64, error)
		GetNextBlocks(
			ctx context.Context,
			chainType chaintype.ChainType,
			blockID int64,
			blockIDList []int64,
		) (*model.BlocksData, error)
		SendBlock(
			ctx context.Context,
			chainType chaintype.ChainType,
			block *model.Block,
			senderPublicKey []byte,
		) (*model.SendBlockResponse, error)
		SendTransaction(
			ctx context.Context,
			chainType chaintype.ChainType,
			transactionBytes,
			senderPublicKey []byte,
		) (*model.SendTransactionResponse, error)
		SendBlockTransactions(
			ctx context.Context,
			chainType chaintype.ChainType,
			transactionsBytes [][]byte,
			senderPublicKey []byte,
		) (*model.SendBlockTransactionsResponse, error)
		RequestBlockTransactions(
			ctx context.Context,
			chainType chaintype.ChainType,
			blockID int64,
			transactionsIDs []int64,
		) (*model.Empty, error)
		RequestDownloadFile(ctx context.Context, snapshotHash []byte, fileChunkNames []string) (*model.FileDownloadResponse, error)
	}
	// P2PServerService represent of P2P server service
	P2PServerService struct {
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		FileService              coreService.FileServiceInterface
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		NodeAddressInfoService   coreService.NodeAddressInfoServiceInterface
		PeerExplorer             strategy.PeerExplorerStrategyInterface
		BlockServices            map[int32]coreService.BlockServiceInterface
		MempoolServices          map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase         string
		Observer                 *observer.Observer
		FeedbackStrategy         feedbacksystem.FeedbackStrategyInterface
	}
)

// NewP2PServerService return new instance of P2P server service
func NewP2PServerService(
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	fileService coreService.FileServiceInterface,
	nodeConfigurationService coreService.NodeConfigurationServiceInterface,
	nodeAddressInfoServiceInterface coreService.NodeAddressInfoServiceInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	mempoolServices map[int32]coreService.MempoolServiceInterface,
	nodeSecretPhrase string,
	observer *observer.Observer,
	feedbackStrategy feedbacksystem.FeedbackStrategyInterface,
) *P2PServerService {
	return &P2PServerService{
		NodeRegistrationService:  nodeRegistrationService,
		FileService:              fileService,
		NodeConfigurationService: nodeConfigurationService,
		NodeAddressInfoService:   nodeAddressInfoServiceInterface,
		PeerExplorer:             peerExplorer,
		BlockServices:            blockServices,
		MempoolServices:          mempoolServices,
		NodeSecretPhrase:         nodeSecretPhrase,
		Observer:                 observer,
		FeedbackStrategy:         feedbackStrategy,
	}
}

// GetNodeAddressesInfo responds to the request of peers a (pending) node address info
// note: since we can return one address per nodeID and node addresses can have more than one state,
// 'confirmed' addresses will be preferred over 'pending' when a node has both versions, when retrieving addresses from a peer
func (ps *P2PServerService) GetNodeAddressesInfo(
	ctx context.Context,
	req *model.GetNodeAddressesInfoRequest,
) (*model.GetNodeAddressesInfoResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		if nodeAddressesInfo, err := ps.NodeAddressInfoService.GetAddressInfoTableWithConsolidatedAddresses(
			model.NodeAddressStatus_NodeAddressConfirmed,
		); err == nil {
			return &model.GetNodeAddressesInfoResponse{
				NodeAddressesInfo: nodeAddressesInfo,
			}, nil
		}
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// SendNodeAddressesInfo receives a node address info from a peer
func (ps *P2PServerService) SendNodeAddressInfo(ctx context.Context, req *model.SendNodeAddressInfoRequest) (*model.Empty, error) {
	var (
		nodeAddressInfo = req.NodeAddressInfoMessage
	)
	if ps.PeerExplorer.ValidateRequest(ctx) {
		// if node receives own address don't do anything
		var nodeAddressInfosToReceive = make([]*model.NodeAddressInfo, 0)
		myAddress, errAddr := ps.NodeConfigurationService.GetMyAddress()
		myPort, errPort := ps.NodeConfigurationService.GetMyPeerPort()
		for _, info := range nodeAddressInfo {
			if errAddr == nil && errPort == nil && myAddress == info.GetAddress() && myPort == info.GetPort() {
				return &model.Empty{}, nil
			}
			// validate node address info message and signature
			if alreadyUpdated, err := ps.NodeAddressInfoService.ValidateNodeAddressInfo(info); err != nil {
				// TODO: blacklist peers that send invalid data (unless failed validation is because this node doesn't exist in nodeRegistry,
				//  or address is already in db or peer sent an old, but valid addressinfo)
				// if validation failed because we already have this address in db, don't return errors (that behavior could be exploited)
				// errorMsg := err.Error()
				// errCasted, ok := err.(blocker.Blocker)
				// if ok {
				// 	errorMsg = errCasted.Message
				// }
				// if errorMsg != "NodeIDNotFound" && errorMsg != "AddressAlreadyUpdatedForNode" && errorMsg != "OutdatedNodeAddressInfo" {
				// 	// blacklist peer!
				// }
			} else if !alreadyUpdated {
				nodeAddressInfosToReceive = append(nodeAddressInfosToReceive, info)
			}
		}
		if err := ps.PeerExplorer.ReceiveNodeAddressInfo(nodeAddressInfosToReceive); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &model.Empty{}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// GetNodeProofOfOrigin generate a proof of origin to be returned to the peer that requested it
func (ps *P2PServerService) GetNodeProofOfOrigin(ctx context.Context, req *model.GetNodeProofOfOriginRequest) (*model.ProofOfOrigin, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		return ps.PeerExplorer.GenerateProofOfOrigin(req.ChallengeMessage, req.Timestamp, ps.NodeSecretPhrase), nil

	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// GetPeerInfo responds to the request of peers a node info
func (ps *P2PServerService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.GetPeerInfoResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		return &model.GetPeerInfoResponse{
			HostInfo: ps.PeerExplorer.GetHostInfo(),
		}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// GetMorePeers contains info other peers
func (ps *P2PServerService) GetMorePeers(ctx context.Context, req *model.Empty) ([]*model.Node, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var nodes []*model.Node
		// only sends the connected (resolved) peers
		for _, hostPeer := range ps.PeerExplorer.GetResolvedPeers() {
			nodes = append(nodes, hostPeer.GetInfo())
		}
		return nodes, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (ps *P2PServerService) SendPeers(
	ctx context.Context,
	peers []*model.Node,
) (*model.Empty, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		// TODO: only accept nodes that are already registered in the node registration
		var compatiblePeers []*model.Node
		for _, peer := range peers {
			if err := p2pUtil.CheckPeerCompatibility(ps.PeerExplorer.GetHostInfo(), peer); err == nil {
				compatiblePeers = append(compatiblePeers, peer)
			}
		}
		err := ps.PeerExplorer.AddToUnresolvedPeers(compatiblePeers, true)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &model.Empty{}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// GetCumulativeDifficulty responds to the request of the cumulative difficulty status of a node
func (ps *P2PServerService) GetCumulativeDifficulty(
	ctx context.Context,
	chainType chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		blockService := ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		lastBlock, err := blockService.GetLastBlock()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &model.GetCumulativeDifficultyResponse{
			CumulativeDifficulty: lastBlock.CumulativeDifficulty,
			Height:               lastBlock.Height,
		}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

func (ps P2PServerService) GetCommonMilestoneBlockIDs(
	ctx context.Context,
	chainType chaintype.ChainType,
	lastBlockID,
	lastMilestoneBlockID int64,
) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		// if `lastBlockID` is supplied
		// check it the last `lastBlockID` got matches with the host's lastBlock then return the response as is
		var (
			height, jump uint32
			blockIds     []int64
			blockService = ps.BlockServices[chainType.GetTypeInt()]
		)
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		if lastBlockID == 0 && lastMilestoneBlockID == 0 {
			return nil, status.Error(codes.InvalidArgument, "either LastBlockID or LastMilestoneBlockID has to be supplied")
		}
		myLastBlock, err := blockService.GetLastBlockCacheFormat()
		if err != nil || myLastBlock == nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failGetLastBlockErr: %v", err.Error()))
		}
		myLastBlockID := myLastBlock.ID
		myBlockchainHeight := myLastBlock.Height

		if _, err := blockService.GetBlockByID(lastBlockID, false); err == nil {
			preparedResponse := &model.GetCommonMilestoneBlockIdsResponse{
				BlockIds: []int64{lastBlockID},
			}
			if lastBlockID == myLastBlockID {
				preparedResponse.Last = true
			}
			return preparedResponse, nil
		}

		// if not, send (assumed) milestoneBlock of the host
		limit := constant.CommonMilestoneBlockIdsLimit
		if lastMilestoneBlockID != 0 {
			lastMilestoneBlock, err := blockService.GetBlockByID(lastMilestoneBlockID, false)
			// this error is handled because when lastMilestoneBlockID is provided, it was expected to be the one returned from this node
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			height = lastMilestoneBlock.GetHeight()
			jump = util.MinUint32(constant.SafeBlockGap, util.MaxUint32(myBlockchainHeight, 1))
		} else if lastBlockID != 0 {
			// TODO: analyze difference of height jump
			height = myBlockchainHeight
			jump = 10
		}

	LoopBlocks:
		for ; limit > 0; limit-- {
			block, err := blockService.GetBlockByHeightCacheFormat(height)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			blockIds = append(blockIds, block.ID)
			switch {
			case height == 0:
				break LoopBlocks
			case height < jump:
				height = 0
			default:
				height -= jump
			}
		}

		return &model.GetCommonMilestoneBlockIdsResponse{BlockIds: blockIds}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

func (ps *P2PServerService) GetNextBlockIDs(
	ctx context.Context,
	chainType chaintype.ChainType,
	reqLimit uint32,
	reqBlockID int64,
) ([]int64, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var (
			blockIds     []int64
			blockService = ps.BlockServices[chainType.GetTypeInt()]
		)
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		limit := constant.PeerGetBlocksLimit
		if reqLimit != 0 && reqLimit < limit {
			limit = reqLimit
		}

		foundBlock, err := blockService.GetBlockByID(reqBlockID, false)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		blocks, err := blockService.GetBlocksFromHeight(foundBlock.Height, limit, false)
		if err != nil {
			return nil, status.Error(codes.Internal, "failedGetBlocks")
		}
		for _, block := range blocks {
			blockIds = append(blockIds, block.ID)
		}

		return blockIds, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

func (ps *P2PServerService) GetNextBlocks(
	ctx context.Context,
	chainType chaintype.ChainType,
	blockID int64,
	blockIDList []int64,
) (*model.BlocksData, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		// TODO: getting data from cache
		var (
			blocksMessage []*model.Block
			blockService  = ps.BlockServices[chainType.GetTypeInt()]
		)
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		commonBlock, err := blockService.GetBlockByID(blockID, false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		blocks, err := blockService.GetBlocksFromHeight(commonBlock.Height, uint32(len(blockIDList)), true)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		for idx, block := range blocks {
			if block.ID != blockIDList[idx] {
				break
			}
			blocksMessage = append(blocksMessage, block)
		}
		return &model.BlocksData{NextBlocks: blocksMessage}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// SendBlock receive block from other node
func (ps *P2PServerService) SendBlock(
	ctx context.Context,
	chainType chaintype.ChainType,
	block *model.Block,
	senderPublicKey []byte,
) (*model.SendBlockResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var md, _ = metadata.FromIncomingContext(ctx)
		if len(md) == 0 {
			return nil, status.Error(
				codes.InvalidArgument,
				"InvalidContext",
			)
		}
		var (
			fullAddress = md.Get(p2pUtil.DefaultConnectionMetadata)[0]
			peer, err   = p2pUtil.ParsePeer(fullAddress)
		)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalidPeer")
		}
		blockService := ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		lastBlock, err := blockService.GetLastBlock()
		if err != nil {
			return nil, status.Error(
				codes.Internal,
				"failGetLastBlock",
			)
		}
		receipt, err := blockService.ReceiveBlock(
			senderPublicKey,
			lastBlock,
			block,
			ps.NodeSecretPhrase,
			peer,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendBlockResponse{
			Receipt: receipt,
		}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ps *P2PServerService) SendTransaction(
	ctx context.Context,
	chainType chaintype.ChainType,
	transactionBytes,
	senderPublicKey []byte,
) (*model.SendTransactionResponse, error) {
	if limitReached, limitLevel := ps.FeedbackStrategy.IsGoroutineLimitReached(constant.FeedbackMinGoroutineSamples); limitReached {
		if limitLevel == constant.FeedbackLimitHigh {
			monitoring.IncreaseP2PTxFiltered()
			return nil, status.Error(codes.Internal, "NodeIsBusy")
		}
	}
	if limitReached, limitLevel := ps.FeedbackStrategy.IsP2PRequestLimitReached(constant.FeedbackMinGoroutineSamples); limitReached {
		if limitLevel == constant.FeedbackLimitCritical {
			monitoring.IncreaseP2PTxFiltered()
			return nil, status.Error(codes.Internal, "TooManyP2PRequests")
		}
	}

	if ps.PeerExplorer.ValidateRequest(ctx) {
		var blockService = ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		lastBlockCacheFormat, err := blockService.GetLastBlockCacheFormat()
		if err != nil {
			return nil, status.Error(
				codes.Internal,
				fmt.Sprintf("failGetLastBlockErr: %v", err.Error()),
			)
		}
		var mempoolService = ps.MempoolServices[chainType.GetTypeInt()]
		if mempoolService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"mempoolServiceNotFoundByThisChainType",
			)
		}
		receipt, err := mempoolService.ReceivedTransaction(
			senderPublicKey,
			transactionBytes,
			lastBlockCacheFormat,
			ps.NodeSecretPhrase,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendTransactionResponse{
			Receipt: receipt,
		}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// SendBlockTransactions receive a list of transaction from other node and calling TransactionReceived Event
func (ps *P2PServerService) SendBlockTransactions(
	ctx context.Context,
	chainType chaintype.ChainType,
	transactionsBytes [][]byte,
	senderPublicKey []byte,
) (*model.SendBlockTransactionsResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var blockService = ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"blockServiceNotFoundByThisChainType",
			)
		}
		lastBlockCacheFormat, err := blockService.GetLastBlockCacheFormat()
		if err != nil {
			return nil, status.Error(
				codes.Internal,
				fmt.Sprintf("failGetLastBlockErr: %v", err.Error()),
			)
		}
		var mempoolService = ps.MempoolServices[chainType.GetTypeInt()]
		if mempoolService == nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"mempoolServiceNotFoundByThisChainType",
			)
		}
		batchReceipts, err := mempoolService.ReceivedBlockTransactions(
			senderPublicKey,
			transactionsBytes,
			lastBlockCacheFormat,
			ps.NodeSecretPhrase,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendBlockTransactionsResponse{
			Receipts: batchReceipts,
		}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

func (ps *P2PServerService) RequestBlockTransactions(
	ctx context.Context,
	chainType chaintype.ChainType,
	blockID int64,
	transactionsIDs []int64,
) (*model.Empty, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var md, _ = metadata.FromIncomingContext(ctx)
		if len(md) == 0 {
			return nil, status.Error(
				codes.InvalidArgument,
				"invalidContext",
			)
		}
		var (
			fullAddress = md.Get(p2pUtil.DefaultConnectionMetadata)[0]
			peer, err   = p2pUtil.ParsePeer(fullAddress)
		)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalidPeer")
		}
		ps.Observer.Notify(observer.BlockTransactionsRequested, transactionsIDs, chainType, blockID, peer)
		return &model.Empty{}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

func (ps *P2PServerService) RequestDownloadFile(
	ctx context.Context,
	snapshotHash []byte,
	fileChunkNames []string,
) (*model.FileDownloadResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var (
			fileChunks = make([][]byte, 0)
			failed     []string
		)
		for _, fileName := range fileChunkNames {
			chunk, err := ps.FileService.ReadFileFromDir(base64.URLEncoding.EncodeToString(snapshotHash), fileName)
			if err != nil {
				failed = append(failed, fileName)
			} else {
				fileChunks = append(fileChunks, chunk)
			}
		}
		res := &model.FileDownloadResponse{
			FileChunks: fileChunks,
			Failed:     failed,
		}
		return res, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}
