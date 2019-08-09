package service

import (
	"context"
	"errors"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"

	"google.golang.org/grpc"
)

var (
	apiLogger *log.Logger
)

// ServerService represent data service node as server
type ServerService struct {
	BlockServices map[int32]coreService.BlockServiceInterface
}

var serverServiceInstance *ServerService

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debugP2P.log"); err != nil {
		panic(err)
	}
}

func NewServerService(blockServices map[int32]coreService.BlockServiceInterface) *ServerService {
	if serverServiceInstance == nil {
		serverServiceInstance = &ServerService{
			BlockServices: blockServices,
		}
	}
	return serverServiceInstance
}

// StartListening to grpc connection
func (ss *ServerService) StartListening(listener net.Listener) error {
	hs, err := GetHostService()
	if err != nil {
		panic(err)
	}
	hostInfo := hs.Host.GetInfo()
	if hostInfo.GetAddress() == "" || hostInfo.GetPort() == 0 {
		log.Fatalf("Address or Port server is not available")
	}

	apiLogger.Info("P2P: Listening to grpc communication...")
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(util.NewServerInterceptor(apiLogger)),
	)
	service.RegisterP2PCommunicationServer(grpcServer, ss)
	return grpcServer.Serve(listener)
}

// GetPeerInfo to return info of this host
func (ss *ServerService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	hs, err := GetHostService()
	if err != nil {
		panic(err)
	}
	hostInfo := hs.Host.GetInfo()
	return &model.Node{
		SharedAddress: hostInfo.GetSharedAddress(),
		Address:       hostInfo.GetAddress(),
		Port:          hostInfo.GetPort(),
	}, nil
}

// GetMorePeers contains info other peers
func (ss *ServerService) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {
	var nodes []*model.Node
	hs, err := GetHostService()
	if err != nil {
		panic(err)
	}
	// only sends the connected (resolved) peers
	for _, hostPeer := range hs.GetResolvedPeers() {
		nodes = append(nodes, hostPeer.GetInfo())
	}
	peers := &model.GetMorePeersResponse{
		Peers: nodes,
	}
	return peers, nil
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (ss *ServerService) SendPeers(ctx context.Context, req *model.SendPeersRequest) (*model.Empty, error) {
	// TODO: only accept nodes that are already registered in the node registration
	hs, err := GetHostService()
	if err != nil {
		panic(err)
	}
	_ = hs.AddToUnresolvedPeers(req.Peers, true)
	return &model.Empty{}, nil
}

// GetCumulativeDifficulty responds to the request of the cummulative difficulty status of a node
func (ss *ServerService) GetCumulativeDifficulty(ctx context.Context, req *model.GetCumulativeDifficultyRequest) (*model.GetCumulativeDifficultyResponse, error) {
	blockService := ss.BlockServices[req.ChainType]
	lastBlock, err := blockService.GetLastBlock()
	if err != nil {
		return nil, err
	}
	return &model.GetCumulativeDifficultyResponse{
		CumulativeDifficulty: lastBlock.CumulativeDifficulty,
		Height:               lastBlock.Height,
	}, nil
}

func (ss *ServerService) GetCommonMilestoneBlockIDs(ctx context.Context, req *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	// if `lastBlockId` is supplied
	// check it the last `lastBlockId` got matches with the host's lastBlock then return the response as is
	chainType := chaintype.GetChainType(req.ChainType)
	blockService := ss.BlockServices[chainType.GetTypeInt()]
	if blockService == nil {
		return nil, errors.New("The block service is not set for this chaintype in this host")
	}

	lastBlockId := req.LastBlockId
	myLastBlock, err := blockService.GetLastBlock()
	if err != nil || myLastBlock == nil {
		return nil, errors.New("failed to get last block")
	}
	myLastBlockId := myLastBlock.ID
	myBlockchainHeight := myLastBlock.Height
	if block, _ := blockService.GetBlockByID(lastBlockId); block != nil || lastBlockId == myLastBlockId {
		preparedResponse := &model.GetCommonMilestoneBlockIdsResponse{
			BlockIds: []int64{lastBlockId},
		}
		if lastBlockId == myLastBlockId {
			preparedResponse.Last = true
		}
		return preparedResponse, nil
	}

	// if not, send (assumed) milestoneBlock of the host
	var height, jump uint32
	limit := constant.CommonMilestoneBlockIdsLimit
	lastMilestoneBlockId := req.LastMilestoneBlockId
	if lastMilestoneBlockId != 0 {
		lastMilestoneBlock, _ := blockService.GetBlockByID(lastMilestoneBlockId)
		if lastMilestoneBlock == nil {
			return &model.GetCommonMilestoneBlockIdsResponse{BlockIds: []int64{}}, errors.New("block not found")
		}
		height = lastMilestoneBlock.GetHeight()
		jump = util.MinUint32(constant.SafeBlockGap, util.MaxUint32(myBlockchainHeight, 1))
	} else if lastBlockId != 0 {
		// TODO: analyze difference of height jump
		height = myBlockchainHeight
		jump = 10
	}

	block, err := blockService.GetBlockByHeight(height)
	if block == nil || err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get block at height %v, probably because of corrupted data", height))
	}
	blockIdAtHeight := block.ID
	blockIds := []int64{}
	for {
		limit = limit - 1
		if height > 0 && limit > 0 {
			blockIds = append(blockIds, blockIdAtHeight)
			height = height - jump
			block, err := blockService.GetBlockByHeight(height)
			if block == nil || err != nil {
				return nil, errors.New(fmt.Sprintf("failed to get block at height %v, probably because of corrupted data", height))
			}
			blockIdAtHeight = block.ID
		} else {
			break
		}
		if limit < 1 {
			break
		}
	}

	return &model.GetCommonMilestoneBlockIdsResponse{BlockIds: blockIds}, nil
}

