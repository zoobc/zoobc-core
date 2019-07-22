package service

import (
	"context"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/p2p/native/util"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// PeerService represent peer service
type PeerServiceClient struct{}

// ClientPeerService to get instance of singleton peer service
func NewPeerServiceClient() *PeerServiceClient {
	return &PeerServiceClient{}
}

// GetPeerInfo to get Peer info
func (psc *PeerServiceClient) GetPeerInfo(destinationPeer *model.Peer) (*model.Node, error) {
	connection, _ := nativeUtil.GrpcDialer(destinationPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(context.Background(), &model.GetPeerInfoRequest{Version: "v1,.0.1"})
	if err != nil {
		log.Warnf("GetPeerInfo could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}

	return res, err
}

// GetMorePeers to collect more peers available
func (psc *PeerServiceClient) GetMorePeers(destinationPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	connection, _ := nativeUtil.GrpcDialer(destinationPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(context.Background(), &model.Empty{})
	if err != nil {
		log.Warnf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// SendPeers sends set of peers to other node (to populate the network)
func (cs PeerService) SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	conn, err := grpc.Dial(util.GetFullAddressPeer(destPeer), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Printf("did not connect %v: %v\n", util.GetFullAddressPeer(destPeer), err)
	}
	p2pClient := service.NewP2PCommunicationClient(conn)
	// ctx := cs.buildContext()
	res, err := p2pClient.SendPeers(context.Background(), &model.SendPeersRequest{
		Peers: peersInfo,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}
