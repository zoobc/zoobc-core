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
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/client"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc/metadata"
)

// PriorityStrategy represent data service node as server
type (
	PriorityStrategy struct {
		Host                    *model.Host
		PeerServiceClient       client.PeerServiceClientInterface
		NodeRegistrationService coreService.NodeRegistrationServiceInterface
		QueryExecutor           query.ExecutorInterface
		ResolvedPeersLock       sync.RWMutex
		UnresolvedPeersLock     sync.RWMutex
		BlacklistedPeersLock    sync.RWMutex
		MaxUnresolvedPeers      int32
		MaxResolvedPeers        int32
		Logger                  *log.Logger
	}
)

func NewPriorityStrategy(
	host *model.Host,
	peerServiceClient client.PeerServiceClientInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	queryExecutor query.ExecutorInterface,
	logger *log.Logger,
) *PriorityStrategy {
	return &PriorityStrategy{
		Host:                    host,
		PeerServiceClient:       peerServiceClient,
		NodeRegistrationService: nodeRegistrationService,
		QueryExecutor:           queryExecutor,
		MaxUnresolvedPeers:      constant.MaxUnresolvedPeers,
		MaxResolvedPeers:        constant.MaxResolvedPeers,
		Logger:                  logger,
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

	ps.Logger.Info("Connecting to priority lists...")

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
					err := ps.RemoveUnresolvedPeer(unresolvedPeer)
					if err != nil {
						ps.Logger.Error(err.Error())
					}
					j++
				}
			}

			if exceedMaxUnresolvedPeers < 1 || (exceedMaxUnresolvedPeers > 0 && j == exceedMaxUnresolvedPeers) {
				err := ps.AddToUnresolvedPeer(peer)
				if err != nil {
					ps.Logger.Error(err)
				}
				i++
			}
		}
	}
}

// GetPriorityPeers, to get a list peer should connect if host in scramble node
func (ps *PriorityStrategy) GetPriorityPeers() map[string]*model.Peer {
	var (
		priorityPeers   = make(map[string]*model.Peer)
		hostFullAddress = p2pUtil.GetFullAddressPeer(&model.Peer{
			Info: ps.Host.GetInfo(),
		})
	)

	if ps.ValidateScrambleNode(ps.Host.GetInfo()) {
		scrambledNodes := ps.NodeRegistrationService.GetLatestScrambledNodes()
		var (
			hostIndex     = scrambledNodes.IndexNodes[hostFullAddress]
			startPeers    = p2pUtil.GetStartIndexPriorityPeer(*hostIndex, scrambledNodes)
			addedPosition = 0
		)
		for addedPosition < constant.PriorityStrategyMaxPriorityPeers {
			var (
				peersPosition = (startPeers + addedPosition + 1) % (len(scrambledNodes.IndexNodes))
				peer          = scrambledNodes.AddressNodes[peersPosition]
				addressPeer   = p2pUtil.GetFullAddressPeer(peer)
			)
			if priorityPeers[addressPeer] != nil {
				break
			}
			if addressPeer != hostFullAddress {
				priorityPeers[addressPeer] = peer
			}
			addedPosition++
		}

	}
	return priorityPeers
}

// ValidateScrambleNode, check node in scramble or not
func (ps *PriorityStrategy) ValidateScrambleNode(node *model.Node) bool {
	scrambledNodes := ps.NodeRegistrationService.GetLatestScrambledNodes()
	address := p2pUtil.GetFullAddress(node)
	return scrambledNodes.IndexNodes[address] != nil
}

// ValidatePriorityPeer, check peer is in priority list peer of host node
func (ps *PriorityStrategy) ValidatePriorityPeer(host, peer *model.Node) bool {
	if ps.ValidateScrambleNode(host) && ps.ValidateScrambleNode(peer) {
		scrambledNodes := ps.NodeRegistrationService.GetLatestScrambledNodes()
		var (
			hostIndex          = *scrambledNodes.IndexNodes[p2pUtil.GetFullAddress(host)]
			peerIndex          = *scrambledNodes.IndexNodes[p2pUtil.GetFullAddress(peer)]
			hostStartPeerIndex = p2pUtil.GetStartIndexPriorityPeer(hostIndex, scrambledNodes)
			hostEndPeerIndex   = (hostStartPeerIndex + constant.PriorityStrategyMaxPriorityPeers - 1) % (len(scrambledNodes.AddressNodes))
		)
		return ps.ValidateRangePriorityPeers(peerIndex, hostStartPeerIndex, hostEndPeerIndex)
	}
	return false
}

