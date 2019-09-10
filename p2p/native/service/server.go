package service

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/interceptor"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"

	"google.golang.org/grpc"
)

var (
	apiLogger *log.Logger
)

// ServerService represent data service node as server
type ServerService struct {
	BlockServices map[int32]coreService.BlockServiceInterface
	Observer      *observer.Observer
}

var serverServiceInstance *ServerService

func init() {
	var (
		err       error
		logLevels []string
	)
	logLevels = viper.GetStringSlice("logLevels")
	if apiLogger, err = util.InitLogger(".log/", "debugP2P.log", logLevels); err != nil {
		panic(err)
	}
}

func NewServerService(blockServices map[int32]coreService.BlockServiceInterface, obsr *observer.Observer) *ServerService {
	if serverServiceInstance == nil {
		serverServiceInstance = &ServerService{
			BlockServices: blockServices,
			Observer:      obsr,
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
	ownerAccountAddress := viper.GetString("ownerAccountAddress")
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.NewServerInterceptor(apiLogger, ownerAccountAddress)),
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
func (ss *ServerService) GetCumulativeDifficulty(ctx context.Context,
	req *model.GetCumulativeDifficultyRequest) (*model.GetCumulativeDifficultyResponse, error) {
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

func (ss *ServerService) GetCommonMilestoneBlockIDs(ctx context.Context,
	req *model.GetCommonMilestoneBlockIdsRequest) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	// if `lastBlockID` is supplied
	// check it the last `lastBlockID` got matches with the host's lastBlock then return the response as is
	chainType := chaintype.GetChainType(req.ChainType)
	lastBlockID := req.LastBlockID
	lastMilestoneBlockID := req.LastMilestoneBlockID

	blockService := ss.BlockServices[chainType.GetTypeInt()]
	if blockService == nil {
		return nil, errors.New("the block service is not set for this chaintype in this host")
	}

	if lastBlockID == 0 && lastMilestoneBlockID == 0 {
		return nil, blocker.NewBlocker(blocker.RequestParameterErr, "either LastBlockID or LastMilestoneBlockID has to be supplied")
	}
	myLastBlock, err := blockService.GetLastBlock()
	if err != nil || myLastBlock == nil {
		return nil, errors.New("failed to get last block")
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
	var height, jump uint32
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

	blockIds := []int64{}
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

func (ss *ServerService) GetNextBlockIDs(ctx context.Context, req *model.GetNextBlockIdsRequest) (*model.BlockIdsResponse, error) {
	chainType := chaintype.GetChainType(req.ChainType)
	blockService := ss.BlockServices[chainType.GetTypeInt()]
	if blockService == nil {
		return nil, errors.New("the block service is not set for this chaintype in this host")
	}
	reqLimit := req.Limit
	reqBlockID := req.BlockId
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

	reqBlockID := req.BlockId
	reqBlockIDList := req.BlockIds
	blocksMessage := []*model.Block{}
	block, err := blockService.GetBlockByID(reqBlockID)
	if err != nil {
		return nil, err
	}
	blocks, err := blockService.GetBlocksFromHeight(block.Height, uint32(len(reqBlockIDList)))
	if err != nil {
		return nil, fmt.Errorf("failed to get the blocks: %v", err)
	}
	for idx, block := range blocks {
		if block.ID != reqBlockIDList[idx] {
			break
		}

		blocksMessage = append(blocksMessage, block)
	}
	return &model.BlocksData{NextBlocks: blocksMessage}, nil
}

// SendBlock receive block from other node and calling BlockReceived Event
func (ss *ServerService) SendBlock(ctx context.Context, req *model.Block) (*model.Empty, error) {
	ss.Observer.Notify(observer.BlockReceived, req, nil)
	return &model.Empty{}, nil
}

// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ss *ServerService) SendTransaction(ctx context.Context, req *model.SendTransactionRequest) (*model.Empty, error) {
	ss.Observer.Notify(observer.TransactionReceived, req.GetTransactionBytes(), nil)
	return &model.Empty{}, nil
}