func (ss *ServerService) GetNextBlockIDs(ctx context.Context, req *model.GetNextBlockIdsRequest) (*model.BlockIdsResponse, error) {
	chainType := chaintype.GetChainType(req.ChainType)
	blockService := ss.BlockServices[chainType.GetTypeInt()]
	if blockService == nil {
		return nil, errors.New("The block service is not set for this chaintype in this host")
	}
	reqLimit := req.Limit
	reqBlockId := req.BlockId
	limit := constant.PeerGetBlocksLimit
	if reqLimit != 0 && reqLimit < limit {
		limit = reqLimit
	}

	foundBlock, err := blockService.GetBlockByID(reqBlockId)
	if foundBlock == nil || foundBlock.ID == -1 || err != nil {
		return &model.BlockIdsResponse{}, errors.New(fmt.Sprintf("the block with id %v is not found", reqBlockId))
	}
	blocks, err := blockService.GetBlocksFromHeight(foundBlock.Height, limit)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get the block IDs: %v\n", err))
	}
	if len(blocks) == 0 {
		return &model.BlockIdsResponse{}, nil
	}

	blockIds := []int64{}
	for _, block := range blocks {
		blockIds = append(blockIds, block.ID)
	}

	return &model.BlockIdsResponse{BlockIds: blockIds}, nil
}

func (ss *ServerService) GetNextBlocks(ctx context.Context, req *model.GetNextBlocksRequest) (*model.BlocksData, error) {
	// TODO: getting data from cache
	chainType := chaintype.GetChainType(req.ChainType)
	blockService := ss.BlockServices[chainType.GetTypeInt()]

	reqBlockId := req.BlockId
	reqBlockIdList := req.BlockIds
	blocksMessage := []*model.Block{}
	block, err := blockService.GetBlockByID(reqBlockId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("can not find block with ID %v: \n", reqBlockId, err))
	}
	blocks, err := blockService.GetBlocksFromHeight(block.Height, uint32(len(reqBlockIdList)))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to get the blocks: %v\n", err))
	}
	for idx, block := range blocks {
		if block.ID != req.BlockIds[idx] {
			break
		}

		blocksMessage = append(blocksMessage, block)
	}
	return &model.BlocksData{Blocks: blocksMessage}, nil
}
