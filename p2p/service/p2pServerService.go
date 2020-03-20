package service

import (
	"context"
	"errors"

	"github.com/zoobc/zoobc-core/common/blocker"
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
		RequestDownloadFile(
			ctx context.Context,
			fileChunkNames []string,
		) (*model.FileDownloadResponse, error)
	}

	P2PServerService struct {
		FileService      coreService.FileServiceInterface
		PeerExplorer     strategy.PeerExplorerStrategyInterface
		BlockServices    map[int32]coreService.BlockServiceInterface
		MempoolServices  map[int32]coreService.MempoolServiceInterface
		NodeSecretPhrase string
		Observer         *observer.Observer
	}
)

func NewP2PServerService(
	fileService coreService.FileServiceInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	mempoolServices map[int32]coreService.MempoolServiceInterface,
	nodeSecretPhrase string,
	observer *observer.Observer,
) *P2PServerService {

	return &P2PServerService{
		FileService:      fileService,
		PeerExplorer:     peerExplorer,
		BlockServices:    blockServices,
		MempoolServices:  mempoolServices,
		NodeSecretPhrase: nodeSecretPhrase,
		Observer:         observer,
	}
}

func (ps *P2PServerService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		return ps.PeerExplorer.GetHostInfo(), nil
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
		err := ps.PeerExplorer.AddToUnresolvedPeers(peers, true)
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
		lastBlock, err := blockService.GetLastBlock(0)
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
		myLastBlock, err := blockService.GetLastBlock(0)
		if err != nil || myLastBlock == nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"failed to get last block",
			)
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
				return nil, err
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
			block, err := blockService.GetBlockByHeight(height)
			if err != nil {
				return nil, err
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
		var blockIds []int64
		blockService := ps.BlockServices[chainType.GetTypeInt()]
		if blockService == nil {
			return nil, errors.New("the block service is not set for this chaintype in this host")
		}
		limit := constant.PeerGetBlocksLimit
		if reqLimit != 0 && reqLimit < limit {
			limit = reqLimit
		}

		foundBlock, err := blockService.GetBlockByID(reqBlockID, false)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, err.Error())
		}
		blocks, err := blockService.GetBlocksFromHeight(foundBlock.Height, limit, false)
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
		var blocksMessage []*model.Block
		blockService := ps.BlockServices[chainType.GetTypeInt()]
		blockService.ChainWriteLock(constant.BlockchainSendingBlocks)
		defer blockService.ChainWriteUnlock(constant.BlockchainSendingBlocks)
		block, err := blockService.GetBlockByID(blockID, false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		blocks, err := blockService.GetBlocksFromHeight(block.Height, uint32(len(blockIDList)), true)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		for idx, block := range blocks {
			if block.ID != blockIDList[idx] {
				break
			}
			err = ps.BlockServices[chainType.GetTypeInt()].PopulateBlockData(block)
			if err != nil {
				return nil, err
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
		md, _ := metadata.FromIncomingContext(ctx)
		fullAddress := md.Get(p2pUtil.DefaultConnectionMetadata)[0]
		peer, _ := p2pUtil.ParsePeer(fullAddress)
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
			peer,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendBlockResponse{
			BatchReceipt: batchReceipt,
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
	if ps.PeerExplorer.ValidateRequest(ctx) {
		lastBlock, err := ps.BlockServices[chainType.GetTypeInt()].GetLastBlock(1)
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
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ps *P2PServerService) SendBlockTransactions(
	ctx context.Context,
	chainType chaintype.ChainType,
	transactionsBytes [][]byte,
	senderPublicKey []byte,
) (*model.SendBlockTransactionsResponse, error) {
	if ps.PeerExplorer.ValidateRequest(ctx) {
		lastBlock, err := ps.BlockServices[chainType.GetTypeInt()].GetLastBlock()
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.BlockErr,
				"fail to get last block",
			)
		}

		batchReceipts, err := ps.MempoolServices[chainType.GetTypeInt()].ReceivedBlockTransactions(
			senderPublicKey,
			transactionsBytes,
			lastBlock,
			ps.NodeSecretPhrase,
		)
		if err != nil {
			return nil, err
		}
		return &model.SendBlockTransactionsResponse{
			BatchReceipts: batchReceipts,
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
		md, _ := metadata.FromIncomingContext(ctx)
		fullAddress := md.Get(p2pUtil.DefaultConnectionMetadata)[0]
		peer, err := p2pUtil.ParsePeer(fullAddress)
		if err != nil {
			_ = status.Error(codes.InvalidArgument, "Invalid requester data")
		}
		ps.Observer.Notify(observer.BlockTransactionsRequested, transactionsIDs, chainType, blockID, peer)
		return &model.Empty{}, nil
	}
	return nil, status.Error(codes.Unauthenticated, "Rejected request")
}

func (ps *P2PServerService) RequestDownloadFile(
	ctx context.Context,
	fileChunkNames []string,
) (*model.FileDownloadResponse, error) {
	var (
		fileChunks = make([][]byte, 0)
		failed     []string
	)
	if ps.PeerExplorer.ValidateRequest(ctx) {
		for _, fileName := range fileChunkNames {
			chunkBytes, err := ps.FileService.ReadFileByName(ps.FileService.GetDownloadPath(), fileName)
			if err != nil {
				failed = append(failed, fileName)
			} else {
				fileChunks = append(fileChunks, chunkBytes)
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
