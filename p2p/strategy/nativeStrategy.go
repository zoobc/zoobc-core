package strategy

import (
	"errors"
	"github.com/zoobc/zoobc-core/p2p/client"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

// NativeStrategy represent data service node as server
type NativeStrategy struct {
	Host                 *model.Host
	ResolvedPeersLock    sync.RWMutex
	UnresolvedPeersLock  sync.RWMutex
	BlacklistedPeersLock sync.RWMutex
	MaxUnresolvedPeers   int32
	MaxResolvedPeers     int32
}

func NewNativeStrategy(
	host *model.Host,
) *NativeStrategy {
	return &NativeStrategy{
		Host:               host,
		MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
		MaxResolvedPeers:   constant.MaxResolvedPeers,
	}
}

/* 	========================================
 *	Resolved Peers Operations
 *	========================================
 */

func (hs *NativeStrategy) GetHostInfo() *model.Node {
	return hs.Host.GetInfo()
}

// GetResolvedPeers returns resolved peers in thread-safe manner
func (hs *NativeStrategy) GetResolvedPeers() map[string]*model.Peer {
	hs.ResolvedPeersLock.RLock()
	defer hs.ResolvedPeersLock.RUnlock()

	var newResolvedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range hs.Host.ResolvedPeers {
		newResolvedPeers[key] = resolvedPeer
	}
	return newResolvedPeers
}

// GetAnyResolvedPeer Get any random resolved peer
func (hs *NativeStrategy) GetAnyResolvedPeer() *model.Peer {
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

// AddToResolvedPeer to add a peer into resolved peer
func (hs *NativeStrategy) AddToResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	hs.ResolvedPeersLock.Lock()
	defer hs.ResolvedPeersLock.Unlock()

	hs.Host.ResolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// RemoveResolvedPeer removes peer from Resolved peer list
func (hs *NativeStrategy) RemoveResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	hs.ResolvedPeersLock.Lock()
	defer hs.ResolvedPeersLock.Unlock()
	delete(hs.Host.ResolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

/* 	========================================
 *	Unresolved Peers Operations
 *	========================================
 */

// GetUnresolvedPeers returns unresolved peers in thread-safe manner
func (hs *NativeStrategy) GetUnresolvedPeers() map[string]*model.Peer {
	hs.UnresolvedPeersLock.RLock()
	defer hs.UnresolvedPeersLock.RUnlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)
	for key, UnresolvedPeer := range hs.Host.UnresolvedPeers {
		newUnresolvedPeers[key] = UnresolvedPeer
	}
	return newUnresolvedPeers
}

// GetAnyUnresolvedPeer Get any unresolved peer
func (hs *NativeStrategy) GetAnyUnresolvedPeer() *model.Peer {
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

// AddToUnresolvedPeer to add a peer into unresolved peer
func (hs *NativeStrategy) AddToUnresolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	hs.UnresolvedPeersLock.Lock()
	defer hs.UnresolvedPeersLock.Unlock()
	hs.Host.UnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
func (hs *NativeStrategy) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error {
	exceedMaxUnresolvedPeers := hs.GetExceedMaxUnresolvedPeers()

	// do not force a peer to go to unresolved list if the list is full n `toForce` is false
	if exceedMaxUnresolvedPeers > 0 && !toForce {
		return errors.New("unresolvedPeers are full")
	}

	unresolvedPeers := hs.GetUnresolvedPeers()
	resolvedPeers := hs.GetResolvedPeers()

	hostAddress := &model.Peer{
		Info: hs.Host.Info,
	}
	for _, node := range newNodes {
		peer := &model.Peer{
			Info: node,
		}
		if unresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			resolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			p2pUtil.GetFullAddressPeer(hostAddress) != p2pUtil.GetFullAddressPeer(peer) {
			for i := int32(0); i < exceedMaxUnresolvedPeers; i++ {
				// removing a peer at random if the UnresolvedPeers has reached max
				peer := hs.GetAnyUnresolvedPeer()
				_ = hs.RemoveUnresolvedPeer(peer)
			}
			_ = hs.AddToUnresolvedPeer(peer)
		}

		if exceedMaxUnresolvedPeers > 0 {
			break
		}
	}
	return nil
}

// RemoveUnresolvedPeer removes peer from unresolved peer list
func (hs *NativeStrategy) RemoveUnresolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	hs.UnresolvedPeersLock.Lock()
	defer hs.UnresolvedPeersLock.Unlock()
	delete(hs.Host.UnresolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

/* 	========================================
 *	Blacklisted Peers Operations
 *	========================================
 */

// GetBlacklistedPeers returns resolved peers in thread-safe manner
func (hs *NativeStrategy) GetBlacklistedPeers() map[string]*model.Peer {
	hs.BlacklistedPeersLock.RLock()
	defer hs.BlacklistedPeersLock.RUnlock()

	var newBlacklistedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range hs.Host.BlacklistedPeers {
		newBlacklistedPeers[key] = resolvedPeer
	}
	return newBlacklistedPeers
}

// AddToBlacklistedPeer to add a peer into resolved peer
func (hs *NativeStrategy) AddToBlacklistedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	hs.BlacklistedPeersLock.Lock()
	defer hs.BlacklistedPeersLock.Unlock()

	hs.Host.BlacklistedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// RemoveBlacklistedPeer removes peer from Blacklisted peer list
func (hs *NativeStrategy) RemoveBlacklistedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	hs.BlacklistedPeersLock.Lock()
	defer hs.BlacklistedPeersLock.Unlock()
	delete(hs.Host.BlacklistedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

// GetAnyKnownPeer Get any known peer
func (hs *NativeStrategy) GetAnyKnownPeer() *model.Peer {
	knownPeers := hs.Host.KnownPeers
	if len(knownPeers) < 1 {
		panic("No well known peer is found")
	}
	randomIdx := int(util.GetSecureRandom()) % len(knownPeers)
	idx := 0
	for _, peer := range knownPeers {
		if idx == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// GetExceedMaxUnresolvedPeers returns number of peers exceeding max number of the unresolved peers
func (hs *NativeStrategy) GetExceedMaxUnresolvedPeers() int32 {
	return int32(len(hs.GetUnresolvedPeers())) - hs.MaxUnresolvedPeers + 1
}

// GetExceedMaxResolvedPeers returns number of peers exceeding max number of the resolved peers
func (hs *NativeStrategy) GetExceedMaxResolvedPeers() int32 {
	return int32(len(hs.GetResolvedPeers())) - hs.MaxResolvedPeers + 1
}

// ResolvePeers looping unresolve peers and adding to (resolve) Peers if get response
func (hs *NativeStrategy) ResolvePeers() {
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

func (hs *NativeStrategy) UpdateResolvedPeers() {
	currentTime := time.Now().UTC()
	for _, peer := range hs.GetResolvedPeers() {
		if currentTime.Unix()-peer.GetLastUpdated() >= constant.SecondsToUpdatePeersConnection {
			go hs.resolvePeer(peer)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (hs *NativeStrategy) resolvePeer(destPeer *model.Peer) {
	_, err := client.NewPeerServiceClient().GetPeerInfo(destPeer)
	if err != nil {
		// TODO: add mechanism to blacklist failing peers
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
func (hs *NativeStrategy) GetMorePeersHandler() (*model.Peer, error) {
	peer := hs.GetAnyResolvedPeer()
	if peer != nil {
		newPeers, err := client.NewPeerServiceClient().GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error accord %v\n", err)
			return nil, err
		}
		err = hs.AddToUnresolvedPeers(newPeers.GetPeers(), true)
		if err != nil {
			log.Warnf("getMorePeers error: %v\n", err)
			return nil, err
		}
		return peer, nil
	}
	return peer, nil
}

func (hs *NativeStrategy) PeerBlacklist(peer *model.Peer, cause string) {
	peer.BlacklistingTime = uint64(time.Now().Unix())
	peer.BlacklistingCause = cause
	_ = hs.AddToBlacklistedPeer(peer)
	_ = hs.RemoveUnresolvedPeer(peer)
	_ = hs.RemoveResolvedPeer(peer)
}

// PeerUnblacklist to update Peer state of peer
func (hs *NativeStrategy) PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	_ = hs.RemoveBlacklistedPeer(peer)
	_ = hs.AddToUnresolvedPeers([]*model.Node{peer.Info}, false)
	return peer
}

// DisconnectPeer moves connected peer to resolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func (hs *NativeStrategy) DisconnectPeer(peer *model.Peer) {
	_ = hs.RemoveResolvedPeer(peer)
	if hs.GetExceedMaxUnresolvedPeers() <= 0 {
		_ = hs.AddToUnresolvedPeer(peer)
	}
}
