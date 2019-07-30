package service

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/p2p/native/util"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// PeerService represent peer service
type PeerServiceClient struct{}

var PeerServiceClientInstance *PeerServiceClient
var once sync.Once

// ClientPeerService to get instance of singleton peer service
func NewPeerServiceClient() *PeerServiceClient {
	once.Do(func() {
		if PeerServiceClientInstance == nil {
			PeerServiceClientInstance = &PeerServiceClient{}
		}
	})
	return PeerServiceClientInstance
}

// GetPeerInfo to get Peer info
func (psc *PeerServiceClient) GetPeerInfo(destPeer *model.Peer) (*model.Node, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
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
func (psc *PeerServiceClient) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
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
func (psc PeerServiceClient) SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendPeers(context.Background(), &model.SendPeersRequest{
		Peers: peersInfo,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// SendBlock send block to selected peer
func (psc PeerServiceClient) SendBlock(destPeer *model.Peer, block *model.Block) (*model.Empty, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendBlock(context.Background(), block)
	if err != nil {
		log.Printf("SendBlock could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// SendTransaction send transaction to selected peer
func (psc PeerServiceClient) SendTransaction(destPeer *model.Peer, block *model.Transaction) (*model.Empty, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendTransaction(context.Background(), block)
	if err != nil {
		log.Printf("SendTransaction could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}
