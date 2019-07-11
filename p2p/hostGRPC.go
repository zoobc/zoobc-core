package p2p

import (
	"context"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

var (
	apiLogger *log.Logger
)

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debugP2P.log"); err != nil {
		panic(err)
	}
}

// Start func to start host service in new goroutine
func (hs HostService) Start() {
	go hs.startListening()
}

// Stop function to stop current running host service
func (hs HostService) Stop(gracefully bool) {
	if hs.GrpcServer != nil {
		if gracefully {
			hs.GrpcServer.GracefulStop()
		} else {
			hs.GrpcServer.Stop()
		}
	}
}

func (hs HostService) startListening() {
	if hs.Host.GetInfo().GetAddress() == "" || hs.Host.GetInfo().GetPort() == 0 {
		log.Fatalf("host is not setup")
	}
	serv, err := net.Listen("tcp", ":"+strconv.Itoa(int(hs.Host.GetInfo().GetPort())))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	apiLogger.Info("P2P: Listening to grpc communication...")
	hs.GrpcServer = grpc.NewServer()
	service.RegisterP2PCommunicationServer(hs.GrpcServer, hs)
	hs.GrpcServer.Serve(serv)
}

// GetPeerInfo to
func (hs HostService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	return &model.Node{SharedAddress: hs.Host.GetInfo().GetSharedAddress(), Port: hs.Host.GetInfo().GetPort()}, nil
}

// GetMorePeers contains info other peers
func (hs HostService) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {

	var nodes []*model.Node
	for _, hostPeer := range hs.Host.Peers {
		nodes = append(nodes, hostPeer.GetInfo())
	}
	for _, hostPeer := range hs.Host.UnresolvedPeers {
		nodes = append(nodes, hostPeer.GetInfo())
	}
	peers := &model.GetMorePeersResponse{
		Peers: nodes,
	}
	return peers, nil
}

// func (hs hostService) AddPeers(ctx context.Context, req *model.PostAddPeersRequest) (*model.PostAddPeersResponse, error) {
// 	hs.Chaintype = hs.getChaintype(ctx)
// 	var peersInfo = []*pbMessages.NodeInfo{}
// 	for _, knownPeer := range hs.Host.GetKnownPeers() {
// 		peerInfo := &pbMessages.NodeInfo{Version: "v1.0.0", SharedAddress: knownPeer.GetFullAddress()}
// 		peersInfo = append(peersInfo, peerInfo)
// 	}
// 	return &pbMessages.PeersInfo{Peers: peersInfo}, nil
// }
