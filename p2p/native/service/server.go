package service

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"

	"google.golang.org/grpc"
)

var (
	apiLogger *log.Logger
)

// ServerService represent data service node as server
type ServerService struct{}

var serverServiceInstance *ServerService

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debugP2P.log"); err != nil {
		panic(err)
	}
}

func NewServerService() *ServerService {
	if serverServiceInstance == nil {
		serverServiceInstance = &ServerService{}
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

// TODO: test case
// SendBlock receive block from other node and calling BlockReceived Event
func (ss *ServerService) SendBlock(ctx context.Context, req *model.Block) (*model.Empty, error) {
	observer.NewObserver().Notify(observer.BlockReceived, req, nil)
	return &model.Empty{}, nil
}

// TODO: test case
// SendTransaction receive transaction from other node and calling TransactionReceived Event
func (ss *ServerService) SendTransaction(ctx context.Context, req *model.Transaction) (*model.Empty, error) {
	observer.NewObserver().Notify(observer.TransactionReceived, req, nil)
	return &model.Empty{}, nil
}
