package service

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

// PeerService represent peer service
type PeerService struct {
	Peer      *model.Peer
	ChainType contract.ChainType
}

// ClientPeerService to get instance of singleton peer service
func ClientPeerService(chaintype contract.ChainType) *PeerService {
	return &PeerService{
		ChainType: chaintype,
	}
}

// GetPeerInfo to get Peer info
func (cs *PeerService) GetPeerInfo(connection *grpc.ClientConn) (*model.Node, error) {
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetPeerInfo(context.Background(), &model.GetPeerInfoRequest{Version: "v1,.0.1"})
	if err != nil {
		log.Warnf("GetPeerInfo could not greet. %v\n", err)
		return nil, err
	}

	return res, err
}

// GetMorePeers to collect more peers available
func (cs *PeerService) GetMorePeers(connection *grpc.ClientConn) (*model.GetMorePeersResponse, error) {
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	// context still not use ctx := cs.buildContext()
	res, err := p2pClient.GetMorePeers(context.Background(), &model.Empty{})
	if err != nil {
		log.Warnf("could not greet. %v\n", err)
		return nil, err
	}
	return res, err
}
