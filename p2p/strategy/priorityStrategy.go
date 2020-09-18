package strategy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
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
		BlockMainService         coreService.BlockServiceInterface
		ResolvedPeersLock        sync.RWMutex
		UnresolvedPeersLock      sync.RWMutex
		BlacklistedPeersLock     sync.RWMutex
		MaxUnresolvedPeers       int32
		MaxResolvedPeers         int32
		Logger                   *log.Logger
		PeerStrategyHelper       PeerStrategyHelperInterface
		Signature                crypto.SignatureInterface
		ScrambleNodeService      coreService.ScrambleNodeServiceInterface
		// PendingNodeAddresses map containing node full address -> timestamp of last time the node tried to connect to that address
		NodeAddressesLastTryConnect     map[string]int64
		NodeAddressesLastTryConnectLock sync.RWMutex
	}
)

func NewPriorityStrategy(
	peerServiceClient client.PeerServiceClientInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	blockMainService coreService.BlockServiceInterface,
	logger *log.Logger,
	peerStrategyHelper PeerStrategyHelperInterface,
	nodeConfigurationService coreService.NodeConfigurationServiceInterface,
	blockchainStatusService coreService.BlockchainStatusServiceInterface,
	signature crypto.SignatureInterface,
	scrambleNodeService coreService.ScrambleNodeServiceInterface,
) *PriorityStrategy {
	return &PriorityStrategy{
		BlockchainStatusService:     blockchainStatusService,
		NodeConfigurationService:    nodeConfigurationService,
		PeerServiceClient:           peerServiceClient,
		NodeRegistrationService:     nodeRegistrationService,
		BlockMainService:            blockMainService,
		MaxUnresolvedPeers:          constant.MaxUnresolvedPeers,
		MaxResolvedPeers:            constant.MaxResolvedPeers,
		Logger:                      logger,
		PeerStrategyHelper:          peerStrategyHelper,
		Signature:                   signature,
		ScrambleNodeService:         scrambleNodeService,
		NodeAddressesLastTryConnect: map[string]int64{},
	}
}

// Start method to start threads which mean goroutines for PriorityStrategy
func (ps *PriorityStrategy) Start() {
	monitoring.SetUnresolvedPeersCount(len(ps.NodeConfigurationService.GetHost().UnresolvedPeers))

	// start p2p process threads
	go ps.ResolvePeersThread()
	go ps.GetMorePeersThread()
	go ps.UpdateBlacklistedStatusThread()
	go ps.ConnectPriorityPeersThread()
	time.Sleep(2 * time.Second)
	go ps.UpdateNodeAddressThread()
	// wait until there is at least one connected (resolved) peer we can communicate to before sync the address info table
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

// getPriorityPeersByFullAddress build a priority peers map with only peers that have and address
func (ps *PriorityStrategy) GetPriorityPeersByFullAddress(priorityPeers map[string]*model.Peer) map[string]*model.Peer {
	var priorityPeersByAddr = make(map[string]*model.Peer)
	for _, pp := range priorityPeers {
		if pp.GetInfo().Address != "" && pp.GetInfo().Port != 0 {
			priorityPeersByAddr[p2pUtil.GetFullAddress(pp.GetInfo())] = pp
		}
	}
	return priorityPeersByAddr
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
		priorityPeers                = ps.GetPriorityPeersByFullAddress(ps.GetPriorityPeers())
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
		priorityPeers = make(map[string]*model.Peer)
		host          = ps.NodeConfigurationService.GetHost()
	)
	lastBlock, err := ps.BlockMainService.GetLastBlock()
	if err != nil {
		return priorityPeers
	}
	scrambledNodes, err := ps.ScrambleNodeService.GetScrambleNodesByHeight(lastBlock.Height)
	if err != nil {
		return priorityPeers
	}

	if ps.ValidateScrambleNode(scrambledNodes, host.GetInfo()) {
		if hostID, err := ps.NodeConfigurationService.GetHostID(); err == nil {
			var (
				hostIDStr     = fmt.Sprintf("%d", hostID)
				hostIndex     = scrambledNodes.IndexNodes[hostIDStr]
				startPeers    = p2pUtil.GetStartIndexPriorityPeer(*hostIndex, scrambledNodes)
				addedPosition = 0
			)
			for addedPosition < constant.PriorityStrategyMaxPriorityPeers {
				var (
					peersPosition = (startPeers + addedPosition + 1) % (len(scrambledNodes.IndexNodes))
					peer          = scrambledNodes.AddressNodes[peersPosition]
					peerIDStr     = fmt.Sprintf("%d", peer.GetInfo().ID)
				)
				if priorityPeers[peerIDStr] != nil {
					break
				}
				if peerIDStr != hostIDStr {
					priorityPeers[peerIDStr] = peer
				}
				addedPosition++
			}
		}
	}
	return priorityPeers
}

