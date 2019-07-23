package service

import (
	"context"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/p2p/native/internal"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
	"google.golang.org/grpc"
)

var (
	apiLogger *log.Logger
)

// HostService represent data service node as server
type HostService struct {
	Host *model.Host
}

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debugP2P.log"); err != nil {
		panic(err)
	}
}

// StartListening to
func (hs *HostService) StartListening(listener net.Listener) error {
	if hs.Host.GetInfo().GetAddress() == "" || hs.Host.GetInfo().GetPort() == 0 {
		log.Fatalf("Address or Port server is not available")
	}

	apiLogger.Info("P2P: Listening to grpc communication...")
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(internal.NewInterceptor(apiLogger)),
	)
	service.RegisterP2PCommunicationServer(grpcServer, hs)
	return grpcServer.Serve(listener)

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
	// only sends the connected (resolved) peers
	for _, hostPeer := range hs.Host.Peers {
		nodes = append(nodes, hostPeer.GetInfo())
	}
	peers := &model.GetMorePeersResponse{
		Peers: nodes,
	}
	return peers, nil
}

// ResolvePeers looping unresolve peers and adding to (resolve) Peers if get response
func (hs *HostService) ResolvePeers() {
	exceedMaxConnectedPeers := nativeUtil.GetExceedMaxConnectedPeers(hs.Host)
	for _, peer := range hs.Host.GetUnresolvedPeers() {
		// removing the connected peers at random until max - 1
		for i := 0; i < exceedMaxConnectedPeers; i++ {
			peer := nativeUtil.GetAnyPeer(hs.Host)
			if peer != nil {
				delete(hs.Host.Peers, nativeUtil.GetFullAddressPeer(peer))
			}
		}

		go hs.resolvePeer(peer)

		if exceedMaxConnectedPeers > 0 {
			break
		}
	}
}

func (hs *HostService) UpdateConnectedPeers() {
	currentTime := time.Now().UTC()
	for _, peer := range hs.Host.GetPeers() {
		if currentTime.Unix()-peer.GetLastUpdated() >= constant.SecondsToUpdatePeersConnection {
			go hs.resolvePeer(peer)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (hs *HostService) resolvePeer(destPeer *model.Peer) {
	_, err := NewPeerServiceClient().GetPeerInfo(destPeer)
	if err != nil {
		nativeUtil.DisconnectPeer(hs.Host, destPeer)
		return
	}
	destPeer.LastUpdated = time.Now().UTC().Unix()
	updatedHost := nativeUtil.AddToResolvedPeer(hs.Host, destPeer)
	hs.Host = updatedHost

	log.Info(nativeUtil.GetFullAddressPeer(destPeer) + " success")
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (hs *HostService) GetMorePeersHandler() {
	peer := nativeUtil.GetAnyPeer(hs.Host)
	if peer != nil {
		newPeers, err := NewPeerServiceClient().GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error accord %v\n", err)
		}
		hs.Host = nativeUtil.AddToUnresolvedPeers(hs.Host, newPeers.GetPeers())
		hs.SendMyPeers(peer)
	}
}

// SendPeers receives set of peers info from other node and put them into the unresolved peers
func (hs HostService) SendPeers(ctx context.Context, req *model.SendPeersRequest) (*model.Empty, error) {
	// TODO: only accept nodes that are already registered
	nativeUtil.AddToUnresolvedPeers(hs.Host, req.Peers)
	return &model.Empty{}, nil
}
