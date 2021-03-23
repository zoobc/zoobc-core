// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package strategy

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/p2p/client"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

// NativeStrategy represent data service node as server
type NativeStrategy struct {
	Host                 *model.Host
	PeerServiceClient    client.PeerServiceClientInterface
	ResolvedPeersLock    sync.RWMutex
	UnresolvedPeersLock  sync.RWMutex
	BlacklistedPeersLock sync.RWMutex
	MaxUnresolvedPeers   int32
	MaxResolvedPeers     int32
	Logger               *log.Logger
}

func NewNativeStrategy(
	host *model.Host,
	peerServiceClient client.PeerServiceClientInterface,
	logger *log.Logger,
) *NativeStrategy {
	return &NativeStrategy{
		Host:               host,
		PeerServiceClient:  peerServiceClient,
		MaxUnresolvedPeers: constant.MaxUnresolvedPeers,
		MaxResolvedPeers:   constant.MaxResolvedPeers,
		Logger:             logger,
	}
}

// Start method to start threads which mean goroutines for NativeStrategy
func (ns *NativeStrategy) Start() {
	// start p2p process threads
	go ns.ResolvePeersThread()
	go ns.GetMorePeersThread()
	go ns.UpdateBlacklistedStatusThread()
}

// ============================================
// Thread resolving peers
// ============================================

