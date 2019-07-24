package service

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// HostService represent data service node as server
type HostService struct {
	Host                *model.Host
	ResolvedPeersLock   sync.RWMutex
	UnresolvedPeersLock sync.RWMutex
}

var hostServiceInstance *HostService

func NewHostService(host *model.Host) *HostService {
	if hostServiceInstance == nil {
		hostServiceInstance = &HostService{Host: host}
	}
	return hostServiceInstance
}

func GetHostService() *HostService {
	if hostServiceInstance == nil {
		panic("The host instance is never initiated yet")
	}
	return hostServiceInstance
}

func (hs *HostService) GetResolvedPeers() map[string]*model.Peer {
	hs.ResolvedPeersLock.RLock()
	defer hs.ResolvedPeersLock.RUnlock()

	var newResolvedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range hs.Host.ResolvedPeers {
		newResolvedPeers[key] = resolvedPeer
	}
	return newResolvedPeers
}

// GetAnyResolvedPeer Get any random connected peer
func (hs *HostService) GetAnyResolvedPeer() *model.Peer {
	resolvedPeers := hs.GetResolvedPeers()
	if len(resolvedPeers) < 1 {
		return nil
	}
	randomIdx := int(util.GetSecureRandom())
	if randomIdx != 0 {
		randomIdx %= len(resolvedPeers)
	}
	idx := 0
	for _, peer := range resolvedPeers {
		if idx == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// GetAnyUnresolvedPeer Get any unresolved peer
func (hs *HostService) GetAnyUnresolvedPeer() *model.Peer {
	unresolvedPeers := hs.GetUnresolvedPeers()
	if len(unresolvedPeers) < 1 {
		return nil
	}
	randomIdx := int(util.GetSecureRandom()) % len(unresolvedPeers)
	idx := 0
	for _, peer := range unresolvedPeers {
		if idx == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// RemoveResolvedPeer removes peer from Resolved peer list
func (hs *HostService) RemoveResolvedPeer(peer *model.Peer) {
	if peer == nil {
		return
	}
	hs.ResolvedPeersLock.Lock()
	defer hs.ResolvedPeersLock.Unlock()
	delete(hs.Host.ResolvedPeers, nativeUtil.GetFullAddressPeer(peer))
}

func (hs *HostService) GetUnresolvedPeers() map[string]*model.Peer {
	hs.UnresolvedPeersLock.RLock()
	defer hs.UnresolvedPeersLock.RUnlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)
	for key, UnresolvedPeer := range hs.Host.UnresolvedPeers {
		newUnresolvedPeers[key] = UnresolvedPeer
	}
	return newUnresolvedPeers
}

// RemoveUnresolvedPeer removes peer from unresolved peer list
func (hs *HostService) RemoveUnresolvedPeer(peer *model.Peer) {
	if peer == nil {
		return
	}
	hs.UnresolvedPeersLock.Lock()
	defer hs.UnresolvedPeersLock.Unlock()
	delete(hs.Host.UnresolvedPeers, nativeUtil.GetFullAddressPeer(peer))
}

func (hs *HostService) AddToUnresolvedPeer(peer *model.Peer) {
	hs.UnresolvedPeersLock.Lock()
	defer hs.UnresolvedPeersLock.Unlock()
	hs.Host.UnresolvedPeers[nativeUtil.GetFullAddressPeer(peer)] = peer
}

// AddToResolvedPeer to move unresolved peer into resolved peer
func (hs *HostService) AddToResolvedPeer(peer *model.Peer) {
	hs.ResolvedPeersLock.Lock()
	defer hs.ResolvedPeersLock.Unlock()

	hs.Host.ResolvedPeers[nativeUtil.GetFullAddressPeer(peer)] = peer
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
func (hs *HostService) AddToUnresolvedPeers(newNodes []*model.Node) {
	unresolvedPeers := hs.GetUnresolvedPeers()
	resolvedPeers := hs.GetResolvedPeers()

	exceedMaxUnresolvedPeers := nativeUtil.GetExceedMaxUnresolvedPeers(unresolvedPeers)
	hostAddress := &model.Peer{
		Info: hs.Host.Info,
	}
	for _, node := range newNodes {
		peer := &model.Peer{
			Info: node,
		}
		if unresolvedPeers[nativeUtil.GetFullAddressPeer(peer)] == nil &&
			resolvedPeers[nativeUtil.GetFullAddressPeer(peer)] == nil &&
			nativeUtil.GetFullAddressPeer(hostAddress) != nativeUtil.GetFullAddressPeer(peer) {
			for i := 0; i < exceedMaxUnresolvedPeers; i++ {
				// removing a peer at random if the UnresolvedPeers has reached max
				peer := hs.GetAnyUnresolvedPeer()
				hs.RemoveUnresolvedPeer(peer)
			}
			hs.AddToUnresolvedPeer(peer)
		}

		if exceedMaxUnresolvedPeers > 0 {
			break
		}
	}
}

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
	resolvedPeers := hs.GetResolvedPeers()
	exceedMaxResolvedPeers := nativeUtil.GetExceedMaxResolvedPeers(resolvedPeers)
	for _, peer := range hs.GetUnresolvedPeers() {
		// removing the connected peers at random until max - 1
		for i := 0; i < exceedMaxResolvedPeers; i++ {
			peer := hs.GetAnyResolvedPeer()
			hs.RemoveResolvedPeer(peer)
		}
		if exceedMaxResolvedPeers < 1 {
			go hs.resolvePeer(peer)
		}

		if exceedMaxResolvedPeers > 0 {
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
	unresolvedPeers := hs.GetUnresolvedPeers()
	if unresolvedPeers[nativeUtil.GetFullAddressPeer(destPeer)] != nil {
		hs.RemoveUnresolvedPeer(destPeer)
	}
	hs.AddToResolvedPeer(destPeer)

	log.Info(nativeUtil.GetFullAddressPeer(destPeer) + " success")
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
	unresolvedPeers := hs.GetUnresolvedPeers()
	hs.RemoveResolvedPeer(peer)
	if nativeUtil.GetExceedMaxUnresolvedPeers(unresolvedPeers) <= 0 {
		hs.UnresolvedPeersLock.Lock()
		defer hs.UnresolvedPeersLock.Unlock()
		hs.AddToUnresolvedPeer(peer)
	}
}
