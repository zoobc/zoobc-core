package service

import (
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
		GetPeerInfo(req *model.GetPeerInfoRequest) (*model.Node, error)
		GetMorePeers(req *model.Empty) ([]*model.Node, error)
		SendPeers(peers []*model.Node) (*model.Empty, error)
		GetCumulativeDifficulty(
			chainType chaintype.ChainType,
		) (*model.GetCumulativeDifficultyResponse, error)
		GetCommonMilestoneBlockIDs(
			chainType chaintype.ChainType,
			lastBlockID int64,
			lastMilestoneBlockID int64,
		) (*model.GetCommonMilestoneBlockIdsResponse, error)
		GetNextBlockIDs(
			chainType chaintype.ChainType,
			reqLimit uint32,
			reqBlockID int64,
		) ([]int64, error)
		GetNextBlocks(
			chainType chaintype.ChainType,
			blockID int64,
			blockIDList []int64,
		) (*model.BlocksData, error)
		SendBlock(
			chainType chaintype.ChainType,
			block *model.Block,
		) (*model.Receipt, error)
		SendTransaction(req *model.SendTransactionRequest) (*model.Receipt, error)
	}

	P2PServerService struct {
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		NodeSecretPhrase string
	}
)

func NewP2PServerService(
	peerExplorer strategy.PeerExplorerStrategyInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeSecretPhrase string,
) *P2PServerService {
	return &P2PServerService{
		PeerExplorer:     peerExplorer,
		BlockServices:    blockServices,
		NodeSecretPhrase: nodeSecretPhrase,
	}
}

func (ps *P2PServerService) GetPeerInfo(req *model.GetPeerInfoRequest) (*model.Node, error) {
	//return ps.PeerExplorer.
	return ps.PeerExplorer.GetHostInfo(), nil
}

// GetMorePeers contains info other peers
func (ps *P2PServerService) GetMorePeers(req *model.Empty) ([]*model.Node, error) {
	var nodes []*model.Node
	// only sends the connected (resolved) peers
	for _, hostPeer := range ps.PeerExplorer.GetResolvedPeers() {
		nodes = append(nodes, hostPeer.GetInfo())
	}
	return nodes, nil
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (ps *P2PServerService) SendPeers(
	peers []*model.Node,
) (*model.Empty, error) {
	// TODO: only accept nodes that are already registered in the node registration
	err := ps.PeerExplorer.AddToUnresolvedPeers(peers, true)
	if err != nil {
		return nil, err
	}
	return &model.Empty{}, nil
}

// GetCumulativeDifficulty responds to the request of the cummulative difficulty status of a node
func (ps *P2PServerService) GetCumulativeDifficulty(
	chainType chaintype.ChainType,
) (*model.GetCumulativeDifficultyResponse, error) {
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

func (ps P2PServerService) GetCommonMilestoneBlockIDs(
	chainType chaintype.ChainType,
	lastBlockID int64,
	lastMilestoneBlockID int64,
) (*model.GetCommonMilestoneBlockIdsResponse, error) {
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
	myLastBlock, err := blockService.GetLastBlock()
	if err != nil || myLastBlock == nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			"failed to get last block",
		)
	}

	if _, err := blockService.GetBlockByID(lastBlockID); err == nil {
		preparedResponse := &model.GetCommonMilestoneBlockIdsResponse{
			BlockIds: []int64{lastBlockID},
		}
		if lastBlockID == myLastBlock.ID {
			preparedResponse.Last = true
		}
		return preparedResponse, nil
	}

	// if not, send (assumed) milestoneBlock of the host
	limit := constant.CommonMilestoneBlockIdsLimit
	if lastMilestoneBlockID != 0 {
		lastMilestoneBlock, err := blockService.GetBlockByID(lastMilestoneBlockID)
		if err != nil {
			return nil, err
		}
		height = lastMilestoneBlock.GetHeight()
		jump = util.MinUint32(constant.SafeBlockGap, util.MaxUint32(myLastBlock.Height, 1))
	} else if lastBlockID != 0 {
		// TODO: analyze difference of height jump
		height = myLastBlock.Height
		jump = 10
	}

	block, err := blockService.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	blockIDAtHeight := block.ID

	for {
		limit--
		if height > 0 && limit > 0 {
			blockIds = append(blockIds, blockIDAtHeight)
			height -= jump
			block, err := blockService.GetBlockByHeight(height)
			if err != nil {
				return nil, err
			}
			blockIDAtHeight = block.ID
		} else {
			break
		}
		if limit < 1 {
			break
		}
	}
	return &model.GetCommonMilestoneBlockIdsResponse{BlockIds: blockIds}, nil
}

func (ps *P2PServerService) GetNextBlockIDs(
	chainType chaintype.ChainType,
	reqLimit uint32,
	reqBlockID int64,
) ([]int64, error) {
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
		return nil, fmt.Errorf("failed to get the block IDs: %v", err)
	}

	if len(blocks) == 0 {
		return blockIds, nil
	}

	for _, block := range blocks {
		blockIds = append(blockIds, block.ID)
	}

	return blockIds, nil
}

func (ps *P2PServerService) GetNextBlocks(
	chainType chaintype.ChainType,
	blockID int64,
	blockIDList []int64,
) (*model.BlocksData, error) {
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

		blocksMessage = append(blocksMessage, block)
	}
	return &model.BlocksData{NextBlocks: blocksMessage}, nil
}

// SendBlock receive block from other node and calling BlockReceived Event
func (ps *P2PServerService) SendBlock(
	chainType chaintype.ChainType,
	block *model.Block,
) (*model.Receipt, error) {
	receipt, err := ps.BlockServices[chainType.GetTypeInt()].ReceiveBlock(block, ps.NodeSecretPhrase)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ps *P2PServerService) SendTransaction(req *model.SendTransactionRequest) (*model.Receipt, error) {
	//ss.Observer.Notify(observer.TransactionReceived, req.GetTransactionBytes(), nil)
	return &model.Receipt{}, nil
}
