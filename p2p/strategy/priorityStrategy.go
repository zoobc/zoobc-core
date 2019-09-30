package strategy

import (
	"errors"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/client"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

type PriorityPeers struct {
	// MainPriority, a list peers should connect
	MainPriority map[string]*model.Peer
	// MutualPriority, a list peers that a host node become main priority peers in other nodes
	MutualPriority map[string]*model.Peer
}

// PriorityStrategy represent data service node as server
type PriorityStrategy struct {
	Host                  *model.Host
	PeerServiceClient     client.PeerServiceClientInterface
	QueryExecutor         query.ExecutorInterface
	NodeRegistrationQuery query.NodeRegistrationQueryInterface
	PriorityPeers         PriorityPeers
	PriorityPeersLock     sync.RWMutex
	ResolvedPeersLock     sync.RWMutex
	UnresolvedPeersLock   sync.RWMutex
	BlacklistedPeersLock  sync.RWMutex
	MaxUnresolvedPeers    int32
	MaxResolvedPeers      int32
}

func NewPriorityStrategy(
	host *model.Host,
	peerServiceClient client.PeerServiceClientInterface,
	queryExecutor query.ExecutorInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
) *PriorityStrategy {
	return &PriorityStrategy{
		Host:              host,
		PeerServiceClient: peerServiceClient,
		QueryExecutor:     queryExecutor,
		PriorityPeers: PriorityPeers{
			MainPriority:   make(map[string]*model.Peer),
			MutualPriority: make(map[string]*model.Peer),
		},
		NodeRegistrationQuery: nodeRegistrationQuery,
		MaxUnresolvedPeers:    constant.MaxUnresolvedPeers,
		MaxResolvedPeers:      constant.MaxResolvedPeers,
	}
}

func (ps *PriorityStrategy) Start() {
	// start p2p process threads
	go ps.ResolvePeersThread()
	go ps.GetMorePeersThread()
	go ps.UpdateBlacklistedStatusThread()
	go ps.ConnectPriorityPeersThread()
}

func (ps *PriorityStrategy) ConnectPriorityPeersThread() {
	go ps.ConnectPriorityPeersGradually()
	ticker := time.NewTicker(time.Duration(constant.ConnectPriorityPeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go ps.ConnectPriorityPeersGradually()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

func (ps *PriorityStrategy) ConnectPriorityPeersGradually() {
	var (
		i int
	)
	hostAddress := &model.Peer{
		Info: ps.Host.Info,
	}

	log.Info("Connecting to priority lists...")

	unresolvedPeers := ps.GetUnresolvedPeers()
	resolvedPeers := ps.GetResolvedPeers()
	blacklistedPeers := ps.GetBlacklistedPeers()
	exceedMaxUnresolvedPeers := ps.GetExceedMaxUnresolvedPeers()

	priorityPeers := ps.GetPriorityPeers()

	for _, peer := range priorityPeers {
		if i >= constant.NumberOfPriorityPeersToBeAdded {
			break
		}

		if unresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			resolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			blacklistedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			p2pUtil.GetFullAddressPeer(hostAddress) != p2pUtil.GetFullAddressPeer(peer) {

			var j int32
			// removing unpriority peer if the UnresolvedPeers has reached max
			for _, unresolvedPeer := range unresolvedPeers {
				if j < exceedMaxUnresolvedPeers {
					break
				}
				if priorityPeers[p2pUtil.GetFullAddressPeer(unresolvedPeer)] == nil {
					_ = ps.RemoveUnresolvedPeer(unresolvedPeer)
					j++
				}
			}

			if exceedMaxUnresolvedPeers < 1 || (exceedMaxUnresolvedPeers > 0 && j == exceedMaxUnresolvedPeers) {
				_ = ps.AddToUnresolvedPeer(peer)
				i++
			}
		}
	}
}

func (ps *PriorityStrategy) GetPriorityPeers() map[string]*model.Peer {
	// Locking write, allowing read
	ps.PriorityPeersLock.RLock()
	defer ps.PriorityPeersLock.RUnlock()

	var priorityPeers = make(map[string]*model.Peer)
	for key, priorityPeer := range ps.PriorityPeers.MainPriority {
		priorityPeers[key] = priorityPeer
	}
	return priorityPeers
}

func (ps *PriorityStrategy) PeerExplorerListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			go ps.BuildScrambleNodes(block.(*model.Block))
		},
	}
}

// BuildScrambleNodes,  sort node registry to build scramble nodes
func (ps *PriorityStrategy) BuildScrambleNodes(block *model.Block) {
	// Check it's time to bild scramble node or not
	if block.GetHeight()%constant.PriorityStrategyBuildScrambleNodesGap == 0 {
		var nodeRegistries []*model.NodeRegistration

		// get node registry list
		rows, err := ps.QueryExecutor.ExecuteSelect(
			ps.NodeRegistrationQuery.GetNodeRegistryAtHeight(block.GetHeight()),
			false,
		)
		if err != nil {
			// TODO: catching err into log file
			return
		}
		nodeRegistries = ps.NodeRegistrationQuery.BuildModel(nodeRegistries, rows)

		// sort node registry
		sort.SliceStable(nodeRegistries, func(i, j int) bool {
			ni, nj := nodeRegistries[i], nodeRegistries[j]
			// bs, Last 8 bytes of block seed hash in big int
			bs := new(big.Int).SetBytes(block.GetBlockSeed()).Int64()
			// Node public key in big int
			nodePKi := new(big.Int).SetBytes(ni.GetNodePublicKey())
			nodePKj := new(big.Int).SetBytes(nj.GetNodePublicKey())
			//  Last 8 bytes Public Key of Node Registry   XOR   Last 8 bytes Block Seed
			resI := new(big.Int).SetInt64(nodePKi.Int64() ^ bs)
			resJ := new(big.Int).SetInt64(nodePKj.Int64() ^ bs)

			res := resI.Cmp(resJ)
			if res == 0 {
				// Compare node public key
				res = nodePKi.Cmp(nodePKj)
			}
			// Ascending sort
			return res < 0
		})

		// Only select sorted node registry until max scramble nodes
		if len(nodeRegistries) > constant.PriorityStrategyMaxScrambleNodes {
			nodeRegistries = nodeRegistries[:constant.PriorityStrategyMaxScrambleNodes]
		}
		ps.CollectPriorityPeers(nodeRegistries)
	}
}

// CollectPriorityPeers, collect new priority peers from scrambled nodes
func (ps *PriorityStrategy) CollectPriorityPeers(scrambleNodes []*model.NodeRegistration) {
	var (
		hostPotition           int
		newPriorityPeers       = make(map[string]*model.Peer)
		newMutualPriorityPeers = make(map[string]*model.Peer)
		isInScrambleNodes      = false
		hostFullAddress        = p2pUtil.GetFullAddressPeer(
			&model.Peer{
				Info: ps.Host.GetInfo(),
			},
		)
	)
	for potition, node := range scrambleNodes {
		// TODO: Adding port of node address in node registry ??
		if hostFullAddress == p2pUtil.GetFullAddress(node.GetNodeAddress(), 8001) {
			isInScrambleNodes = true
			hostPotition = potition
			break
		}
	}

	if isInScrambleNodes {
		var (
			addedPotition       = 1
			mainPeersPotition   int
			mainResetPotiton    int
			mutualPeersPotition int
			mutualResetPotition = len(scrambleNodes) - 1
		)
		/*
			Find Peers Priority
			For now priority peers is next node after host postition until max priotity peers
		*/
		for addedPotition <= constant.PriorityStrategyMaxPriorityPeers {
			// Get main peers potition
			if (hostPotition + addedPotition) < len(scrambleNodes) {
				mainPeersPotition = hostPotition + addedPotition
			} else {
				if mainResetPotiton != hostPotition {
					mainPeersPotition = mainResetPotiton
				}
				mainResetPotiton++
			}

			// Get mutual peers potition
			if (hostPotition - addedPotition) >= 0 {
				mutualPeersPotition = hostPotition - addedPotition
			} else {
				if mutualPeersPotition == 0 {
					mutualResetPotition = len(scrambleNodes)
					mutualPeersPotition = len(scrambleNodes) - 1
				} else {
					mutualPeersPotition = mutualResetPotition
				}
				mutualResetPotition--
			}

			newPriorityPeer := &model.Peer{
				Info: &model.Node{
					Address: scrambleNodes[mainPeersPotition].GetNodeAddress(),
					// TODO: Adding Port of node address in node registry ??
					Port: 8001,
				},
			}
			newMutualPeer := &model.Peer{
				Info: &model.Node{
					Address: scrambleNodes[mutualPeersPotition].GetNodeAddress(),
					// TODO: Adding Port of node address in node registry ??
					Port: 8001,
				},
			}
			newPriorityPeers[p2pUtil.GetFullAddressPeer(newPriorityPeer)] = newPriorityPeer
			newMutualPriorityPeers[p2pUtil.GetFullAddressPeer(newMutualPeer)] = newMutualPeer

			if mainResetPotiton == len(scrambleNodes) {
				break
			}
			addedPotition++
		}
	}

	// save new priority peers
	ps.PriorityPeersLock.Lock()
	ps.PriorityPeers.MainPriority = newPriorityPeers
	ps.PriorityPeers.MutualPriority = newMutualPriorityPeers
	ps.PriorityPeersLock.Unlock()
}

// ============================================
// Thread resolving peers
// ============================================

// ResolvePeersThread to periodically try get response from peers in UnresolvedPeer list
func (ps *PriorityStrategy) ResolvePeersThread() {
	go ps.ResolvePeers()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go ps.ResolvePeers()
			go ps.UpdateResolvedPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// ResolvePeers looping unresolved peers and adding to (resolve) Peers if get response
func (ps *PriorityStrategy) ResolvePeers() {
	exceedMaxResolvedPeers := ps.GetExceedMaxResolvedPeers()
	priorityPeers := ps.GetPriorityPeers()
	resolvedPeers := ps.GetResolvedPeers()
	unresolvedPeers := ps.GetUnresolvedPeers()
	var (
		removedResolvedPeers    int32
		priorityUnresolvedPeers = make(map[string]*model.Peer)
	)

	// collecting unresolved peers that are priority
	for _, unresolvedPeer := range unresolvedPeers {
		if priorityPeers[p2pUtil.GetFullAddressPeer(unresolvedPeer)] != nil && resolvedPeers[p2pUtil.GetFullAddressPeer(unresolvedPeer)] == nil {
			priorityUnresolvedPeers[p2pUtil.GetFullAddressPeer(unresolvedPeer)] = unresolvedPeer
		}
	}

	// making room in the resolvedPeers list for the priority peers
	maxToRemove := exceedMaxResolvedPeers - int32(1) + int32(len(priorityUnresolvedPeers))
	for _, resolvedPeer := range resolvedPeers {
		if removedResolvedPeers >= maxToRemove {
			break
		}

		if priorityPeers[p2pUtil.GetFullAddressPeer(resolvedPeer)] == nil {
			_ = ps.RemoveResolvedPeer(resolvedPeer)
			removedResolvedPeers++
		}
	}

	// resolving to priority unresolvedPeers until the resolvedPeers list is full
	var i int32
	maxAddedPeers := removedResolvedPeers - exceedMaxResolvedPeers + int32(1)
	for _, priorityUnresolvedPeer := range priorityUnresolvedPeers {
		if i >= maxAddedPeers {
			break
		}
		go ps.resolvePeer(priorityUnresolvedPeer)
		i++
	}

	// resolving other peers that are not priority if resolvedPeers is not full yet
	for _, peer := range ps.GetUnresolvedPeers() {
		if i >= maxAddedPeers {
			break
		}

		if priorityUnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil {
			go ps.resolvePeer(peer)
			i++
		}
	}
}

func (ps *PriorityStrategy) UpdateResolvedPeers() {
	currentTime := time.Now().UTC()
	for _, peer := range ps.GetResolvedPeers() {
		if currentTime.Unix()-peer.GetLastUpdated() >= constant.SecondsToUpdatePeersConnection {
			go ps.resolvePeer(peer)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (ps *PriorityStrategy) resolvePeer(destPeer *model.Peer) {
	_, err := ps.PeerServiceClient.GetPeerInfo(destPeer)
	if err != nil {
		// TODO: add mechanism to blacklist failing peers
		ps.DisconnectPeer(destPeer)
		return
	}
	if destPeer != nil {
		destPeer.LastUpdated = time.Now().UTC().Unix()
	}
	_ = ps.RemoveUnresolvedPeer(destPeer)
	_ = ps.AddToResolvedPeer(destPeer)
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (ps *PriorityStrategy) GetMorePeersHandler() (*model.Peer, error) {
	peer := ps.GetAnyResolvedPeer()
	if peer != nil {
		newPeers, err := ps.PeerServiceClient.GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error: %v\n", err)
			return nil, err
		}
		err = ps.AddToUnresolvedPeers(newPeers.GetPeers(), false)
		if err != nil {
			log.Warnf("getMorePeers error: %v\n", err)
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
func (ps *PriorityStrategy) GetMorePeersThread() {
	syncPeers := func() {
		peer, err := ps.GetMorePeersHandler()
		if err != nil {
			return
		}
		var myPeers []*model.Node
		myResolvedPeers := ps.GetResolvedPeers()
		for _, peer := range myResolvedPeers {
			myPeers = append(myPeers, peer.Info)
		}
		if peer == nil {
			return
		}
		myPeers = append(myPeers, ps.Host.GetInfo())
		_, _ = ps.PeerServiceClient.SendPeers(
			peer,
			myPeers,
		)
	}

	go syncPeers()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go func() {
				go syncPeers()
			}()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// UpdateBlacklistedStatusThread to periodically check blacklisting time of black listed peer,
// every 60sec if there are blacklisted peers to unblacklist
func (ps *PriorityStrategy) UpdateBlacklistedStatusThread() {
	ticker := time.NewTicker(time.Duration(constant.UpdateBlacklistedStatusGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint64(time.Now().Unix())
				for _, p := range ps.Host.GetBlacklistedPeers() {
					if p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+constant.BlacklistingPeriod <= curTime {
						_ = ps.PeerUnblacklist(p)
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

/* 	========================================
 *	Resolved Peers Operations
 *	========================================
 */

func (ps *PriorityStrategy) GetHostInfo() *model.Node {
	return ps.Host.GetInfo()
}

// GetResolvedPeers returns resolved peers in thread-safe manner
func (ps *PriorityStrategy) GetResolvedPeers() map[string]*model.Peer {
	ps.ResolvedPeersLock.RLock()
	defer ps.ResolvedPeersLock.RUnlock()

	var newResolvedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range ps.Host.ResolvedPeers {
		newResolvedPeers[key] = resolvedPeer
	}
	return newResolvedPeers
}

// GetAnyResolvedPeer Get any random resolved peer
func (ps *PriorityStrategy) GetAnyResolvedPeer() *model.Peer {
	resolvedPeers := ps.GetResolvedPeers()
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
func (ps *PriorityStrategy) AddToResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ps.ResolvedPeersLock.Lock()
	defer ps.ResolvedPeersLock.Unlock()

	ps.Host.ResolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// RemoveResolvedPeer removes peer from Resolved peer list
func (ps *PriorityStrategy) RemoveResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ps.ResolvedPeersLock.Lock()
	defer ps.ResolvedPeersLock.Unlock()
	delete(ps.Host.ResolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

/* 	========================================
 *	Unresolved Peers Operations
 *	========================================
 */

// GetUnresolvedPeers returns unresolved peers in thread-safe manner
func (ps *PriorityStrategy) GetUnresolvedPeers() map[string]*model.Peer {
	ps.UnresolvedPeersLock.RLock()
	defer ps.UnresolvedPeersLock.RUnlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)
	for key, UnresolvedPeer := range ps.Host.UnresolvedPeers {
		newUnresolvedPeers[key] = UnresolvedPeer
	}
	return newUnresolvedPeers
}

// GetAnyUnresolvedPeer Get any unresolved peer
func (ps *PriorityStrategy) GetAnyUnresolvedPeer() *model.Peer {
	unresolvedPeers := ps.GetUnresolvedPeers()
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
func (ps *PriorityStrategy) AddToUnresolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ps.UnresolvedPeersLock.Lock()
	defer ps.UnresolvedPeersLock.Unlock()
	ps.Host.UnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
// toForce: this parameter is not used in the PriorityStrategy because
//			only priority nodes can forcefully get into peers list
func (ps *PriorityStrategy) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error {
	exceedMaxUnresolvedPeers := ps.GetExceedMaxUnresolvedPeers()

	// do not force a peer to go to unresolved list if the list is full
	if exceedMaxUnresolvedPeers > 1 {
		return errors.New("unresolvedPeers are full")
	}

	var peersAdded int32

	unresolvedPeers := ps.GetUnresolvedPeers()
	resolvedPeers := ps.GetResolvedPeers()

	hostAddress := &model.Peer{
		Info: ps.Host.Info,
	}
	for _, node := range newNodes {
		peer := &model.Peer{
			Info: node,
		}
		if unresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			resolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			p2pUtil.GetFullAddressPeer(hostAddress) != p2pUtil.GetFullAddressPeer(peer) {
			_ = ps.AddToUnresolvedPeer(peer)
			peersAdded++
		}

		if exceedMaxUnresolvedPeers+peersAdded > 1 {
			break
		}
	}
	return nil
}

// RemoveUnresolvedPeer removes peer from unresolved peer list
func (ps *PriorityStrategy) RemoveUnresolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ps.UnresolvedPeersLock.Lock()
	defer ps.UnresolvedPeersLock.Unlock()
	delete(ps.Host.UnresolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

/* 	========================================
 *	Blacklisted Peers Operations
 *	========================================
 */

// GetBlacklistedPeers returns resolved peers in thread-safe manner
func (ps *PriorityStrategy) GetBlacklistedPeers() map[string]*model.Peer {
	ps.BlacklistedPeersLock.RLock()
	defer ps.BlacklistedPeersLock.RUnlock()

	var newBlacklistedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range ps.Host.BlacklistedPeers {
		newBlacklistedPeers[key] = resolvedPeer
	}
	return newBlacklistedPeers
}

// AddToBlacklistedPeer to add a peer into resolved peer
func (ps *PriorityStrategy) AddToBlacklistedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ps.BlacklistedPeersLock.Lock()
	defer ps.BlacklistedPeersLock.Unlock()

	ps.Host.BlacklistedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	return nil
}

// RemoveBlacklistedPeer removes peer from Blacklisted peer list
func (ps *PriorityStrategy) RemoveBlacklistedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	ps.BlacklistedPeersLock.Lock()
	defer ps.BlacklistedPeersLock.Unlock()
	delete(ps.Host.BlacklistedPeers, p2pUtil.GetFullAddressPeer(peer))
	return nil
}

// ======================================================
// Exposed Functions
// ======================================================

// GetAnyKnownPeer Get any known peer
func (ps *PriorityStrategy) GetAnyKnownPeer() *model.Peer {
	knownPeers := ps.Host.KnownPeers
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
func (ps *PriorityStrategy) GetExceedMaxUnresolvedPeers() int32 {
	return int32(len(ps.GetUnresolvedPeers())) - ps.MaxUnresolvedPeers + 1
}

// GetExceedMaxResolvedPeers returns number of peers exceeding max number of the resolved peers
func (ps *PriorityStrategy) GetExceedMaxResolvedPeers() int32 {
	return int32(len(ps.GetResolvedPeers())) - ps.MaxResolvedPeers + 1
}

func (ps *PriorityStrategy) PeerBlacklist(peer *model.Peer, cause string) {
	peer.BlacklistingTime = uint64(time.Now().Unix())
	peer.BlacklistingCause = cause
	_ = ps.AddToBlacklistedPeer(peer)
	_ = ps.RemoveUnresolvedPeer(peer)
	_ = ps.RemoveResolvedPeer(peer)
}

// PeerUnblacklist to update Peer state of peer
func (ps *PriorityStrategy) PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	_ = ps.RemoveBlacklistedPeer(peer)
	_ = ps.AddToUnresolvedPeers([]*model.Node{peer.Info}, false)
	return peer
}

// DisconnectPeer moves connected peer to resolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func (ps *PriorityStrategy) DisconnectPeer(peer *model.Peer) {
	_ = ps.RemoveResolvedPeer(peer)
	if ps.GetExceedMaxUnresolvedPeers() <= 0 {
		_ = ps.AddToUnresolvedPeer(peer)
	}
}
