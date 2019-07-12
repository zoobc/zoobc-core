package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

// NewKnownPeer to generate new peer model
func NewKnownPeer(address string, port int) *model.Peer {
	peer := new(model.Peer)
	nodeInfo := new(model.Node)

	nodeInfo.Address = address
	nodeInfo.Port = uint32(port)
	peer.Info = nodeInfo
	return peer
}

// ParseKnownPeers to
func ParseKnownPeers(peers []string) ([]*model.Peer, error) {
	var (
		knownPeers = []*model.Peer{}
	)

	for _, peerData := range peers {
		peerInfo := strings.Split(peerData, ":")
		peerAddress := peerInfo[0]
		if !util.ValidateIP4(peerAddress) {
			fmt.Println("invalid ip address " + peerAddress)
		}

		peerPort, err := strconv.Atoi(peerInfo[1])
		if err != nil {
			return nil, errors.New("invalid port number in the passed knownPeers list")
		}
		knownPeers = append(knownPeers, NewKnownPeer(peerAddress, peerPort))
	}
	return knownPeers, nil
}

// GetFullAddressPeer is
func GetFullAddressPeer(peer *model.Peer) string {
	return peer.Info.Address + ":" + strconv.Itoa(int(peer.Info.Port))
}

// AddToResolvedPeer to
func AddToResolvedPeer(host *model.Host, peer *model.Peer) *model.Host {
	delete(host.UnresolvedPeers, GetFullAddressPeer(peer))
	host.Peers[GetFullAddressPeer(peer)] = peer
	return host
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list in host
func AddToUnresolvedPeers(host *model.Host, newNodes []*model.Node) *model.Host {
	hostAddress := &model.Peer{
		Info: host.Info,
	}
	for _, node := range newNodes {
		peer := &model.Peer{
			Info: node,
		}
		if host.UnresolvedPeers[GetFullAddressPeer(peer)] == nil &&
			host.Peers[GetFullAddressPeer(peer)] == nil &&
			GetFullAddressPeer(hostAddress) != GetFullAddressPeer(peer) {
			host.UnresolvedPeers[GetFullAddressPeer(peer)] = peer
		}
	}
	return host
}

// PeerUnblacklist to
func PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	if peer.State == model.PeerState_BLACKLISTED {
		peer.State = model.PeerState_NON_CONNECTED
	}
	return peer
}
