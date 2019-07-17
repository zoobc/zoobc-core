package service

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/p2p/native/util"
	"google.golang.org/grpc"
)

// PeerService represent peer service
type PeerService struct {
	Peer      *model.Peer
	ChainType contract.ChainType
}

// ClientPeerService to get instance of singleton peer service
func ClientPeerService(chaintype contract.ChainType) *PeerService {
	return &PeerService{}
}

// GetPeerInfo to get Peer info
func (cs *PeerService) GetPeerInfo(destPeer *model.Peer) (*model.Node, error) {
	conn, err := grpc.Dial(util.GetFullAddressPeer(destPeer), grpc.WithInsecure())
	if err != nil {
		log.Warnf("could not make dial connection: %v\n", err)
		return nil, err
	}
	defer conn.Close()
	p2pClient := service.NewP2PCommunicationClient(conn)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(context.Background(), &model.GetPeerInfoRequest{Version: "v1.0.1"})
	if err != nil {
		log.Warnf("GetPeerInfo could not greet. %v\n", err)
		return nil, err
	}

	log.Infof("got GetPeerInfo response from %v = %v\n", util.GetFullAddressPeer(destPeer), res)
	return res, err
}

// GetMorePeers to collect more peers available
func (cs *PeerService) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	conn, err := grpc.Dial(util.GetFullAddressPeer(destPeer), grpc.WithInsecure())
	if err != nil {
		log.Warnf("could not make dial connection. %v\n", err)
		return nil, err
	}
	defer conn.Close()
	p2pClient := service.NewP2PCommunicationClient(conn)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(context.Background(), &model.Empty{})
	if err != nil {
		log.Warnf("could not greet. %v\n", err)
		return nil, err
	}
	return res, err
}

// SendPeers sends set of peers to other node (to populate the network)
func (cs PeerService) SendPeers(destPeer model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	conn, err := grpc.Dial(util.GetFullAddressPeer(&destPeer), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		log.Printf("did not connect: %v\n", err)
	}
	p2pClient := service.NewP2PCommunicationClient(conn)
	// ctx := cs.buildContext()
	res, err := p2pClient.SendPeers(context.Background(), &model.SendPeersRequest{
		Peers: peersInfo,
	})
	if err != nil {
		log.Printf("could not greet: %v\n", err)
		return nil, err
	}
	return res, err
}
