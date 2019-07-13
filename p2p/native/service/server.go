package service

import (
	"context"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"

	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
	"google.golang.org/grpc"
)

var (
	apiLogger *log.Logger
)

// HostService represent data service node as server
type HostService struct {
	Host       *model.Host
	GrpcServer *grpc.Server
	ChainType  contract.ChainType
}

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debugP2P.log"); err != nil {
		panic(err)
	}
}

// StopServer function to stop current running host service
// func (hs *HostService) stopServer(gracefully bool) {
// 	if hs.GrpcServer != nil {
// 		if gracefully {
// 			hs.GrpcServer.GracefulStop()
// 		} else {
// 			hs.GrpcServer.Stop()
// 		}
// 	}
// }

// StartListening to
func (hs *HostService) StartListening() {
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
	err2 := hs.GrpcServer.Serve(serv)

	if err2 != nil {
		log.Fatalf("GRPC failed to serve: %v", err)
	}
}

// GetPeerInfo to
func (hs *HostService) GetPeerInfo(ctx context.Context, req *model.GetPeerInfoRequest) (*model.Node, error) {
	return &model.Node{
		SharedAddress: hs.Host.GetInfo().GetSharedAddress(),
		Address:       hs.Host.GetInfo().GetAddress(),
		Port:          hs.Host.GetInfo().GetPort(),
	}, nil
}

// GetMorePeers contains info other peers
func (hs *HostService) GetMorePeers(ctx context.Context, req *model.Empty) (*model.GetMorePeersResponse, error) {

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

// ResolvePeers looping unresolve peers and adding to (resolve) Peers if get response
func (hs *HostService) ResolvePeers() {
	for _, peer := range hs.Host.GetUnresolvedPeers() {
		go hs.resolvePeer(peer)
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (hs *HostService) resolvePeer(destPeer *model.Peer) {
	_, err := ClientPeerService(hs.ChainType).GetPeerInfo(destPeer)
	if err != nil {
		return
	}
	updatedHost := nativeUtil.AddToResolvedPeer(hs.Host, destPeer)
	hs.Host = updatedHost

	log.Info(nativeUtil.GetFullAddressPeer(destPeer) + " success")
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (hs *HostService) GetMorePeersHandler() {
	peer := nativeUtil.GetAnyPeer(hs.Host)
	if peer != nil {
		newPeers, err := ClientPeerService(hs.ChainType).GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error accord %v\n", err)
		}
		hs.Host = nativeUtil.AddToUnresolvedPeers(hs.Host, newPeers.GetPeers())
	}
}
