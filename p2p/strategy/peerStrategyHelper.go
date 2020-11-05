package strategy

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"sync"
)

type (
	PeerStrategyHelperInterface interface {
		GetRandomPeerWithoutRepetition(peers map[string]*model.Peer, mutex *sync.Mutex) *model.Peer
	}

	// PeerStrategyHelper helper functions shared by peer strategies
	PeerStrategyHelper struct {
	}
)

func NewPeerStrategyHelper() *PeerStrategyHelper {
	return &PeerStrategyHelper{}
}

// GetRandomPeerWithoutRepetition get a random peer from a list. the returned peer is removed from the list (peer parameter) so that,
// when the function is called again, the same peer wont' be selected.
// NOTE: this function is thread-safe and can be used with concurrent goroutines
func (ps *PeerStrategyHelper) GetRandomPeerWithoutRepetition(peers map[string]*model.Peer, mutex *sync.Mutex) *model.Peer {
	var (
		peer *model.Peer
	)
	randomIdx := int(util.GetSecureRandom()) % len(peers)
	idx := 0
	for _, knownPeer := range peers {
		if idx == randomIdx {
			peer = knownPeer
			break
		}
		idx++
	}
	mutex.Lock()
	defer mutex.Unlock()
	delete(peers, p2pUtil.GetFullAddressPeer(peer))
	return peer
}
