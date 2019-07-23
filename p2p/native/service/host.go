package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

func (hs HostService) SendMyPeers(peer *model.Peer) {
	peers := hs.Host.GetPeers()
	var myPeersInfo []*model.Node
	myPeersInfo = append(myPeersInfo, hs.Host.GetInfo())
	for _, peer := range peers {
		myPeersInfo = append(myPeersInfo, peer.Info)
	}

	_, err := NewPeerServiceClient().SendPeers(peer, myPeersInfo)
	if err != nil {
		log.Printf("failed to send the host peers to %s: %v", nativeUtil.GetFullAddressPeer(peer), err)
	}
}
