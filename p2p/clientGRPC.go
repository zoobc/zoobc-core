package p2p

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
)

// GetPeerInfo to get Peer info
func (cs PeerService) GetPeerInfo(destPeer *model.Peer) (*model.Node, error) {
	conn, err := grpc.Dial(util.GetFullAddressPeer(destPeer), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Printf("did not connect: %v\n", err)
	}
	p2pClient := service.NewP2PCommunicationClient(conn)
	// ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(context.Background(), &model.GetPeerInfoRequest{Version: "v1.0.1"})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return nil, err
	}
	// HostService(models.Mainchain{}).Host.Log("got GetPeerInfo response from: " + destPeer.GetFullAddress() + " = " + res.ToString())
	return res, err
}

// GetMorePeers to collect more peers available
func (cs PeerService) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	conn, err := grpc.Dial(util.GetFullAddressPeer(destPeer), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Printf("did not connect: %v\n", err)
	}
	p2pClient := service.NewP2PCommunicationClient(conn)
	// ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(context.Background(), &model.Empty{})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return nil, err
	}
	return res, err
}
