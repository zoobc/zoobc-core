package service

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// SendMyPeers sends resolved peers of a host including the address of the host itself
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
	resolvingCount := int32(0)

	for i := int32(0); i < exceedMaxResolvedPeers; i++ {
		peer := hs.GetAnyResolvedPeer()
		hs.DisconnectPeer(peer)
	}

	for _, peer := range hs.GetUnresolvedPeers() {
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
	_ = hs.RemoveUnresolvedPeer(destPeer)
	_ = hs.AddToResolvedPeer(destPeer)
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (hs *HostService) GetMorePeersHandler() {
	peer := hs.GetAnyResolvedPeer()
	if peer != nil {
		newPeers, err := NewPeerServiceClient().GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error accord %v\n", err)
		}
		_ = hs.AddToUnresolvedPeers(newPeers.GetPeers(), true)
		hs.SendMyPeers(peer)
	}
}

func (hs *HostService) PeerBlacklist(peer *model.Peer, cause string) {
	peer.BlacklistingTime = uint64(time.Now().Unix())
	peer.BlacklistingCause = cause
	_ = hs.AddToBlacklistedPeer(peer)
	_ = hs.RemoveUnresolvedPeer(peer)
	_ = hs.RemoveResolvedPeer(peer)
}

// PeerUnblacklist to update Peer state of peer
func (hs *HostService) PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	_ = hs.RemoveBlacklistedPeer(peer)
	_ = hs.AddToUnresolvedPeers([]*model.Node{peer.Info}, false)
	return peer
}

// DisconnectPeer moves connected peer to resolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func (hs *HostService) DisconnectPeer(peer *model.Peer) {
	_ = hs.RemoveResolvedPeer(peer)
	if hs.GetExceedMaxUnresolvedPeers() <= 0 {
		_ = hs.AddToUnresolvedPeer(peer)
	}
}
