package client

import (
	"context"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"google.golang.org/grpc"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

type (
	// PeerServiceClientInterface acts as interface for PeerServiceClient
	PeerServiceClientInterface interface {
		GetPeerInfo(destPeer *model.Peer) (*model.Node, error)
		GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error)
		SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error)
		SendBlock(destPeer *model.Peer, senderPublicKey []byte, block *model.Block) (*model.Receipt, error)
		SendTransaction(destPeer *model.Peer, senderPublicKey []byte, transactionBytes []byte) (*model.Receipt, error)

		GetCumulativeDifficulty(*model.Peer, chaintype.ChainType) (*model.GetCumulativeDifficultyResponse, error)
		GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, lastBlockID,
			astMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error)
		GetNextBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, blockID int64, limit uint32) (*model.BlockIdsResponse, error)
		GetNextBlocks(destPeer *model.Peer, chaintype chaintype.ChainType, blockIds []int64, blockID int64) (*model.BlocksData, error)
	}
	// PeerService represent peer service
	PeerServiceClient struct {
		Dialer Dialer
	}
)

// PeerService represent peer service
type Dialer func(destinationPeer *model.Peer) (*grpc.ClientConn, error)

var (
	PeerServiceClientInstance *PeerServiceClient
	once                      sync.Once
)

// ClientPeerService to get instance of singleton peer service
func NewPeerServiceClient() PeerServiceClientInterface {
	once.Do(func() {
		if PeerServiceClientInstance == nil {
			apiLogger, _ := util.InitLogger(".log/", "debugP2PClient.log")
			PeerServiceClientInstance = &PeerServiceClient{
				Dialer: func(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
					conn, err := grpc.Dial(
						p2pUtil.GetFullAddressPeer(destinationPeer),
						grpc.WithInsecure(),
						grpc.WithUnaryInterceptor(interceptor.NewClientInterceptor(apiLogger)),
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
func (psc *PeerServiceClient) SendBlock(destPeer *model.Peer, senderPublicKey []byte, block *model.Block) (*model.Receipt, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendBlock(context.Background(), &model.SendBlockRequest{
		SenderPublicKey: senderPublicKey,
		Block:           block,
	})
	if err != nil {
		return nil, err
	}
	return res, err
}

// SendTransaction send transaction to selected peer
func (psc *PeerServiceClient) SendTransaction(destPeer *model.Peer, senderPublicKey []byte, transactionBytes []byte) (*model.Receipt, error) {
	connection, _ := psc.Dialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)

	res, err := p2pClient.SendTransaction(context.Background(), &model.SendTransactionRequest{
		SenderPublicKey:  senderPublicKey,
		TransactionBytes: transactionBytes,
	})
	if err != nil {
		return nil, err
	}
	return res, err
}

// GetCumulativeDifficulty request the cumulative difficulty status of a node
func (psc PeerServiceClient) GetCumulativeDifficulty(destPeer *model.Peer,
	chaintype chaintype.ChainType) (*model.GetCumulativeDifficultyResponse, error) {
	connection, _ := grpc.Dial(
		p2pUtil.GetFullAddressPeer(destPeer),
		grpc.WithInsecure(),
		// grpc.WithUnaryInterceptor(),
	)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetCumulativeDifficulty(context.Background(), &model.GetCumulativeDifficultyRequest{
		ChainType: chaintype.GetTypeInt(),
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetCommonMilestoneBlockIDs request the blockIds that may act as milestone block
func (psc PeerServiceClient) GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType, lastBlockID,
	lastMilestoneBlockID int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	connection, _ := grpc.Dial(
		p2pUtil.GetFullAddressPeer(destPeer),
		grpc.WithInsecure(),
		// grpc.WithUnaryInterceptor(),
	)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetCommonMilestoneBlockIDs(context.Background(), &model.GetCommonMilestoneBlockIdsRequest{
		ChainType:            chaintype.GetTypeInt(),
		LastBlockID:          lastBlockID,
		LastMilestoneBlockID: lastMilestoneBlockID,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlockIDs request the blockIds of the next blocks requested
func (psc PeerServiceClient) GetNextBlockIDs(destPeer *model.Peer, chaintype chaintype.ChainType,
	blockID int64, limit uint32) (*model.BlockIdsResponse, error) {
	connection, _ := grpc.Dial(
		p2pUtil.GetFullAddressPeer(destPeer),
		grpc.WithInsecure(),
		// grpc.WithUnaryInterceptor(),
	)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetNextBlockIDs(context.Background(), &model.GetNextBlockIdsRequest{
		ChainType: chaintype.GetTypeInt(),
		BlockId:   blockID,
		Limit:     limit,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlocks request the next blocks matching the array of blockIds
func (psc PeerServiceClient) GetNextBlocks(destPeer *model.Peer, chaintype chaintype.ChainType, blockIds []int64,
	blockID int64) (*model.BlocksData, error) {
	connection, _ := grpc.Dial(
		p2pUtil.GetFullAddressPeer(destPeer),
		grpc.WithInsecure(),
	)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetNextBlocks(context.Background(), &model.GetNextBlocksRequest{
		ChainType: chaintype.GetTypeInt(),
		BlockId:   blockID,
		BlockIds:  blockIds,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", p2pUtil.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}