func (ps *PriorityStrategy) ValidateRangePriorityPeers(peerIndex, hostStartPeerIndex, hostEndPeerIndex int) bool {
	if hostEndPeerIndex > hostStartPeerIndex {
		return peerIndex >= hostStartPeerIndex && peerIndex <= hostEndPeerIndex
	}
	if peerIndex >= hostStartPeerIndex {
		return true
	}
	return peerIndex <= hostEndPeerIndex
}

// ValidateRequest, to validate incoming request based on metadata in context and Priority strategy
func (ps *PriorityStrategy) ValidateRequest(ctx context.Context) bool {
	if ctx != nil {
		md, _ := metadata.FromIncomingContext(ctx)
		// Check have default context
		if len(md.Get(p2pUtil.DefaultConnectionMetadata)) != 0 {
			// Check host in scramble nodes
			if ps.ValidateScrambleNode(ps.Host.GetInfo()) {
				var (
					fullAddress           = md.Get(p2pUtil.DefaultConnectionMetadata)[0]
					nodeRequester         = p2pUtil.GetNodeInfo(fullAddress)
					resolvedPeers         = ps.GetResolvedPeers()
					exceedUnresolvedPeers = ps.GetExceedMaxUnresolvedPeers()
					blacklistedPeers      = ps.GetBlacklistedPeers()
				)
				// add into unresolved peers if any space
				if (exceedUnresolvedPeers < 1) && (resolvedPeers[fullAddress] == nil) && (blacklistedPeers[fullAddress] == nil) {
					if err := ps.AddToUnresolvedPeer(&model.Peer{Info: nodeRequester}); err != nil {
						ps.Logger.Warn(err.Error())
					}
				}

				// Check host is in priority peer list of requester
				// Or requester is in priority peers of host
				// Or requester is in resolved peers of host
				// Or unrelosovedPeers still have available space
				return ps.ValidatePriorityPeer(nodeRequester, ps.Host.GetInfo()) ||
					ps.ValidatePriorityPeer(ps.Host.GetInfo(), nodeRequester) ||
					(resolvedPeers[fullAddress] != nil) ||
					(blacklistedPeers[fullAddress] == nil && exceedUnresolvedPeers < 1)

			}
			return true
		}
	}
	return false
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
			ps.DisconnectPeer(resolvedPeer)
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
		go ps.resolvePeer(priorityUnresolvedPeer, true)
		i++
	}

	// resolving other peers that are not priority if resolvedPeers is not full yet
	for _, peer := range ps.GetUnresolvedPeers() {
		// add one space in resolved peers
		// will use to cycle non priority resolved peer
		if i >= maxAddedPeers+1 {
			break
		}

		if priorityUnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil {
			// unresolved peer that non priority when failed connet will remove permanently
			go ps.resolvePeer(peer, false)
			i++
		}
	}
}

