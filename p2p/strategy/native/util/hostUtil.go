package util

import (
	"time"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

// NewHost to
func NewHost(address string, port uint32, knownPeers []*model.Peer) *model.Host {
	host := new(model.Host)
	nodeInfo := new(model.Node)

	nodeInfo.Address = address
	nodeInfo.Port = port
	host.Info = nodeInfo

	knownPeersMap := make(map[string]*model.Peer)
	unresolvedPeersMap := make(map[string]*model.Peer)
	for _, peer := range knownPeers {
		knownPeersMap[GetFullAddressPeer(peer)] = peer

		// so that known peers and unresolved peer will have different reference of object
		newPeer := *peer
		unresolvedPeersMap[GetFullAddressPeer(peer)] = &newPeer
	}
	host.Peers = make(map[string]*model.Peer)
	host.KnownPeers = knownPeersMap
	host.UnresolvedPeers = unresolvedPeersMap
	return host
}

// GetAnyPeer Get any peer
func GetAnyPeer(hs *model.Host) *model.Peer {
	if len(hs.Peers) < 1 {
		return nil
	}
	randomIdx := int(util.GetSecureRandom()) % len(hs.Peers)
	idx := 0
	for _, peer := range hs.Peers {
		if idx == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

func GetTickerTime(duration uint32) *time.Ticker {
	return time.NewTicker(time.Duration(duration) * time.Second)
}
