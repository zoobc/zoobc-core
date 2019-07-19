package service

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func (hs HostService) SendMyPeers(peer *model.Peer) {
	peers := hs.Host.GetPeers()
	var myPeersInfo []*model.Node
	myPeersInfo = append(myPeersInfo, hs.Host.GetInfo())
	for _, peer := range peers {
		myPeersInfo = append(myPeersInfo, peer.Info)
	}

	ClientPeerService(chaintype.GetChainType(0)).SendPeers(peer, myPeersInfo)
}
