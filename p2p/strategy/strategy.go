package strategy

/**
strategy package includes different peer to peer management strategy that we'll use
in zoobc.
*/

import "github.com/zoobc/zoobc-core/common/model"

type (
	PeerExplorerStrategyInterface interface {
		GetHostInfo() *model.Node
		GetAnyResolvedPeer() *model.Peer
		GetMorePeersHandler() (*model.Peer, error)
		GetUnresolvedPeers() map[string]*model.Peer
		GetResolvedPeers() map[string]*model.Peer
		AddToUnresolvedPeer(peer *model.Peer) error
		AddToResolvedPeer(peer *model.Peer) error
		AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error
		RemoveResolvedPeer(peer *model.Peer) error
		RemoveUnresolvedPeer(peer *model.Peer) error
		GetBlacklistedPeers() map[string]*model.Peer
		RemoveBlacklistedPeer(peer *model.Peer) error
		ResolvePeers()
		UpdateResolvedPeers()
		DisconnectPeer(peer *model.Peer)
		PeerUnblacklist(peer *model.Peer) *model.Peer
	}
)
