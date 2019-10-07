package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/strategy"
)

type (
	P2PServerServiceInterface interface {
		GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error)
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
	}

	P2PServerService struct {
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
	}
)

func NewP2PServerService(
	peerExplorer strategy.PeerExplorerStrategyInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	mempoolServices map[int32]coreService.MempoolServiceInterface,
	nodeSecretPhrase string,
) *P2PServerService {

	return &P2PServerService{
		PeerExplorer:     peerExplorer,
		BlockServices:    blockServices,
		MempoolServices:  mempoolServices,
		NodeSecretPhrase: nodeSecretPhrase,
	}
}

func (ps *P2PServerService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		return ps.PeerExplorer.GetHostInfo(), nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
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
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (ps *P2PServerService) SendPeers(
	ctx context.Context,
	peers []*model.Node,
) (*model.Empty, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		// TODO: only accept nodes that are already registered in the node registration
		err := ps.PeerExplorer.AddToUnresolvedPeers(peers, true)
		if err != nil {
			return nil, err
		}
		return &model.Empty{}, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}

// GetCumulativeDifficulty responds to the request of the cumulative difficulty status of a node
func (ps *P2PServerService) GetCumulativeDifficulty(
	ctx context.Context,
	chainType chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		blockService := ps.BlockServices[chainType.GetTypeInt()]
		lastBlock, err := blockService.GetLastBlock()
		if err != nil {
			return nil, err
		}
		return &model.GetCumulativeDifficultyResponse{
			CumulativeDifficulty: lastBlock.CumulativeDifficulty,
			Height:               lastBlock.Height,
		}, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
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
		)

		blockService := ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"the block service is not set for this chaintype in this host",
			)
		}

		if lastBlockID == 0 && lastMilestoneBlockID == 0 {
			return nil, blocker.NewBlocker(blocker.RequestParameterErr, "either LastBlockID or LastMilestoneBlockID has to be supplied")
		}
		myLastBlock, err := blockService.GetLastBlock()
		if err != nil || myLastBlock == nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"failed to get last block",
			)
		}
		myLastBlockID := myLastBlock.ID
		myBlockchainHeight := myLastBlock.Height

		if _, err := blockService.GetBlockByID(lastBlockID); err == nil {
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
			lastMilestoneBlock, err := blockService.GetBlockByID(lastMilestoneBlockID)
			// this error is handled because when lastMilestoneBlockID is provided, it was expected to be the one returned from this node
			if err != nil {
				return nil, err
			}
			height = lastMilestoneBlock.GetHeight()
			jump = util.MinUint32(constant.SafeBlockGap, util.MaxUint32(myBlockchainHeight, 1))
		} else if lastBlockID != 0 {
			// TODO: analyze difference of height jump
			height = myBlockchainHeight
			jump = 10
		}

		for ; limit > 0; limit-- {
			block, err := blockService.GetBlockByHeight(height)
			if err != nil {
				return nil, err
			}
			blockIds = append(blockIds, block.ID)
			switch {
			case height == 0:
				break
			case height < jump:
				height = 0
			default:
				height -= jump
			}
		}

		return &model.GetCommonMilestoneBlockIdsResponse{BlockIds: blockIds}, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}

func (ps *P2PServerService) GetNextBlockIDs(
	ctx context.Context,
	chainType chaintype.ChainType,
	reqLimit uint32,
	reqBlockID int64,
) ([]int64, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		var blockIds []int64
		blockService := ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, errors.New("the block service is not set for this chaintype in this host")
		}
		limit := constant.PeerGetBlocksLimit
		if reqLimit != 0 && reqLimit < limit {
			limit = reqLimit
		}

		foundBlock, err := blockService.GetBlockByID(reqBlockID)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, err.Error())
		}
		blocks, err := blockService.GetBlocksFromHeight(foundBlock.Height, limit)
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"failed to get block id",
			)
		}

		if len(blocks) == 0 {
			return blockIds, nil
		}

		for _, block := range blocks {
			blockIds = append(blockIds, block.ID)
		}

		return blockIds, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}

func (ps *P2PServerService) GetNextBlocks(
	ctx context.Context,
	chainType chaintype.ChainType,
	blockID int64,
	blockIDList []int64,
) (*model.BlocksData, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		// TODO: getting data from cache
		var blocksMessage []*model.Block
		blockService := ps.BlockServices[chainType.GetTypeInt()]

		block, err := blockService.GetBlockByID(blockID)
		if err != nil {
			return nil, err
		}
		blocks, err := blockService.GetBlocksFromHeight(block.Height, uint32(len(blockIDList)))
		if err != nil {
			return nil, fmt.Errorf("failed to get the blocks: %v", err)
		}
		for idx, block := range blocks {
			if block.ID != blockIDList[idx] {
				break
			}

			txs, _ := ps.BlockServices[chainType.GetTypeInt()].GetTransactionsByBlockID(block.ID)
			block.Transactions = txs
			blocksMessage = append(blocksMessage, block)
		}
		return &model.BlocksData{NextBlocks: blocksMessage}, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}

// SendBlock receive block from other node
func (ps *P2PServerService) SendBlock(
	ctx context.Context,
	chainType chaintype.ChainType,
	block *model.Block,
	senderPublicKey []byte,
) (*model.SendBlockResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		lastBlock, err := ps.BlockServices[chainType.GetTypeInt()].GetLastBlock()
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"fail to get last block",
			)
		}
		batchReceipt, err := ps.BlockServices[chainType.GetTypeInt()].ReceiveBlock(
			senderPublicKey,
			lastBlock,
			block,
			ps.NodeSecretPhrase,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendBlockResponse{
			BatchReceipt: batchReceipt,
		}, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ps *P2PServerService) SendTransaction(
	ctx context.Context,
	chainType chaintype.ChainType,
	transactionBytes,
	senderPublicKey []byte,
) (*model.SendTransactionResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		lastBlock, err := ps.BlockServices[chainType.GetTypeInt()].GetLastBlock()
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"fail to get last block",
			)
		}

		batchReceipt, err := ps.MempoolServices[chainType.GetTypeInt()].ReceivedTransaction(
			senderPublicKey,
			transactionBytes,
			lastBlock,
			ps.NodeSecretPhrase,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendTransactionResponse{
			BatchReceipt: batchReceipt,
		}, nil
	}
	return nil, blocker.NewBlocker(blocker.ValidationErr, "Rejected request")
}