// ResolvePeersThread to periodically try get response from peers in UnresolvedPeer list
func (ns *NativeStrategy) ResolvePeersThread() {
	go ns.ResolvePeers()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go ns.ResolvePeers()
			go ns.UpdateResolvedPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// ResolvePeers looping unresolved peers and adding to (resolve) Peers if get response
func (ns *NativeStrategy) ResolvePeers() {
	exceedMaxResolvedPeers := ns.GetExceedMaxResolvedPeers()
	resolvingCount := int32(0)

	for i := int32(0); i < exceedMaxResolvedPeers; i++ {
		peer := ns.GetAnyResolvedPeer()
		ns.DisconnectPeer(peer)
	}

	for _, peer := range ns.GetUnresolvedPeers() {
		go ns.resolvePeer(peer)
		resolvingCount++

		if exceedMaxResolvedPeers > 0 || resolvingCount >= exceedMaxResolvedPeers {
			break
		}
	}
}

func (ns *NativeStrategy) UpdateResolvedPeers() {
	currentTime := time.Now().UTC()
	for _, peer := range ns.GetResolvedPeers() {
		if currentTime.Unix()-peer.GetResolvingTime() >= constant.SecondsToUpdatePeersConnection {
			go ns.resolvePeer(peer)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (ns *NativeStrategy) resolvePeer(destPeer *model.Peer) {
	_, err := ns.PeerServiceClient.GetPeerInfo(destPeer)
	if err != nil {
		// TODO: add mechanism to blacklist failing peers
		ns.DisconnectPeer(destPeer)
		return
	}
	if destPeer != nil {
		destPeer.ResolvingTime = time.Now().UTC().Unix()
	}
	if err := ns.RemoveUnresolvedPeer(destPeer); err != nil {
		ns.Logger.Warn(err.Error())
	}
	if err := ns.AddToResolvedPeer(destPeer); err != nil {
		ns.Logger.Warn(err.Error())
	}
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (ns *NativeStrategy) GetMorePeersHandler() (*model.Peer, error) {
	peer := ns.GetAnyResolvedPeer()
	if peer != nil {
		newPeers, err := ns.PeerServiceClient.GetMorePeers(peer)
		if err != nil {
			ns.Logger.Infof("getMorePeers Error accord %v\n", err)
			return nil, err
		}
		err = ns.AddToUnresolvedPeers(newPeers.GetPeers(), true)
		if err != nil {
			ns.Logger.Infof("getMorePeers error: %v\n", err)
			return nil, err
		}
		return peer, nil
	}
	return peer, nil
}

// ===========================================
// 	Thread for gettingMorePeers
// ===========================================

// GetMorePeersThread to periodically request more peers from another node in Peers list
func (ns *NativeStrategy) GetMorePeersThread() {
	syncPeers := func() {
		var (
			nodes []*model.Node
			peer  *model.Peer
			err   error
		)

		peer, err = ns.GetMorePeersHandler()
		if err != nil {
			ns.Logger.Warn(err.Error())
			return
		}

		myResolvedPeers := ns.GetResolvedPeers()
		for _, p := range myResolvedPeers {
			nodes = append(nodes, p.Info)
		}
		if peer == nil {
			return
		}
		nodes = append(nodes, ns.Host.GetInfo())
		_, _ = ns.PeerServiceClient.SendPeers(
			peer,
			nodes,
		)
	}

	go syncPeers()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go syncPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// UpdateBlacklistedStatusThread to periodically check blacklisting time of black listed peer,
// every 60sec if there are blacklisted peers to unblacklist
func (ns *NativeStrategy) UpdateBlacklistedStatusThread() {
	ticker := time.NewTicker(time.Duration(constant.UpdateBlacklistedStatusGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint64(time.Now().Unix())
				for _, p := range ns.Host.GetBlacklistedPeers() {
					if p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+constant.BlacklistingPeriod <= curTime {
						_ = ns.PeerUnblacklist(p)
					}
				}
				break
			case <-sigs:
				ticker.Stop()
				return
			}
		}
	}()
}

func (ns *NativeStrategy) GetPriorityPeersByFullAddress(priorityPeers map[string]*model.Peer) map[string]*model.Peer {
	var priorityPeersByAddr = make(map[string]*model.Peer)
	for _, pp := range priorityPeers {
		if pp.GetInfo().Address != "" && pp.GetInfo().Port != 0 {
			priorityPeersByAddr[p2pUtil.GetFullAddress(pp.GetInfo())] = pp
		}
	}
	return priorityPeersByAddr
}

// GetPriorityPeers returns resolved peers in thread-safe manner
func (ns *NativeStrategy) GetPriorityPeers() map[string]*model.Peer {
	return make(map[string]*model.Peer)
}

/* 	========================================
 *	Resolved Peers Operations
 *	========================================
 */

func (ns *NativeStrategy) GetHostInfo() *model.Node {
	return ns.Host.GetInfo()
}

// GetResolvedPeers returns resolved peers in thread-safe manner
func (ns *NativeStrategy) GetResolvedPeers() map[string]*model.Peer {
	ns.ResolvedPeersLock.RLock()
	defer ns.ResolvedPeersLock.RUnlock()

	var newResolvedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range ns.Host.ResolvedPeers {
		newResolvedPeers[key] = resolvedPeer
	}
	return newResolvedPeers
}

// GetAnyResolvedPeer Get any random resolved peer
func (ns *NativeStrategy) GetAnyResolvedPeer() *model.Peer {
	resolvedPeers := ns.GetResolvedPeers()
	if len(resolvedPeers) < 1 {
		return nil
	}
	randomIdx := uint64(util.GetSecurePositiveRandom())
	if randomIdx != 0 {
		randomIdx %= uint64(len(resolvedPeers))
	}
	idx := 0
	for _, peer := range resolvedPeers {
		if uint64(idx) == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// AddToResolvedPeer to add a peer into resolved peer
func (ns *NativeStrategy) AddToResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ns.ResolvedPeersLock.Lock()
	defer ns.ResolvedPeersLock.Unlock()

	ns.Host.ResolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// RemoveResolvedPeer removes peer from Resolved peer list
func (ns *NativeStrategy) RemoveResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ns.ResolvedPeersLock.Lock()
	defer ns.ResolvedPeersLock.Unlock()
	delete(ns.Host.ResolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

/* 	========================================
 *	Unresolved Peers Operations
 *	========================================
 */

// GetUnresolvedPeers returns unresolved peers in thread-safe manner
func (ns *NativeStrategy) GetUnresolvedPeers() map[string]*model.Peer {
	ns.UnresolvedPeersLock.RLock()
	defer ns.UnresolvedPeersLock.RUnlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)
	for key, UnresolvedPeer := range ns.Host.UnresolvedPeers {
		newUnresolvedPeers[key] = UnresolvedPeer
	}
	return newUnresolvedPeers
}

// GetAnyUnresolvedPeer Get any unresolved peer
func (ns *NativeStrategy) GetAnyUnresolvedPeer() *model.Peer {
	unresolvedPeers := ns.GetUnresolvedPeers()
	if len(unresolvedPeers) < 1 {
		return nil
	}
	randomIdx := uint64(util.GetSecurePositiveRandom()) % uint64(len(unresolvedPeers))
	idx := 0
	for _, peer := range unresolvedPeers {
		if uint64(idx) == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// AddToUnresolvedPeer to add a peer into unresolved peer
func (ns *NativeStrategy) AddToUnresolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ns.UnresolvedPeersLock.Lock()
	defer ns.UnresolvedPeersLock.Unlock()
	ns.Host.UnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
// toForce: if it's true, when the unresolvedPeers list is full, we will kick another one inside
//			(by choosing 1 random node)
func (ns *NativeStrategy) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error {
	exceedMaxUnresolvedPeers := ns.GetExceedMaxUnresolvedPeers()

	// do not force a peer to go to unresolved list if the list is full n `toForce` is false
	if exceedMaxUnresolvedPeers > 1 && !toForce {
		return errors.New("unresolvedPeers are full")
	}

	var addedPeers int32
	unresolvedPeers := ns.GetUnresolvedPeers()
	resolvedPeers := ns.GetResolvedPeers()

	hostAddress := &model.Peer{
		Info: ns.Host.Info,
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
				peer := ns.GetAnyUnresolvedPeer()
				if err := ns.RemoveUnresolvedPeer(peer); err != nil {
					ns.Logger.Warn(err.Error())
				}
			}
			if err := ns.AddToUnresolvedPeer(peer); err != nil {
				ns.Logger.Warn(err.Error())
			}
			addedPeers++
		}

		if exceedMaxUnresolvedPeers+addedPeers > 1 {
			break
		}
	}
	return nil
}

// RemoveUnresolvedPeer removes peer from unresolved peer list
func (ns *NativeStrategy) RemoveUnresolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ns.UnresolvedPeersLock.Lock()
	defer ns.UnresolvedPeersLock.Unlock()
	delete(ns.Host.UnresolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

/* 	========================================
 *	Blacklisted Peers Operations
 *	========================================
 */

// GetBlacklistedPeers returns resolved peers in thread-safe manner
func (ns *NativeStrategy) GetBlacklistedPeers() map[string]*model.Peer {
	ns.BlacklistedPeersLock.RLock()
	defer ns.BlacklistedPeersLock.RUnlock()

	var newBlacklistedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range ns.Host.BlacklistedPeers {
		newBlacklistedPeers[key] = resolvedPeer
	}
	return newBlacklistedPeers
}

// AddToBlacklistPeer to add a peer into resolved peer
func (ns *NativeStrategy) AddToBlacklistPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ns.BlacklistedPeersLock.Lock()
	defer ns.BlacklistedPeersLock.Unlock()

	ns.Host.BlacklistedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// RemoveBlacklistedPeer removes peer from Blacklisted peer list
func (ns *NativeStrategy) RemoveBlacklistedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ns.BlacklistedPeersLock.Lock()
	defer ns.BlacklistedPeersLock.Unlock()
	delete(ns.Host.BlacklistedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

// ======================================================
// Exposed Functions
// ======================================================

// GetAnyKnownPeer Get any known peer
func (ns *NativeStrategy) GetAnyKnownPeer() *model.Peer {
	knownPeers := ns.Host.KnownPeers
	if len(knownPeers) < 1 {
		panic("No well known peer is found")
	}
	randomIdx := uint64(util.GetSecurePositiveRandom()) % uint64(len(knownPeers))
	idx := 0
	for _, peer := range knownPeers {
		if uint64(idx) == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// GetExceedMaxUnresolvedPeers returns number of peers exceeding max number of the unresolved peers
func (ns *NativeStrategy) GetExceedMaxUnresolvedPeers() int32 {
	return int32(len(ns.GetUnresolvedPeers())) - ns.MaxUnresolvedPeers + 1
}

// GetExceedMaxResolvedPeers returns number of peers exceeding max number of the resolved peers
func (ns *NativeStrategy) GetExceedMaxResolvedPeers() int32 {
	return int32(len(ns.GetResolvedPeers())) - ns.MaxResolvedPeers + 1
}

// PeerBlacklist process to add blacklisted peer
func (ns *NativeStrategy) PeerBlacklist(peer *model.Peer, cause string) error {
	peer.BlacklistingTime = uint64(time.Now().Unix())
	if err := ns.AddToBlacklistPeer(peer); err != nil {
		ns.Logger.Warn(err.Error())
		return err
	}
	if err := ns.RemoveUnresolvedPeer(peer); err != nil {
		ns.Logger.Warn(err.Error())
		return err
	}
	if err := ns.RemoveResolvedPeer(peer); err != nil {
		ns.Logger.Warn(err.Error())
		return err
	}
	return nil
}

// PeerUnblacklist to update Peer state of peer
func (ns *NativeStrategy) PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	if err := ns.RemoveBlacklistedPeer(peer); err != nil {
		ns.Logger.Warn(err.Error())
	}
	if err := ns.AddToUnresolvedPeers([]*model.Node{peer.Info}, false); err != nil {
		ns.Logger.Warn(err.Error())
	}
	return peer
}

// DisconnectPeer moves connected peer to resolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func (ns *NativeStrategy) DisconnectPeer(peer *model.Peer) {
	_ = ns.RemoveResolvedPeer(peer)
	if ns.GetExceedMaxUnresolvedPeers() <= 0 {
		if err := ns.AddToUnresolvedPeer(peer); err != nil {
			ns.Logger.Warn(err.Error())
		}
	}
}

func (ns *NativeStrategy) ValidateRequest(context.Context) bool {
	return true
}

// TODO implement this method
// GetNodeAddressesInfo request a list of node addresses from peers
func (ns *NativeStrategy) SyncNodeAddressInfoTable(nodeRegistrations []*model.NodeRegistration) (map[int64]*model.NodeAddressInfo, error) {
	return nil, nil
}

// TODO implement this method
// ReceiveNodeAddressInfo receive a node address info from a peer (server side of SendNodeAddressInfo client api call)
func (ns *NativeStrategy) ReceiveNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error {
	return nil
}

// TODO implement this method
// UpdateOwnNodeAddressInfo check if nodeAddress in db must be updated and broadcast the new address
func (ns *NativeStrategy) UpdateOwnNodeAddressInfo(nodeAddress string, port uint32, nodeSecretPhrase string) error {
	return nil
}

// TODO implement this method
// GenerateProofOfOrigin generate a proof of origin message from a challenge request and sign it
func (ns *NativeStrategy) GenerateProofOfOrigin(
	challenge []byte,
	timestamp int64,
	nodeSecretPhrase string,
) *model.ProofOfOrigin {
	return nil
}
