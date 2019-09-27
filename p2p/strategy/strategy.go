package strategy

/**
strategy package includes different peer to peer management strategy that we'll use
in zoobc.
*/

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	PeerExplorerStrategyInterface interface {
		Start()
		GetHostInfo() *model.Node
		GetAnyResolvedPeer() *model.Peer
		GetMorePeersHandler() (*model.Peer, error)
		GetUnresolvedPeers() map[string]*model.Peer
		GetResolvedPeers() map[string]*model.Peer
		AddToUnresolvedPeers(newNodes []*model.Node, toForce bool) error
		GetBlacklistedPeers() map[string]*model.Peer
		DisconnectPeer(peer *model.Peer)
		PeerUnblacklist(peer *model.Peer) *model.Peer
		PeerExploerListener() observer.Listener
	}
)
