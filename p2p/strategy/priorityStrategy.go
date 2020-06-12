package strategy

import (
	"context"
	"errors"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p/client"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc/metadata"
)

type (
	// PriorityStrategy represent data service node as server
	PriorityStrategy struct {
		BlockchainStatusService  coreService.BlockchainStatusServiceInterface
		NodeConfigurationService coreService.NodeConfigurationServiceInterface
		PeerServiceClient        client.PeerServiceClientInterface
		NodeRegistrationService  coreService.NodeRegistrationServiceInterface
		QueryExecutor            query.ExecutorInterface
		BlockQuery               query.BlockQueryInterface
		ResolvedPeersLock        sync.RWMutex
		UnresolvedPeersLock      sync.RWMutex
		BlacklistedPeersLock     sync.RWMutex
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
		PeerStrategyHelper       PeerStrategyHelperInterface
	}
)

var (
	peerResolved = make(chan bool)
)

func NewPriorityStrategy(
	peerServiceClient client.PeerServiceClientInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	logger *log.Logger,
	peerStrategyHelper PeerStrategyHelperInterface,
	nodeConfigurationService coreService.NodeConfigurationServiceInterface,
	blockchainStatusService coreService.BlockchainStatusServiceInterface,
) *PriorityStrategy {
	return &PriorityStrategy{
		BlockchainStatusService:  blockchainStatusService,
		NodeConfigurationService: nodeConfigurationService,
		PeerServiceClient:        peerServiceClient,
		NodeRegistrationService:  nodeRegistrationService,
		QueryExecutor:            queryExecutor,
		BlockQuery:               blockQuery,
		MaxUnresolvedPeers:       constant.MaxUnresolvedPeers,
		MaxResolvedPeers:         constant.MaxResolvedPeers,
		Logger:                   logger,
		PeerStrategyHelper:       peerStrategyHelper,
	}
}

// Start method to start threads which mean goroutines for PriorityStrategy
func (ps *PriorityStrategy) Start() {
	// start p2p process threads
	go ps.ResolvePeersThread()
	go ps.GetMorePeersThread()
	go ps.UpdateBlacklistedStatusThread()
	go ps.ConnectPriorityPeersThread()
	// wait until there is at least one connected (resolved) peer we can communicate to
	<-peerResolved
	go ps.UpdateNodeAddressThread()
	go ps.SyncNodeAddressInfoTableThread()
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
		i                            int
		unresolvedPriorityPeersCount int
		resolvedPriorityPeersCount   int
		unresolvedPeers              = ps.GetUnresolvedPeers()
		resolvedPeers                = ps.GetResolvedPeers()
		blacklistedPeers             = ps.GetBlacklistedPeers()
		exceedMaxUnresolvedPeers     = ps.GetExceedMaxUnresolvedPeers() - 1
		priorityPeers                = ps.GetPriorityPeers()
		hostModelPeer                = &model.Peer{
			Info: ps.NodeConfigurationService.GetHost().Info,
		}
		hostAddress = p2pUtil.GetFullAddressPeer(hostModelPeer)
	)
	ps.Logger.Infoln("Connecting to priority lists...")

	for _, peer := range priorityPeers {
		if i >= constant.NumberOfPriorityPeersToBeAdded {
			break
		}
		priorityNodeAddress := p2pUtil.GetFullAddressPeer(peer)

		if unresolvedPeers[priorityNodeAddress] == nil &&
			resolvedPeers[priorityNodeAddress] == nil &&
			blacklistedPeers[priorityNodeAddress] == nil &&
			hostAddress != priorityNodeAddress {

			newPeer := *peer

			// removing non priority peers and replacing if no space
			if exceedMaxUnresolvedPeers >= 0 {
				for _, unresolvedPeer := range unresolvedPeers {
					unresolvedNodeAddress := p2pUtil.GetFullAddressPeer(unresolvedPeer)
					if priorityPeers[unresolvedNodeAddress] == nil {
						err := ps.RemoveUnresolvedPeer(unresolvedPeer)
						if err != nil {
							ps.Logger.Error(err.Error())
							continue
						}
						delete(unresolvedPeers, unresolvedNodeAddress)
						break
					}
				}
			}

			// add the priority peers to unresolvedPeers
			err := ps.AddToUnresolvedPeer(&newPeer)
			if err != nil {
				ps.Logger.Error(err)
			}
			unresolvedPeers[priorityNodeAddress] = &newPeer
			exceedMaxUnresolvedPeers++
			i++
		}
	}

	if monitoring.IsMonitoringActive() {
		for _, peer := range priorityPeers {
			priorityNodeAddress := p2pUtil.GetFullAddressPeer(peer)
			if unresolvedPeers[priorityNodeAddress] != nil {
				unresolvedPriorityPeersCount++
			}
			if resolvedPeers[priorityNodeAddress] != nil {
				resolvedPriorityPeersCount++
			}
		}
		// metrics monitoring
		monitoring.SetResolvedPriorityPeersCount(resolvedPriorityPeersCount)
		monitoring.SetUnresolvedPriorityPeersCount(unresolvedPriorityPeersCount)
	}
}

