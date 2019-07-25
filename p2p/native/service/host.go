package service

import (
	"sync"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

// HostService represent data service node as server
type HostService struct {
	Host                 *model.Host
	ResolvedPeersLock    sync.RWMutex
	UnresolvedPeersLock  sync.RWMutex
	BlacklistedPeersLock sync.RWMutex
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

/* 	========================================
 *	Resolved Peers Operations
 *	========================================
 */

// GetResolvedPeers returns resolved peers in thread-safe manner
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

// AddToResolvedPeer to add a peer into resolved peer
func (hs *HostService) AddToResolvedPeer(peer *model.Peer) {
	hs.ResolvedPeersLock.Lock()
	defer hs.ResolvedPeersLock.Unlock()

	hs.Host.ResolvedPeers[nativeUtil.GetFullAddressPeer(peer)] = peer
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

/* 	========================================
 *	Unresolved Peers Operations
 *	========================================
 */

// GetUnresolvedPeers returns unresolved peers in thread-safe manner
func (hs *HostService) GetUnresolvedPeers() map[string]*model.Peer {
	hs.UnresolvedPeersLock.RLock()
	defer hs.UnresolvedPeersLock.RUnlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)
	for key, UnresolvedPeer := range hs.Host.UnresolvedPeers {
		newUnresolvedPeers[key] = UnresolvedPeer
	}
	return newUnresolvedPeers
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

// RemoveUnresolvedPeer removes peer from unresolved peer list
func (hs *HostService) RemoveUnresolvedPeer(peer *model.Peer) {
	if peer == nil {
		return
	}
	hs.UnresolvedPeersLock.Lock()
	defer hs.UnresolvedPeersLock.Unlock()
	delete(hs.Host.UnresolvedPeers, nativeUtil.GetFullAddressPeer(peer))
}

// AddToUnresolvedPeer to add a peer into unresolved peer
func (hs *HostService) AddToUnresolvedPeer(peer *model.Peer) {
	hs.UnresolvedPeersLock.Lock()
	defer hs.UnresolvedPeersLock.Unlock()
	hs.Host.UnresolvedPeers[nativeUtil.GetFullAddressPeer(peer)] = peer
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
func (hs *HostService) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) {
	exceedMaxUnresolvedPeers := hs.GetExceedMaxUnresolvedPeers()

	// do not force a peer to go to unresolved list if the list is full n `toForce` is false
	if exceedMaxUnresolvedPeers > 0 && !toForce {
		return
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

/* 	========================================
 *	Blacklisted Peers Operations
 *	========================================
 */

// GetBlacklistedPeers returns resolved peers in thread-safe manner
func (hs *HostService) GetBlacklistedPeers() map[string]*model.Peer {
	hs.BlacklistedPeersLock.RLock()
	defer hs.BlacklistedPeersLock.RUnlock()

	var newBlacklistedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range hs.Host.BlacklistedPeers {
		newBlacklistedPeers[key] = resolvedPeer
	}
	return newBlacklistedPeers
}

// AddToBlacklistedPeer to add a peer into resolved peer
func (hs *HostService) AddToBlacklistedPeer(peer *model.Peer) {
	hs.BlacklistedPeersLock.Lock()
	defer hs.BlacklistedPeersLock.Unlock()

	hs.Host.BlacklistedPeers[nativeUtil.GetFullAddressPeer(peer)] = peer
}

// RemoveBlacklistedPeer removes peer from Blacklisted peer list
func (hs *HostService) RemoveBlacklistedPeer(peer *model.Peer) {
	if peer == nil {
		return
	}
	hs.BlacklistedPeersLock.Lock()
	defer hs.BlacklistedPeersLock.Unlock()
	delete(hs.Host.BlacklistedPeers, nativeUtil.GetFullAddressPeer(peer))
}

// GetExceedMaxUnresolvedPeers returns number of peers exceeding max number of the unresolved peers
func (hs *HostService) GetExceedMaxUnresolvedPeers() int {
	return len(hs.GetUnresolvedPeers()) - constant.MaxUnresolvedPeers + 1
}

// GetExceedMaxResolvedPeers returns number of peers exceeding max number of the connected peers
func (hs *HostService) GetExceedMaxResolvedPeers() int {
	return len(hs.GetResolvedPeers()) - constant.MaxResolvedPeers + 1
}
