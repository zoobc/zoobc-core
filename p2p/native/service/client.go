package service

import (
	"context"
	"sync"

	"google.golang.org/grpc"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// PeerService represent peer service
type Dialer func(destinationPeer *model.Peer) (*grpc.ClientConn, error)

type PeerServiceClient struct {
	Dialer Dialer
}

var PeerServiceClientInstance *PeerServiceClient
var once sync.Once

// ClientPeerService to get instance of singleton peer service
func NewPeerServiceClient() *PeerServiceClient {
	once.Do(func() {
		if PeerServiceClientInstance == nil {
			apiLogger, _ = util.InitLogger(".log/", "debugP2PClient.log")
			PeerServiceClientInstance = &PeerServiceClient{
				Dialer: func(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
					conn, err := grpc.Dial(
						nativeUtil.GetFullAddressPeer(destinationPeer),
						grpc.WithInsecure(),
						grpc.WithUnaryInterceptor(util.NewClientInterceptor(apiLogger)),
					)
					if err != nil {
						return nil, err
					}
					return conn, nil
				},
			}
		}
	})
	return PeerServiceClientInstance
}

// GetPeerInfo to get Peer info
func (psc *PeerServiceClient) GetPeerInfo(destPeer *model.Peer) (*model.Node, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(context.Background(), &model.GetPeerInfoRequest{Version: "v1,.0.1"})
	if err != nil {
		return nil, err
	}
	return res, err
}

// GetMorePeers to collect more peers available
func (psc *PeerServiceClient) GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(context.Background(), &model.Empty{})
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendPeers sends set of peers to other node (to populate the network)
func (psc *PeerServiceClient) SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendPeers(context.Background(), &model.SendPeersRequest{
		Peers: peersInfo,
	})
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendBlock send block to selected peer
func (psc *PeerServiceClient) SendBlock(destPeer *model.Peer, block *model.Block) (*model.Empty, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendBlock(context.Background(), block)
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendTransaction send transaction to selected peer
func (psc *PeerServiceClient) SendTransaction(destPeer *model.Peer, transactionBytes []byte) (*model.Empty, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendTransaction(context.Background(), &model.SendTransactionRequest{
		TransactionBytes: transactionBytes,
	})
	if err != nil {
		return nil, err
	}
	return res, err
}