func (ps *PriorityStrategy) UpdateResolvedPeers() {
	currentTime := time.Now().UTC()
	for _, peer := range ps.GetResolvedPeers() {
		if currentTime.Unix()-peer.GetLastUpdated() >= constant.SecondsToUpdatePeersConnection {
			go ps.resolvePeer(peer, true)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (ps *PriorityStrategy) resolvePeer(destPeer *model.Peer, wantToKeep bool) {
	_, err := ps.PeerServiceClient.GetPeerInfo(destPeer)
	if err != nil {
		// TODO: add mechanism to blacklist failing peers
		// will add into unresolved peer list if want to keep
		// sotherwise remove permanently
		if wantToKeep {
			ps.DisconnectPeer(destPeer)
			return
		}
		if err := ps.RemoveResolvedPeer(destPeer); err != nil {
			ps.Logger.Warn(err)
		}
		if err := ps.RemoveUnresolvedPeer(destPeer); err != nil {
			ps.Logger.Warn(err)
		}
		return
	}
	if destPeer != nil {
		destPeer.LastUpdated = time.Now().UTC().Unix()
	}
	if err = ps.RemoveUnresolvedPeer(destPeer); err != nil {
		ps.Logger.Error(err.Error())
	}
	if err = ps.AddToResolvedPeer(destPeer); err != nil {
		ps.Logger.Error(err.Error())
	}
}

// GetMorePeersHandler request peers to random peer in list and if get new peers will add to unresolved peer
func (ps *PriorityStrategy) GetMorePeersHandler() (*model.Peer, error) {
	peer := ps.GetAnyResolvedPeer()
	if peer != nil {
		newPeers, err := ps.PeerServiceClient.GetMorePeers(peer)
		if err != nil {
			ps.Logger.Infof("getMorePeers Error: %v\n", err)
			return nil, err
		}
		err = ps.AddToUnresolvedPeers(newPeers.GetPeers(), false)
		if err != nil {
			ps.Logger.Errorf("getMorePeers error: %v\n", err)
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
			ps.Logger.Warn(err.Error())
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
	ps.UnresolvedPeersLock.Lock()
	defer ps.UnresolvedPeersLock.Unlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)
	for key, UnresolvedPeer := range ps.Host.UnresolvedPeers {
		newUnresolvedPeers[key] = UnresolvedPeer
	}
	// Add known peers into unresolved peer list if the list is empty
	if len(newUnresolvedPeers) == 0 {
		for key, peer := range ps.Host.GetKnownPeers() {
			var (
				newPeer           = *peer
				newUnresvoledPeer = *peer
			)
			newUnresolvedPeers[key] = &newPeer
			ps.Host.UnresolvedPeers[key] = &newUnresvoledPeer
		}

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
		return errors.New("AddToUnresolvedPeer Err, peer is nil")
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
	var exceedMaxUnresolvedPeers = ps.GetExceedMaxUnresolvedPeers()

	// do not force a peer to go to unresolved list if the list is full
	if exceedMaxUnresolvedPeers >= 1 {
		return errors.New("unresolvedPeers are full")
	}
	var (
		peersAdded      int32
		unresolvedPeers = ps.GetUnresolvedPeers()
		resolvedPeers   = ps.GetResolvedPeers()
		hostAddress     = &model.Peer{
			Info: ps.Host.Info,
		}
	)
	for _, node := range newNodes {
		peer := &model.Peer{
			Info: node,
		}
		if unresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			resolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil &&
			p2pUtil.GetFullAddressPeer(hostAddress) != p2pUtil.GetFullAddressPeer(peer) {
			if err := ps.AddToUnresolvedPeer(peer); err != nil {
				ps.Logger.Error(err.Error())
			}
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
		return errors.New("RemoveUnresolvedPeer Err, peer is nil")
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
		return errors.New("AddToBlacklistedPeer Err, peer is nil")
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
	if err := ps.AddToBlacklistedPeer(peer); err != nil {
		ps.Logger.Error(err.Error())
	}
	if err := ps.RemoveUnresolvedPeer(peer); err != nil {
		ps.Logger.Error(err.Error())
	}
	if err := ps.RemoveResolvedPeer(peer); err != nil {
		ps.Logger.Error(err.Error())
	}
}

// PeerUnblacklist to update Peer state of peer
func (ps *PriorityStrategy) PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	if err := ps.RemoveBlacklistedPeer(peer); err != nil {
		ps.Logger.Error(err.Error())
	}
	if err := ps.AddToUnresolvedPeers([]*model.Node{peer.Info}, false); err != nil {
		ps.Logger.Error(err.Error())
	}

	return peer
}

// DisconnectPeer moves connected peer to unresolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func (ps *PriorityStrategy) DisconnectPeer(peer *model.Peer) {
	if err := ps.RemoveResolvedPeer(peer); err != nil {
		ps.Logger.Error(err.Error())
	}

	if ps.GetExceedMaxUnresolvedPeers() <= 0 {
		if err := ps.AddToUnresolvedPeer(peer); err != nil {
			ps.Logger.Error(err.Error())
		}
	}
}
