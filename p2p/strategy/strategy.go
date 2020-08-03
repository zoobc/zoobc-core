package strategy

/**
strategy package includes different peer to peer management strategy that we'll use
in zoobc.
*/

import (
	"context"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PeerExplorerStrategyInterface interface {
		Start()
		GetHostInfo() *model.Node
		GetAnyResolvedPeer() *model.Peer
		GetMorePeersHandler() (*model.Peer, error)
		GetUnresolvedPeers() map[string]*model.Peer
		GetResolvedPeers() map[string]*model.Peer
		GetPriorityPeers() map[string]*model.Peer
		GetPriorityPeersByFullAddress(priorityPeers map[string]*model.Peer) (priorityPeersByAddr map[string]*model.Peer)
		AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error
		GetBlacklistedPeers() map[string]*model.Peer
		PeerBlacklist(peer *model.Peer, cause string) error
		DisconnectPeer(peer *model.Peer)
		PeerUnblacklist(peer *model.Peer) *model.Peer
		ValidateRequest(ctx context.Context) bool
		SyncNodeAddressInfoTable(nodeRegistrations []*model.NodeRegistration) (map[int64]*model.NodeAddressInfo, error)
		ReceiveNodeAddressInfo(nodeAddressInfo *model.NodeAddressInfo) error
		UpdateOwnNodeAddressInfo(nodeAddress string, port uint32, forceBroadcast bool) error
		GenerateProofOfOrigin(challengeMessage []byte, timestamp int64, secretPhrase string) *model.ProofOfOrigin
	}
)
