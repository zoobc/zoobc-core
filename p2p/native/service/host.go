package service

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func (hs HostService) sendMyPeers(peer model.Peer) {
	peers := hs.Host.GetPeers()
	var myPeersInfo []*model.Node
	for _, peer := range peers {
		myPeersInfo = append(myPeersInfo, peer.Info)
	}

	ClientPeerService(chaintype.GetChainType(0)).SendPeers(peer, myPeersInfo)
}

func (hs HostService) AddToUnresolvedPeers(peersInfo []*model.Node) error {
	var err error
	for _, peerInfo := range peersInfo {
		err = hs.AddPeerToUnresolvedPeers(peerInfo)
		if err != nil {
			return err
		}

		if hs.IsUnresolvedPeersFull() {
			err = hs.DeleteUnresolvedPeer(1)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (hs HostService) IsUnresolvedPeersFull() bool {
	return true
}

func (hs HostService) DeleteUnresolvedPeer(n int32) error {
	return nil
}

func (hs HostService) AddPeerToUnresolvedPeers(peerInfo *model.Node) error {
	return nil
}
