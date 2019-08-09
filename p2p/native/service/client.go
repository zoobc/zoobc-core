package service

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/p2p/native/util"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

type (
	// PeerServiceClientInterface acts as interface for PeerServiceClient
	PeerServiceClientInterface interface {
		GetPeerInfo(destPeer *model.Peer) (*model.Node, error)
		GetMorePeers(destPeer *model.Peer) (*model.GetMorePeersResponse, error)
		SendPeers(destPeer *model.Peer, peersInfo []*model.Node) (*model.Empty, error)

		GetCumulativeDifficulty(*model.Peer, contract.ChainType) (*model.GetCumulativeDifficultyResponse, error)
		GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, lastBlockId, lastMilestoneBlockId int64) (*model.GetCommonMilestoneBlockIdsResponse, error)
		GetNextBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, blockId int64, limit uint32) (*model.BlockIdsResponse, error)
		GetNextBlocks(destPeer *model.Peer, chaintype contract.ChainType, blockIds []int64, blockId int64) (*model.BlocksData, error)
	}
	// PeerService represent peer service
	PeerServiceClient struct{}
)

var PeerServiceClientInstance *PeerServiceClient
var once sync.Once

// ClientPeerService to get instance of singleton peer service
func NewPeerServiceClient() PeerServiceClientInterface {
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

// GetCumulativeDifficulty request the cumulative difficulty status of a node
func (psc PeerServiceClient) GetCumulativeDifficulty(destPeer *model.Peer, chaintype contract.ChainType) (*model.GetCumulativeDifficultyResponse, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetCumulativeDifficulty(context.Background(), &model.GetCumulativeDifficultyRequest{
		ChainType: chaintype.GetTypeInt(),
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetCommonMilestoneBlockIDs request the blockIds that may act as milestone block
func (psc PeerServiceClient) GetCommonMilestoneBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, lastBlockId, lastMilestoneBlockId int64) (*model.GetCommonMilestoneBlockIdsResponse, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetCommonMilestoneBlockIDs(context.Background(), &model.GetCommonMilestoneBlockIdsRequest{
		ChainType:            chaintype.GetTypeInt(),
		LastBlockId:          lastBlockId,
		LastMilestoneBlockId: lastMilestoneBlockId,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlockIDs request the blockIds of the next blocks requested
func (psc PeerServiceClient) GetNextBlockIDs(destPeer *model.Peer, chaintype contract.ChainType, blockId int64, limit uint32) (*model.BlockIdsResponse, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetNextBlockIDs(context.Background(), &model.GetNextBlockIdsRequest{
		ChainType: chaintype.GetTypeInt(),
		BlockId:   blockId,
		Limit:     limit,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}

// GetNextBlocks request the next blocks matching the array of blockIds
func (psc PeerServiceClient) GetNextBlocks(destPeer *model.Peer, chaintype contract.ChainType, blockIds []int64, blockId int64) (*model.BlocksData, error) {
	connection, _ := nativeUtil.GrpcDialer(destPeer)
	defer connection.Close()
	p2pClient := service.NewP2PCommunicationClient(connection)
	res, err := p2pClient.GetNextBlocks(context.Background(), &model.GetNextBlocksRequest{
		ChainType: chaintype.GetTypeInt(),
		BlockId:   blockId,
		BlockIds:  blockIds,
	})
	if err != nil {
		log.Printf("could not greet %v: %v\n", util.GetFullAddressPeer(destPeer), err)
		return nil, err
	}
	return res, err
}