// GetPriorityPeers, to get a list peer should connect if host in scramble node
func (ps *PriorityStrategy) GetPriorityPeers() map[string]*model.Peer {
	var (
		priorityPeers   = make(map[string]*model.Peer)
		host            = ps.NodeConfigurationService.GetHost()
		hostFullAddress = p2pUtil.GetFullAddressPeer(&model.Peer{
			Info: host.GetInfo(),
		})
	)
	lastBlock, err := util.GetLastBlock(ps.QueryExecutor, ps.BlockQuery)
	if err != nil {
		return priorityPeers
	}
	scrambledNodes, err := ps.NodeRegistrationService.GetScrambleNodesByHeight(lastBlock.Height)
	if err != nil {
		return priorityPeers
	}

	if ps.ValidateScrambleNode(scrambledNodes, host.GetInfo()) {
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
func (ps *PriorityStrategy) ValidateScrambleNode(scrambledNodes *model.ScrambledNodes, node *model.Node) bool {
	var address = p2pUtil.GetFullAddress(node)
	return scrambledNodes.IndexNodes[address] != nil
}

// ValidatePriorityPeer, check peer is in priority list peer of host node
func (ps *PriorityStrategy) ValidatePriorityPeer(scrambledNodes *model.ScrambledNodes, host, peer *model.Node) bool {
	if ps.ValidateScrambleNode(scrambledNodes, host) && ps.ValidateScrambleNode(scrambledNodes, peer) {
		priorityPeers, err := p2pUtil.GetPriorityPeersByNodeFullAddress(p2pUtil.GetFullAddress(host), scrambledNodes)
		if err != nil {
			return false
		}
		return priorityPeers[p2pUtil.GetFullAddress(peer)] != nil
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

// ValidateRequest to validate incoming request based on metadata in context and Priority strategy
func (ps *PriorityStrategy) ValidateRequest(ctx context.Context) bool {
	if ctx != nil {
		md, ok := metadata.FromIncomingContext(ctx) // check ok
		if !ok {
			return false
		}
		// Check have default context
		if len(md.Get(p2pUtil.DefaultConnectionMetadata)) != 0 {
			var (
				host = ps.NodeConfigurationService.GetHost()
			)
			// get scramble node
			// NOTE: calling this query is highly possibility issued error `db is locked`. Need optimize
			lastBlock, err := util.GetLastBlock(ps.QueryExecutor, ps.BlockQuery)
			if err != nil {
				ps.Logger.Errorf("ValidateRequestFailGetLastBlock: %v", err)
				return false
			}
			scrambledNodes, err := ps.NodeRegistrationService.GetScrambleNodesByHeight(lastBlock.Height)
			if err != nil {
				ps.Logger.Errorf("FailGetScrambleNodesByHeight: %v", err)
				return false
			}

			// STEF this validates always the local node's host against scramble nodes
			// Check host in scramble nodes
			if ps.ValidateScrambleNode(scrambledNodes, host.GetInfo()) {
				var (
					fullAddress         = md.Get(p2pUtil.DefaultConnectionMetadata)[0]
					nodeRequester       = p2pUtil.GetNodeInfo(fullAddress)
					resolvedPeers       = ps.GetResolvedPeers()
					unresolvedPeers     = ps.GetUnresolvedPeers()
					blacklistedPeers    = ps.GetBlacklistedPeers()
					isAddedToUnresolved = false
				)

				if unresolvedPeers[fullAddress] == nil && blacklistedPeers[fullAddress] == nil {
					if len(unresolvedPeers) < int(constant.MaxUnresolvedPeers) {
						if err = ps.AddToUnresolvedPeer(&model.Peer{Info: nodeRequester}); err != nil {
							ps.Logger.Error(err.Error())
						} else {
							isAddedToUnresolved = true
						}
					} else {
						for _, peer := range unresolvedPeers {
							// add peer requester into unresolved and remove the old one in unresolved peers
							// removing one of unresolved peers will do when already stayed more than max stayed
							// and not priority peers
							if peer.UnresolvingTime >= constant.PriorityStrategyMaxStayedInUnresolvedPeers &&
								!ps.ValidatePriorityPeer(scrambledNodes, host.GetInfo(), peer.GetInfo()) {
								if err = ps.RemoveUnresolvedPeer(peer); err == nil {
									if err = ps.AddToUnresolvedPeer(&model.Peer{Info: nodeRequester}); err != nil {
										ps.Logger.Error(err.Error())
										break
									}
									isAddedToUnresolved = true
									break
								}
							}
						}
					}
				}

				// Check host is in priority peer list of requester
				// Or requester is in priority peers of host
				// Or requester is in resolved peers of host
				// Or requester is in unresolved peers of host And not in blacklisted peer
				// Or requester added into unresolved peers
				return ps.ValidatePriorityPeer(scrambledNodes, nodeRequester, host.GetInfo()) ||
					ps.ValidatePriorityPeer(scrambledNodes, host.GetInfo(), nodeRequester) ||
					(resolvedPeers[fullAddress] != nil) ||
					(unresolvedPeers[fullAddress] != nil) ||
					isAddedToUnresolved
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
		if i >= maxAddedPeers {
			break
		}

		if priorityUnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] == nil {
			// unresolved peer that non priority when failed connet will remove permanently
			go ps.resolvePeer(peer, false)
			i++
		}
	}
}

// UpdateResolvedPeers use to maintaining resolved peers
func (ps *PriorityStrategy) UpdateResolvedPeers() {
	var (
		priorityPeers = ps.GetPriorityPeers()
		currentTime   = time.Now().UTC()
	)
	for _, peer := range ps.GetResolvedPeers() {
		// priority peers no need to maintenance
		if priorityPeers[p2pUtil.GetFullAddressPeer(peer)] == nil {
			if currentTime.Unix()-peer.GetResolvingTime() >= constant.SecondsToUpdatePeersConnection {
				go ps.resolvePeer(peer, true)
			}
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
		destPeer.ResolvingTime = time.Now().UTC().Unix()
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
			ps.Logger.Warnf("getMorePeers error: %v\n", err)
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
		var (
			nodes []*model.Node
			peer  *model.Peer
			err   error
		)

		peer, err = ps.GetMorePeersHandler()
		if err != nil {
			ps.Logger.Warn(err.Error())
			return
		}
		myResolvedPeers := ps.GetResolvedPeers()
		for _, p := range myResolvedPeers {
			nodes = append(nodes, p.Info)
		}
		if peer == nil {
			return
		}

		nodes = append(nodes, ps.NodeConfigurationService.GetHost().GetInfo())
		_, _ = ps.PeerServiceClient.SendPeers(
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

// GetMorePeersThread to periodically request more peers from another node in Peers list
func (ps *PriorityStrategy) UpdateNodeAddressThread() {
	var (
		timeInterval uint
	)
	myAddressDynamic := ps.NodeConfigurationService.IsMyAddressDynamic()
	currentAddr, err := ps.NodeConfigurationService.GetMyAddress()
	if myAddressDynamic {
		timeInterval = constant.UpdateNodeAddressGap
	} else {
		// if can't connect to any resolved peers, wait till we hopefully get some more from the network
		timeInterval = constant.ResolvePeersGap * 10
	}
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	if err != nil {
		ps.Logger.Warnf("Cannot get address from node. %s", err)
	}
	for {
		// start updating and broadcasting own address when finished downloading the bc,
		// otherwise all new node address info will contain the genesis block, which is a predictable behavior and can be exploited
		if !ps.BlockchainStatusService.IsFirstDownloadFinished((&chaintype.MainChain{})) {
			continue
		}
		var (
			host         = ps.NodeConfigurationService.GetHost()
			secretPhrase = ps.NodeConfigurationService.GetNodeSecretPhrase()
		)
		if !ps.NodeConfigurationService.IsMyAddressDynamic() {
			if err := ps.UpdateOwnNodeAddressInfo(
				currentAddr,
				host.GetInfo().GetPort(),
				secretPhrase,
				true,
			); err == nil {
				// break the loop if address is static (from config file)
				ticker.Stop()
				return
			}
			// get node's public ip address from external source internet and check if differs from current one
		} else if ipAddr, err := (&util.IPUtil{}).DiscoverNodeAddress(); err == nil && ipAddr != nil {
			if err = ps.UpdateOwnNodeAddressInfo(
				ipAddr.String(),
				host.GetInfo().GetPort(),
				secretPhrase,
				false,
			); err != nil {
				ps.Logger.Error(err)
			}
		} else {
			ps.Logger.Error("Cannot get node address from external source")
		}
		select {
		case <-ticker.C:
			continue
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

func (ps *PriorityStrategy) SyncNodeAddressInfoTableThread() {
	// sync the registry with nodeAddressInfo from p2p network as soon as node starts,
	// to have as many priority peers as possible to download the bc from
	if err := ps.getRegistryAndSyncAddressInfoTable(); err != nil {
		ps.Logger.Error(err)
	}

	ticker := time.NewTicker(time.Duration(2) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			// wait until bc has finished downloading and sync nodeAddressInfo table again, to make sure we have all updated addresses
			// note: when bc is full downloaded the node should be able to validate all address info
			// messages
			if ps.BlockchainStatusService.IsFirstDownloadFinished((&chaintype.MainChain{})) {
				if err := ps.getRegistryAndSyncAddressInfoTable(); err != nil {
					ps.Logger.Error(err)
				} else {
					ticker.Stop()
					return
				}
			}
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

func (ps *PriorityStrategy) getRegistryAndSyncAddressInfoTable() error {
	if nodeRegistry, err := ps.NodeRegistrationService.GetNodeRegistry(); err != nil {
		ps.Logger.Fatal(err)
	} else if _, err := ps.SyncNodeAddressInfoTable(nodeRegistry); err != nil {
		return err
	}
	return nil
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
				for _, p := range ps.NodeConfigurationService.GetHost().GetBlacklistedPeers() {
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
	return ps.NodeConfigurationService.GetHost().GetInfo()
}

// GetResolvedPeers returns resolved peers in thread-safe manner
func (ps *PriorityStrategy) GetResolvedPeers() map[string]*model.Peer {
	ps.ResolvedPeersLock.RLock()
	defer ps.ResolvedPeersLock.RUnlock()

	var newResolvedPeers = make(map[string]*model.Peer)
	for key, resolvedPeer := range ps.NodeConfigurationService.GetHost().ResolvedPeers {
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
	var (
		host = ps.NodeConfigurationService.GetHost()
	)
	ps.ResolvedPeersLock.Lock()
	defer func() {
		ps.ResolvedPeersLock.Unlock()
		monitoring.SetResolvedPeersCount(len(ps.NodeConfigurationService.GetHost().ResolvedPeers))
	}()

	host.ResolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	ps.NodeConfigurationService.SetHost(host)
	peerResolved <- true
	return nil
}

// RemoveResolvedPeer removes peer from Resolved peer list
func (ps *PriorityStrategy) RemoveResolvedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	var (
		host = ps.NodeConfigurationService.GetHost()
	)
	ps.ResolvedPeersLock.Lock()
	defer func() {
		ps.ResolvedPeersLock.Unlock()
		monitoring.SetResolvedPeersCount(len(host.ResolvedPeers))
	}()
	err := ps.PeerServiceClient.DeleteConnection(peer)
	if err != nil {
		return err
	}

	delete(host.ResolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	ps.NodeConfigurationService.SetHost(host)
	return nil
}

/* 	========================================
 *	Unresolved Peers Operations
 *	========================================
 */

// GetUnresolvedPeers returns unresolved peers in thread-safe manner
func (ps *PriorityStrategy) GetUnresolvedPeers() map[string]*model.Peer {
	var (
		host = ps.NodeConfigurationService.GetHost()
	)
	ps.UnresolvedPeersLock.Lock()
	defer ps.UnresolvedPeersLock.Unlock()

	var newUnresolvedPeers = make(map[string]*model.Peer)

	// Add known peers into unresolved peer list if the unresolved peers is empty
	if len(host.UnresolvedPeers) == 0 {
		// putting this initialization in a condition to prevent unneeded lock of resolvedPeers and blacklistedPeers
		var (
			resolvedPeers    = ps.GetResolvedPeers()
			blacklistedPeers = ps.GetBlacklistedPeers()
			hostAddressPeer  = &model.Peer{
				Info: host.Info,
			}
			hostAddress = p2pUtil.GetFullAddressPeer(hostAddressPeer)
			counter     int32
		)

		for key, peer := range host.GetKnownPeers() {
			if counter >= constant.MaxUnresolvedPeers {
				break
			}
			peerAddress := p2pUtil.GetFullAddressPeer(peer)
			if resolvedPeers[peerAddress] == nil &&
				blacklistedPeers[peerAddress] == nil &&
				peerAddress != hostAddress {
				newPeer := *peer
				host.UnresolvedPeers[key] = &newPeer
			}
			counter++
		}
		ps.NodeConfigurationService.SetHost(host)
	}

	for key, UnresolvedPeer := range host.UnresolvedPeers {
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
		return errors.New("AddToUnresolvedPeer Err, peer is nil")
	}
	var (
		host = ps.NodeConfigurationService.GetHost()
	)
	ps.UnresolvedPeersLock.Lock()
	defer func() {
		ps.UnresolvedPeersLock.Unlock()
		monitoring.SetUnresolvedPeersCount(len(host.UnresolvedPeers))
	}()
	peer.UnresolvingTime = time.Now().UTC().Unix()
	host.UnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	ps.NodeConfigurationService.SetHost(host)
	return nil
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
// toForce: this parameter is not used in the PriorityStrategy because
//			only priority nodes can forcefully get into peers list
func (ps *PriorityStrategy) AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error {
	var exceedMaxUnresolvedPeers = ps.GetExceedMaxUnresolvedPeers()

	// do not force a peer to go to unresolved list if the list is full
	if exceedMaxUnresolvedPeers >= 1 {
		var rejectedPeers = "rejected peers when unresolved full: "
		for _, node := range newNodes {
			rejectedPeers = fmt.Sprintf("%s %s,", rejectedPeers, p2pUtil.GetFullAddress(node))
		}
		ps.Logger.Warn(rejectedPeers)
		return errors.New("unresolvedPeers are full")
	}
	var (
		peersAdded      int32
		unresolvedPeers = ps.GetUnresolvedPeers()
		resolvedPeers   = ps.GetResolvedPeers()
		hostAddress     = &model.Peer{
			Info: ps.NodeConfigurationService.GetHost().Info,
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
	var (
		host = ps.NodeConfigurationService.GetHost()
	)
	if peer == nil {
		return errors.New("RemoveUnresolvedPeer Err, peer is nil")
	}
	ps.UnresolvedPeersLock.Lock()
	defer func() {
		ps.UnresolvedPeersLock.Unlock()
		monitoring.SetUnresolvedPeersCount(len(host.UnresolvedPeers))
	}()
	delete(host.UnresolvedPeers, p2pUtil.GetFullAddressPeer(peer))
	ps.NodeConfigurationService.SetHost(host)
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
	for key, resolvedPeer := range ps.NodeConfigurationService.GetHost().BlacklistedPeers {
		newBlacklistedPeers[key] = resolvedPeer
	}
	return newBlacklistedPeers
}

// AddToBlacklistedPeer to add a peer into blacklisted peer
func (ps *PriorityStrategy) AddToBlacklistedPeer(peer *model.Peer, cause string) error {
	if peer == nil {
		return errors.New("AddToBlacklisted Peer Err, peer is nil")
	}
	var (
		host = ps.NodeConfigurationService.GetHost()
	)

	peer.BlacklistingTime = uint64(time.Now().UTC().Unix())
	peer.BlacklistingCause = cause

	ps.BlacklistedPeersLock.Lock()
	defer ps.BlacklistedPeersLock.Unlock()
	host.BlacklistedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
	ps.NodeConfigurationService.SetHost(host)
	return nil

}

// RemoveBlacklistedPeer removes peer from Blacklisted peer list
func (ps *PriorityStrategy) RemoveBlacklistedPeer(peer *model.Peer) error {
	if peer == nil {
		return errors.New("peer is nil")
	}
	var (
		host = ps.NodeConfigurationService.GetHost()
	)
	ps.BlacklistedPeersLock.Lock()
	defer ps.BlacklistedPeersLock.Unlock()
	delete(host.BlacklistedPeers, p2pUtil.GetFullAddressPeer(peer))
	ps.NodeConfigurationService.SetHost(host)
	return nil
}

// ======================================================
// Exposed Functions
// ======================================================

// GetAnyKnownPeer Get any known peer
func (ps *PriorityStrategy) GetAnyKnownPeer() *model.Peer {
	knownPeers := ps.NodeConfigurationService.GetHost().KnownPeers
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

// PeerBlacklist process adding peer into blacklist
func (ps *PriorityStrategy) PeerBlacklist(peer *model.Peer, cause string) error {
	if err := ps.AddToBlacklistedPeer(peer, cause); err != nil {
		ps.Logger.Warn(err.Error())
		return err
	}
	if err := ps.RemoveUnresolvedPeer(peer); err != nil {
		ps.Logger.Warn(err.Error())
		return err
	}
	if err := ps.RemoveResolvedPeer(peer); err != nil {
		ps.Logger.Warn(err.Error())
		return err
	}

	return nil
}

// PeerUnblacklist to update Peer state of peer
func (ps *PriorityStrategy) PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	if err := ps.RemoveBlacklistedPeer(peer); err != nil {
		ps.Logger.Error(err.Error())
	}
	if err := ps.AddToUnresolvedPeers([]*model.Node{peer.Info}, false); err != nil {
		ps.Logger.Warn(err.Error())
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

// SyncNodeAddressInfoTable synchronize node_address_info table by downloading and merging all addresses from peers
// Note: the node will try to rebroadcast every node address that is updated (new or updated version of an existing one)
func (ps *PriorityStrategy) SyncNodeAddressInfoTable(nodeRegistrations []*model.NodeRegistration) (map[int64]*model.NodeAddressInfo, error) {
	resolvedPeers := ps.NodeConfigurationService.GetHost().GetResolvedPeers()
	if len(resolvedPeers) < 1 {
		return nil, blocker.NewBlocker(blocker.AppErr, "SyncNodeAddressInfoTable: No resolved peers found")
	}
	var (
		finished          bool
		nodeAddressesInfo = make(map[int64]*model.NodeAddressInfo, 0)
		mutex             = &sync.Mutex{}
	)

	// Copy map to not interfere with original
	peers := make(map[string]*model.Peer)
	for key, value := range resolvedPeers {
		peers[key] = value
	}

	for len(peers) > 0 && !finished {
		peer := ps.PeerStrategyHelper.GetRandomPeerWithoutRepetition(peers, mutex)
		res, err := ps.PeerServiceClient.GetNodeAddressesInfo(peer, nodeRegistrations)
		if err != nil {
			ps.Logger.Warn(err)
			continue
		}

		// validate downloaded address list
		for _, nodeAddressInfo := range res.NodeAddressesInfo {
			if found, err := ps.NodeRegistrationService.ValidateNodeAddressInfo(nodeAddressInfo); err != nil {
				if found {
					ps.Logger.Warnf("Received invalid node address info message from peer %s:%d. Error: %s",
						peer.Info.Address,
						peer.Info.Port,
						err)
				} else {
					ps.Logger.Warnf("NodeID %d not found in db. skipping node address %s:%d send by peer %s:%d. %s",
						nodeAddressInfo.NodeID,
						nodeAddressInfo.Address,
						nodeAddressInfo.Port,
						peer.Info.Address,
						peer.Info.Port,
						err)
				}
				continue
			}
			// only keep most updated version of nodeAddressInfo for same nodeID (eg. if already downloaded an outdated record)
			curNodeAddressInfo, ok := nodeAddressesInfo[nodeAddressInfo.NodeID]
			if !ok || (ok && curNodeAddressInfo.BlockHeight < nodeAddressInfo.BlockHeight) {
				nodeAddressesInfo[nodeAddressInfo.NodeID] = nodeAddressInfo
				if err := ps.ReceiveNodeAddressInfo(nodeAddressInfo); err != nil {
					ps.Logger.Error(err)
				}
			}
		}
		// all address info have been fetched
		if len(nodeAddressesInfo) == len(nodeRegistrations) {
			finished = true
		}
	}
	return nodeAddressesInfo, nil
}

// ReceiveNodeAddressInfo receive a node address info from a peer
func (ps *PriorityStrategy) ReceiveNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error {
	// add it to nodeAddressInfo table
	updated, err := ps.NodeRegistrationService.UpdateNodeAddressInfo(nodeAddressInfo)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if updated {
		// re-broadcast updated node address info
		for _, peer := range ps.GetResolvedPeers() {
			go func(peer *model.Peer) {
				if _, err := ps.PeerServiceClient.SendNodeAddressInfo(peer, nodeAddressInfo); err != nil {
					ps.Logger.Warnf("Cannot send node address info message to peer %s:%d. Error: %s",
						peer.Info.Address,
						peer.Info.Port,
						err.Error())
				}
			}(peer)
		}
	}
	return nil
}

// UpdateOwnNodeAddressInfo check if nodeAddress in db must be updated and broadcast the new address
func (ps *PriorityStrategy) UpdateOwnNodeAddressInfo(nodeAddress string, port uint32, nodeSecretPhrase string, forceBroadcast bool) error {
	var (
		updated         bool
		nodeAddressInfo *model.NodeAddressInfo
		nodePublicKey   = crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSecretPhrase)
		resolvedPeers   = ps.GetResolvedPeers()
	)
	// first update the host address to the newest
	if ps.GetHostInfo().Address != nodeAddress {
		ps.NodeConfigurationService.SetMyAddress(nodeAddress)
	}
	nr, err := ps.NodeRegistrationService.GetNodeRegistrationByNodePublicKey(nodePublicKey)
	if nr != nil && err == nil {
		if nodeAddressInfo, err = ps.NodeRegistrationService.GenerateNodeAddressInfo(
			nr.GetNodeID(),
			nodeAddress,
			port,
			nodeSecretPhrase); err != nil {
			return err
		}
		if updated, err = ps.NodeRegistrationService.UpdateNodeAddressInfo(nodeAddressInfo); err != nil {
			return err
		}
		if len(resolvedPeers) == 0 {
			return blocker.NewBlocker(blocker.P2PPeerError,
				fmt.Sprintf("Address %s:%d cannot be broadcast now because there are no resolved peers. "+
					"Retrying later...",
					nodeAddress, port))
		}
		// broadcast, if node addressInfo has been updated
		if updated || forceBroadcast {
			for _, peer := range resolvedPeers {
				go func() {
					ps.Logger.Debugf("Broadcasting node addresses %s:%d  to %s:%d",
						nodeAddressInfo.Address,
						nodeAddressInfo.Port,
						peer.Info.Address,
						peer.Info.Port)
					if _, err := ps.PeerServiceClient.SendNodeAddressInfo(peer, nodeAddressInfo); err != nil {
						ps.Logger.Warnf("Could not send updated node address info to peer %s:%d. %s",
							peer.Info.Address,
							peer.Info.Port,
							err)
					}
				}()
			}
		}
	}
	return nil
}