// ValidateScrambleNode, check node in scramble or not
func (ps *PriorityStrategy) ValidateScrambleNode(scrambledNodes *model.ScrambledNodes, node *model.Node) bool {
	var nodeID = node.GetID()
	if nodeID == 0 {
		if node.Address != "" && node.Port != 0 {
			nais, err := ps.NodeRegistrationService.GetNodeAddressInfoFromDbByAddressPort(node.Address, node.Port,
				[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending,
					model.NodeAddressStatus_NodeAddressConfirmed})
			if err != nil || len(nais) == 0 {
				return false
			}
			for _, nai := range nais {
				naiIDStr := fmt.Sprintf("%d", nai.GetNodeID())
				if scrambledNodes.IndexNodes[naiIDStr] != nil {
					if ps.NodeConfigurationService.GetHost().GetInfo().GetAddress() == node.Address &&
						ps.NodeConfigurationService.GetHost().GetInfo().GetPort() == node.Port { // only reset host.NodeID if nai is host
						ps.NodeConfigurationService.SetHostID(nai.GetNodeID())
					}
					return true
				}
			}
		}
		return false
	}
	var nodeIDStr = fmt.Sprintf("%d", nodeID)
	return scrambledNodes.IndexNodes[nodeIDStr] != nil
}

// ValidatePriorityPeer, check peer is in priority list peer of host node
func (ps *PriorityStrategy) ValidatePriorityPeer(scrambledNodes *model.ScrambledNodes, host, peer *model.Node) bool {
	if ps.ValidateScrambleNode(scrambledNodes, host) && ps.ValidateScrambleNode(scrambledNodes, peer) {
		if host.GetID() == 0 || peer.GetID() == 0 {
			return false
		}
		priorityPeers, err := p2pUtil.GetPriorityPeersByNodeID(host.GetID(), scrambledNodes)
		if err != nil {
			return false
		}
		peerIDStr := fmt.Sprintf("%d", peer.GetID())
		return priorityPeers[peerIDStr] != nil
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
				host     = ps.NodeConfigurationService.GetHost()
				version  = md.Get("version")[0]
				codename = md.Get("codename")[0]
			)

			// validate peer compatibility
			if err := p2pUtil.CheckPeerCompatibility(host.GetInfo(), &model.Node{Version: version, CodeName: codename}); err != nil {
				return false
			}

			// get scramble node
			lastBlock, err := ps.BlockMainService.GetLastBlock()
			if err != nil {
				ps.Logger.Errorf("ValidateRequestFailGetLastBlock: %v", err)
				return false
			}
			scrambledNodes, err := ps.ScrambleNodeService.GetScrambleNodesByHeight(lastBlock.Height)
			if err != nil {
				ps.Logger.Errorf("FailGetScrambleNodesByHeight: %v", err)
				return false
			}

			// this validates always the local node's host against scramble nodes
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
	ticker1 := time.NewTicker(time.Duration(constant.ResolvePendingPeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go ps.ResolvePeers()
			go ps.UpdateResolvedPeers()
		case <-ticker1.C:
			go ps.resolvePendingAddresses()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// ResolvePeers looping unresolved peers and adding to (resolve) Peers if get response
func (ps *PriorityStrategy) ResolvePeers() {
	exceedMaxResolvedPeers := ps.GetExceedMaxResolvedPeers()
	priorityPeers := ps.GetPriorityPeersByFullAddress(ps.GetPriorityPeers())
	resolvedPeers := ps.GetResolvedPeers()
	unresolvedPeers := ps.GetUnresolvedPeers()
	var (
		removedResolvedPeers    int32
		priorityUnresolvedPeers = make(map[string]*model.Peer)
	)

	// collecting unresolved peers that are priority
	for _, unresolvedPeer := range unresolvedPeers {
		fullAddr := p2pUtil.GetFullAddressPeer(unresolvedPeer)
		if priorityPeers[fullAddr] != nil {
			// remove unresolved priority peers when already in resolved peers
			if resolvedPeers[fullAddr] != nil {
				if err := ps.RemoveUnresolvedPeer(unresolvedPeer); err != nil {
					ps.Logger.Warn(err)
				}
				continue
			}
			// override unresolved peer info, since priority peers have nodeID and node address status too
			unresolvedPeer.Info = priorityPeers[fullAddr].GetInfo()
			priorityUnresolvedPeers[fullAddr] = unresolvedPeer
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
			// unresolved peer that non priority when failed connect will remove permanently
			go ps.resolvePeer(peer, false)
			i++
		}
	}
}

// UpdateResolvedPeers use to maintaining resolved peers
func (ps *PriorityStrategy) UpdateResolvedPeers() {
	var (
		priorityPeers = ps.GetPriorityPeersByFullAddress(ps.GetPriorityPeers())
		currentTime   = time.Now().UTC()
	)
	for _, peer := range ps.GetResolvedPeers() {
		fullAddr := p2pUtil.GetFullAddressPeer(peer)
		if priorityPeers[fullAddr] != nil {
			// override resolved peer info, since priority peers have nodeID and node address status too
			peer.Info = priorityPeers[fullAddr].Info
		}
		// priority peers no need to maintenance
		if priorityPeers[fullAddr] == nil && currentTime.Unix()-peer.GetResolvingTime() >= constant.SecondsToUpdatePeersConnection {
			go ps.resolvePeer(peer, true)
		}
	}
}

// resolvePeer send request to a peer and add to resolved peer if get response
func (ps *PriorityStrategy) resolvePeer(destPeer *model.Peer, wantToKeep bool) {
	var (
		errPoorig, errNodeAddressInfo, errGetPeerInfo error
		pendingAddressesInfo, confirmedAddressesInfo  []*model.NodeAddressInfo
		poorig                                        *model.ProofOfOrigin
		peerInfoResult                                *model.GetPeerInfoResponse
		destPeerInfo                                  = destPeer.GetInfo()
		peerNodeID                                    = destPeerInfo.GetID()
	)

	// if peer nodeID = 0, check if the address is a pending  node address info
	if peerNodeID == 0 {
		nais, err := ps.NodeRegistrationService.GetNodeAddressInfoFromDbByAddressPort(
			destPeerInfo.GetAddress(),
			destPeerInfo.GetPort(),
			[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending},
		)
		if err != nil {
			ps.Logger.Warn(blocker.NewBlocker(
				blocker.P2PPeerError,
				fmt.Sprintln("resolvePeer node address info :  ", err.Error()),
			))
		}
		if len(nais) > 0 {
			nai := nais[0]
			destPeer.Info.ID = nai.GetNodeID()
			destPeer.Info.AddressStatus = nai.GetStatus()
			peerNodeID = nai.GetNodeID()
		}
	}

	// only validate priority peers addresses (the ones with nodeID)
	if peerNodeID != 0 {
		if pendingAddressesInfo, errNodeAddressInfo = ps.NodeRegistrationService.GetNodeAddressesInfoFromDb(
			[]int64{peerNodeID},
			[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending},
		); errNodeAddressInfo == nil {
			if len(pendingAddressesInfo) > 0 {
				// validate node address by asking a proof of origin to destPeer and if valid, confirm address info in db
				poorig, errPoorig = ps.PeerServiceClient.GetNodeProofOfOrigin(destPeer)
				if errPoorig == nil && poorig != nil {
					if errNodeAddressInfo = ps.NodeRegistrationService.ConfirmPendingNodeAddress(
						pendingAddressesInfo[0]); errNodeAddressInfo == nil {
						destPeer.Info.AddressStatus = model.NodeAddressStatus_NodeAddressConfirmed
					}
				} else if confirmedAddressesInfo, errNodeAddressInfo = ps.NodeRegistrationService.GetNodeAddressesInfoFromDb(
					[]int64{pendingAddressesInfo[0].GetNodeID()},
					[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressConfirmed},
				); errNodeAddressInfo == nil && len(confirmedAddressesInfo) > 0 {
					nai := confirmedAddressesInfo[0]
					// validate node address by asking a proof of origin to destPeer and if valid, confirm address info in db
					tmpDestPeer := &model.Peer{Info: &model.Node{
						ID:            nai.NodeID,
						Address:       nai.Address,
						SharedAddress: nai.Address,
						Port:          nai.Port,
						AddressStatus: nai.Status,
					}}
					poorig, errPoorig = ps.PeerServiceClient.GetNodeProofOfOrigin(tmpDestPeer)
					if errPoorig == nil && poorig != nil {
						// previous confirmed address is re-confirmed and pending address failed validation, so remove pending address
						_ = ps.NodeRegistrationService.DeletePendingNodeAddressInfo(pendingAddressesInfo[0].GetNodeID())
						// remove also unresolved peer who failed validation and change it with the new, confirmed, peer
						_ = ps.RemoveUnresolvedPeer(destPeer)
						destPeer = tmpDestPeer
						// this is an extra safe for the edge case where destPeer reports a different address status from what is in db
						destPeer.Info.AddressStatus = model.NodeAddressStatus_NodeAddressConfirmed
					}
				}
			}
		}
	}

	if poorig == nil && errPoorig == nil {
		peerInfoResult, errGetPeerInfo = ps.PeerServiceClient.GetPeerInfo(destPeer)
	}

	if errPoorig != nil || errGetPeerInfo != nil {
		// TODO: add mechanism to blacklist failing peers
		// will add into unresolved peer list if want to keep
		// otherwise remove permanently
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
		destPeer.Info.Version = peerInfoResult.GetHostInfo().GetVersion()
		destPeer.Info.CodeName = peerInfoResult.GetHostInfo().GetCodeName()
	}
	if err := ps.RemoveUnresolvedPeer(destPeer); err != nil {
		ps.Logger.Error(err.Error())
	}

	if err := ps.AddToResolvedPeer(destPeer); err != nil {
		ps.Logger.Error(err.Error())
	}
}

// resolvePendingAddresses get the list of pending addresses and resolve them
func (ps *PriorityStrategy) resolvePendingAddresses() {
	// get all pending nodeAddressInfo
	nais, err := ps.NodeRegistrationService.GetNodeAddressesInfoFromDb([]int64{},
		[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending})
	if err != nil {
		return
	}

	for _, nai := range nais {
		go func(nai *model.NodeAddressInfo) {
			ps.rndDelay()
			destPeer := &model.Peer{
				Info: &model.Node{
					ID:            nai.NodeID,
					Address:       nai.Address,
					SharedAddress: nai.Address,
					Port:          nai.Port,
				},
			}
			// validate node address by asking a proof of origin to destPeer and if valid, confirm address info in db
			if poorig, errPoorig := ps.PeerServiceClient.GetNodeProofOfOrigin(destPeer); errPoorig != nil || poorig == nil {
				fullAddress := fmt.Sprintf("%s:%d", nai.GetAddress(), nai.GetPort())
				ps.NodeAddressesLastTryConnectLock.Lock()
				defer ps.NodeAddressesLastTryConnectLock.Unlock()
				if _, ok := ps.NodeAddressesLastTryConnect[fullAddress]; !ok {
					ps.NodeAddressesLastTryConnect[fullAddress] = time.Now().Unix()
					return
				}
				// delete expired pending node addresses
				if time.Now().Unix() > ps.NodeAddressesLastTryConnect[fullAddress]+constant.UnresolvedPendingPeerExpirationTimeOffset {
					if err := ps.NodeRegistrationService.DeletePendingNodeAddressInfo(nai.GetNodeID()); err != nil {
						ps.Logger.Errorf("cannot delete pending address for node %d", nai.GetNodeID())
						return
					}
					delete(ps.NodeAddressesLastTryConnect, fullAddress)
					// remove unresolved/resolved peers with this node address, if there are any
					if err := ps.RemoveUnresolvedPeer(destPeer); err != nil {
						if err := ps.RemoveResolvedPeer(destPeer); err != nil {
							ps.Logger.Error(err)
						}
					}
					return
				}
			}
			if errNodeAddressInfo := ps.NodeRegistrationService.ConfirmPendingNodeAddress(nai); errNodeAddressInfo != nil {
				ps.Logger.Error(errNodeAddressInfo)
				return
			}
			destPeer.Info.AddressStatus = model.NodeAddressStatus_NodeAddressConfirmed
			if err := ps.AddToResolvedPeer(destPeer); err != nil {
				ps.Logger.Error(err)
			}
		}(nai)
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

// UpdateNodeAddressThread to periodically update node own's dynamic address and broadcast it
// note: if address is dynamic, if will be fetched periodically from internet, but will be re-broadcast only if it is changed
func (ps *PriorityStrategy) UpdateNodeAddressThread() {
	var (
		timeInterval uint
	)
	currentAddr, err := ps.NodeConfigurationService.GetMyAddress()
	myAddressDynamic := ps.NodeConfigurationService.IsMyAddressDynamic()
	if myAddressDynamic {
		timeInterval = constant.UpdateNodeAddressGap
	} else {
		// if can't connect to any resolved peers, wait till we hopefully get some more from the network
		timeInterval = constant.ResolvePeersGap * 2
	}
	ticker := time.NewTicker(time.Duration(timeInterval) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	if err != nil {
		ps.Logger.Warnf("Cannot get address from node. %s", err)
	}

	for {
		var (
			host = ps.NodeConfigurationService.GetHost()
			err  error
		)

		if !ps.NodeConfigurationService.IsMyAddressDynamic() {
			err = ps.UpdateOwnNodeAddressInfo(
				currentAddr,
				host.GetInfo().GetPort(),
				true,
			)
			if err != nil {
				errCasted, ok := err.(blocker.Blocker)
				if ok && errCasted.Message == "AddressAlreadyUpdatedForNode" {
					// break the loop if address is static (from config file)
					ticker.Stop()
					return
				}
			}
			// get node's public ip address from external source internet and check if differs from current one
		} else if ipAddr, err := (&util.IPUtil{}).DiscoverNodeAddress(); err == nil && ipAddr != nil {
			if err = ps.UpdateOwnNodeAddressInfo(
				ipAddr.String(),
				host.GetInfo().GetPort(),
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
	ps.rndDelay()
	// sync the registry with nodeAddressInfo from p2p network as soon as node starts,
	// to have as many priority peers as possible to download the bc from
	if err := ps.getRegistryAndSyncAddressInfoTable(); err != nil {
		ps.Logger.Error(err)
	}

	var (
		// first sync cycle: wait until the blockchain is fully downloaded, then sync
		bootstrapTicker = time.NewTicker(time.Duration(constant.ResolvePeersGap*2) * time.Second)
		// second sync life cycle: after first cycle is complete,
		// sync every hour to make sure node has an updated address info table
		ticker = time.NewTicker(time.Duration(constant.SyncNodeAddressGap) * time.Minute)
	)
	// make sure to not trigger the second ticker until the first cycle is concluded
	ticker.Stop()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		// wait until bc has finished downloading and sync nodeAddressInfo table again, to make sure we have all updated addresses
		// note: when bc is full downloaded the node should be able to validate all address info messages
		err := ps.getRegistryAndSyncAddressInfoTable()
		if err == nil && ps.BlockchainStatusService.IsFirstDownloadFinished((&chaintype.MainChain{})) {
			bootstrapTicker.Stop()
			ticker = time.NewTicker(time.Duration(constant.SyncNodeAddressGap) * time.Minute) // todo:andy-shi88 revert this later
		}
		select {
		case <-bootstrapTicker.C:
			continue
		case <-ticker.C:
			continue
		case <-sigs:
			bootstrapTicker.Stop()
			return
		}
	}
}

// getRegistryAndSyncAddressInfoTable synchronize node address info table with the network
func (ps *PriorityStrategy) getRegistryAndSyncAddressInfoTable() error {
	if nodeRegistry, err := ps.NodeRegistrationService.GetRegisteredNodes(); err != nil {
		return err
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
		host             = ps.NodeConfigurationService.GetHost()
		resolvedPeers    = ps.GetResolvedPeers()
		blacklistedPeers = ps.GetBlacklistedPeers()
		peerAddress      = p2pUtil.GetFullAddressPeer(peer)
		hostAddressInfo  = &model.Peer{
			Info: host.Info,
		}
		hostAddress = p2pUtil.GetFullAddressPeer(hostAddressInfo)
	)
	_, isInResolvedPeers := resolvedPeers[peerAddress]
	_, isInBlacklistedPeers := blacklistedPeers[peerAddress]
	if peerAddress == hostAddress || isInResolvedPeers || isInBlacklistedPeers {
		return nil
	}

	ps.UnresolvedPeersLock.Lock()
	defer func() {
		ps.UnresolvedPeersLock.Unlock()
		monitoring.SetUnresolvedPeersCount(len(host.UnresolvedPeers))
	}()
	peer.UnresolvingTime = time.Now().UTC().Unix()
	// in case it doesn't have a nodeID, check if this unresolved peer is in address info table and assign proper id and address status
	if peer.GetInfo() != nil && peer.Info.ID == 0 {
		if nais, err := ps.NodeRegistrationService.GetNodeAddressInfoFromDbByAddressPort(
			peer.Info.Address,
			peer.Info.Port,
			[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressPending, model.NodeAddressStatus_NodeAddressConfirmed},
		); err == nil {
			// the record set is ordered by status, so excluding records with 'unset' status,
			// we get pending addresses first and if not present, confirmed
			for _, nai := range nais {
				if nai.GetStatus() != model.NodeAddressStatus_Unset {
					peer.Info.ID = nai.GetNodeID()
					peer.Info.AddressStatus = nai.GetStatus()
				}
			}
		} else {
			ps.Logger.Warnln("AddToUnresolvedPeer: ", err.Error())
		}
	}
	host.UnresolvedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
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
	// don't blacklist if only have 1 known peer
	if host.KnownPeers[p2pUtil.GetFullAddressPeer(peer)] != nil && len(host.KnownPeers) == 1 {
		return nil
	}

	peer.BlacklistingTime = uint64(time.Now().UTC().Unix())
	peer.BlacklistingCause = cause

	ps.BlacklistedPeersLock.Lock()
	defer ps.BlacklistedPeersLock.Unlock()
	host.BlacklistedPeers[p2pUtil.GetFullAddressPeer(peer)] = peer
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
		finished               bool
		nodeAddressesInfo      = make(map[int64]*model.NodeAddressInfo)
		mutex                  = &sync.Mutex{}
		syncMyAddressWithPeers bool
		myAddressInfo          *model.NodeAddressInfo
		curNodeRegistration    *model.NodeRegistration
		err                    error
	)

	// Copy map to not interfere with original
	peers := make(map[string]*model.Peer)
	for key, value := range resolvedPeers {
		peers[key] = value
	}

	// if current node is registered, broadcast it back to its peers, in case they don't know its address
	if curNodeRegistration, err = ps.NodeRegistrationService.GetNodeRegistrationByNodePublicKey(ps.NodeConfigurationService.
		GetNodePublicKey()); err != nil {
		return nil, err
	} else if curNodeRegistration != nil {
		// node own address is always 'confirmed'
		if myAddressesInfo, err := ps.NodeRegistrationService.GetNodeAddressesInfoFromDb([]int64{curNodeRegistration.GetNodeID()},
			[]model.NodeAddressStatus{model.NodeAddressStatus_NodeAddressConfirmed}); err != nil {
			return nil, err
		} else if len(myAddressesInfo) > 0 {
			myAddressInfo = myAddressesInfo[0]
			syncMyAddressWithPeers = true
		} else if myAddress, err := ps.NodeConfigurationService.GetMyAddress(); err == nil {
			// in case we current node is registered but its address isn't in node_address_info table (that shouldn't happen at this point),
			// try generating a new node address info, update node db and broadcast the address
			if myPort, err := ps.NodeConfigurationService.GetMyPeerPort(); err == nil {
				if err = ps.UpdateOwnNodeAddressInfo(myAddress, myPort, true); err != nil {
					ps.Logger.Errorf("Cannot update own address info. "+
						"Other nodes might not be able to add it to their priority peers: %s", err)
				}
			}
		}
	}

	for len(peers) > 0 && !finished {
		peer := ps.PeerStrategyHelper.GetRandomPeerWithoutRepetition(peers, mutex)
		res, err := ps.PeerServiceClient.GetNodeAddressesInfo(peer, nodeRegistrations)
		if err != nil {
			ps.Logger.Warn(err)
			continue
		}

		// validate and sync downloaded address list
		// if local node address is not part of received list, send it back to the peer
		var (
			myAddressFound = false
		)
		for _, nodeAddressInfo := range res.NodeAddressesInfo {
			if myAddressInfo != nil && myAddressInfo.GetNodeID() == nodeAddressInfo.GetNodeID() {
				if nodeAddressInfo.GetAddress() == myAddressInfo.GetAddress() &&
					nodeAddressInfo.GetPort() == myAddressInfo.GetPort() {
					myAddressFound = true
				} else {
					// don't re-broadcast own address if doesn't match with the one the node has
					ps.Logger.Debugf(
						"Wrong node address for this node reported by: %s:%d",
						peer.GetInfo().GetAddress(),
						peer.GetInfo().GetPort(),
					)
					continue
				}
			}

			if alreadyUpdated, err := ps.NodeRegistrationService.ValidateNodeAddressInfo(nodeAddressInfo); err != nil || alreadyUpdated {
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

		if syncMyAddressWithPeers && !myAddressFound {
			go ps.sendAddressInfoToPeer(peer, myAddressInfo)
		}
		// all address info have been fetched
		if len(nodeAddressesInfo) == len(nodeRegistrations) {
			finished = true
		}
	}
	return nodeAddressesInfo, nil
}

// ReceiveNodeAddressInfo receive a node address info from a peer and save it to db
func (ps *PriorityStrategy) ReceiveNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error {
	// skip if received address is own address
	myAddress, errAddr := ps.NodeConfigurationService.GetMyAddress()
	myPort, errPort := ps.NodeConfigurationService.GetMyPeerPort()
	if errAddr == nil &&
		errPort == nil &&
		nodeAddressInfo.GetAddress() == myAddress &&
		nodeAddressInfo.GetPort() == myPort {
		return nil
	}

	nodeRegistry, err := ps.NodeRegistrationService.GetNodeRegistrationByNodeID(nodeAddressInfo.NodeID)
	if err != nil {
		return err
	}
	if nodeRegistry.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeRegistered) {
		// add it to nodeAddressInfo table
		if updated, _ := ps.NodeRegistrationService.UpdateNodeAddressInfo(nodeAddressInfo, model.NodeAddressStatus_NodeAddressPending); updated {
			// re-broadcast updated node address info
			for _, peer := range ps.GetResolvedPeers() {
				go ps.sendAddressInfoToPeer(peer, nodeAddressInfo)
			}
		}
	}
	// do not add to address info if still in queue or node got deleted
	return nil
}

// UpdateOwnNodeAddressInfo check if nodeAddress in db must be updated and broadcast the new address
func (ps *PriorityStrategy) UpdateOwnNodeAddressInfo(nodeAddress string, port uint32, forceBroadcast bool) error {
	if nodeAddress == "" || port == 0 {
		return blocker.NewBlocker(
			blocker.P2PPeerError,
			fmt.Sprintf("Invalid own address or port info %s:%d", nodeAddress, port),
		)
	}

	var (
		updated          bool
		nodeAddressInfo  *model.NodeAddressInfo
		nodeSecretPhrase = ps.NodeConfigurationService.GetNodeSecretPhrase()
		nodePublicKey    = crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSecretPhrase)
		resolvedPeers    = ps.GetResolvedPeers()
		hostInfo         = ps.GetHostInfo()
	)
	// first update the host address to the newest
	if hostInfo.GetAddress() != nodeAddress {
		ps.NodeConfigurationService.SetMyAddress(nodeAddress, port)
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
		// set status to 'confirmed' when updating own address
		if nr.GetRegistrationStatus() == uint32(model.NodeRegistrationState_NodeRegistered) {
			// only update own address info table if node is registered (out of queue or not removed)
			updated, err = ps.NodeRegistrationService.UpdateNodeAddressInfo(
				nodeAddressInfo,
				model.NodeAddressStatus_NodeAddressConfirmed,
			)
			if err != nil {
				ps.Logger.Warnf("cannot update nodeAddressInfo: %s", err)
			}
		}

		// broadcast, wether or not node is in queue
		if updated || forceBroadcast {
			if len(resolvedPeers) == 0 {
				return blocker.NewBlocker(blocker.P2PPeerError,
					fmt.Sprintf("Address %s:%d cannot be broadcast now because there are no resolved peers. "+
						"Retrying later...",
						nodeAddress, port))
			}
			for _, peer := range resolvedPeers {
				go ps.sendAddressInfoToPeer(peer, nodeAddressInfo)
			}
		}
	}
	return nil
}

// GenerateProofOfOrigin generate a proof of origin message from a challenge request and sign it
func (ps *PriorityStrategy) GenerateProofOfOrigin(
	challenge []byte,
	timestamp int64,
	nodeSecretPhrase string,
) *model.ProofOfOrigin {
	poorig := &model.ProofOfOrigin{
		MessageBytes: challenge,
		Timestamp:    timestamp,
	}

	poorig.Signature = ps.Signature.SignByNode(
		util.GetProofOfOriginUnsignedBytes(poorig),
		nodeSecretPhrase,
	)
	return poorig
}

// rndDelay introduce a delay of 0 to 10 seconds (steps are in millis) to avoid sending all requests at once
func (ps *PriorityStrategy) rndDelay() {
	rndTimer := util.GetFastRandom(util.GetFastRandomSeed(), constant.SyncNodeAddressDelay)
	time.Sleep(time.Duration(rndTimer) * time.Millisecond)
}

func (ps *PriorityStrategy) sendAddressInfoToPeer(peer *model.Peer, nodeAddressInfo *model.NodeAddressInfo) {
	ps.rndDelay()
	peerInfo := peer.GetInfo()
	// don't broadcast to peer with same address to be broadcast
	if peerInfo.Address == nodeAddressInfo.Address && peerInfo.Port == nodeAddressInfo.Port {
		return
	}
	ps.Logger.Debugf("Broadcasting node addresses %s:%d to %s:%d. timestamp: %d",
		nodeAddressInfo.Address,
		nodeAddressInfo.Port,
		peerInfo.Address,
		peerInfo.Port,
		time.Now().Unix())
	if _, err := ps.PeerServiceClient.SendNodeAddressInfo(peer, nodeAddressInfo); err != nil {
		ps.Logger.Warnf("Cannot send node address info message to peer %s:%d. Error: %s",
			peer.Info.Address,
			peer.Info.Port,
			err.Error())
	}
}
