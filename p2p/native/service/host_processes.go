package service

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

func (hs *HostService) SendMyPeers(peer *model.Peer) {
	peers := hs.GetResolvedPeers()
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

// ResolvePeers looping unresolve peers and adding to (resolve) Peers if get response
func (hs *HostService) ResolvePeers() {
	exceedMaxResolvedPeers := hs.GetExceedMaxResolvedPeers()
	resolvingCount := 0

	for _, peer := range hs.GetUnresolvedPeers() {
		// removing the connected peers at random until max - 1
		for i := 0; i < exceedMaxResolvedPeers; i++ {
			peer := hs.GetAnyResolvedPeer()
			hs.RemoveResolvedPeer(peer)
		}
		go hs.resolvePeer(peer)
		resolvingCount++

		if exceedMaxResolvedPeers > 0 || resolvingCount >= exceedMaxResolvedPeers {
			break
		}
	}
}

func (hs *HostService) UpdateResolvedPeers() {
	currentTime := time.Now().UTC()
	for _, peer := range hs.GetResolvedPeers() {
		if currentTime.Unix()-peer.GetLastUpdated() >= constant.SecondsToUpdatePeersConnection {
			go hs.resolvePeer(peer)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (hs *HostService) resolvePeer(destPeer *model.Peer) {
	_, err := NewPeerServiceClient().GetPeerInfo(destPeer)
	if err != nil {
		hs.DisconnectPeer(destPeer)
		return
	}
	if destPeer != nil {
		destPeer.LastUpdated = time.Now().UTC().Unix()
	}
	hs.RemoveUnresolvedPeer(destPeer)
	hs.AddToResolvedPeer(destPeer)
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (hs *HostService) GetMorePeersHandler() {
	peer := hs.GetAnyResolvedPeer()
	if peer != nil {
		newPeers, err := NewPeerServiceClient().GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error accord %v\n", err)
		}
		hs.AddToUnresolvedPeers(newPeers.GetPeers())
		hs.SendMyPeers(peer)
	}
}

// PeerUnblacklist to update Peer state of peer
func (hs *HostService) PeerUnblacklist(peer *model.Peer) *model.Peer {
	// TODO: handle unblacklisting and blacklisting

	// peer.BlacklistingCause = ""
	// peer.BlacklistingTime = 0
	// if peer.State == model.PeerState_BLACKLISTED {
	// 	peer.State = model.PeerState_NON_CONNECTED
	// }
	return peer
}

// DisconnectPeer moves connected peer to resolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func (hs *HostService) DisconnectPeer(peer *model.Peer) {
	hs.RemoveResolvedPeer(peer)
	if hs.GetExceedMaxUnresolvedPeers() <= 0 {
		hs.AddToUnresolvedPeer(peer)
	}
}
