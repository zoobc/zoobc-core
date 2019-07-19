package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

// NewHost to initialize new server node
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

// GetAnyPeer Get any random peer
func GetAnyPeer(hs *model.Host) *model.Peer {
	if len(hs.Peers) < 1 {
		return nil
	}
	randomIdx := int(util.GetSecureRandom())
	if randomIdx != 0 {
		randomIdx %= len(hs.Peers)
	}
	idx := 0
	for _, peer := range hs.Peers {
		if idx == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// GetAnyUnresolvedPeer Get any unresolved peer
func GetAnyUnresolvedPeer(hs *model.Host) *model.Peer {
	if len(hs.UnresolvedPeers) < 1 {
		return nil
	}
	randomIdx := int(util.GetSecureRandom()) % len(hs.UnresolvedPeers)
	idx := 0
	for _, peer := range hs.UnresolvedPeers {
		if idx == randomIdx {
			return peer
		}
		idx++
	}
	return nil
}

// NewKnownPeer to parse address & port into Peer structur
func NewKnownPeer(address string, port int) *model.Peer {
	peer := new(model.Peer)
	nodeInfo := new(model.Node)

	nodeInfo.Address = address
	nodeInfo.Port = uint32(port)
	peer.Info = nodeInfo
	return peer
}

// ParseKnownPeers to parse list of string peers into list of Peer structur
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

// GetFullAddressPeer to get full address of peers
func GetFullAddressPeer(peer *model.Peer) string {
	return peer.Info.Address + ":" + strconv.Itoa(int(peer.Info.Port))
}

// AddToResolvedPeer to move unresolved peer into resolved peer
func AddToResolvedPeer(host *model.Host, peer *model.Peer) *model.Host {
	delete(host.UnresolvedPeers, GetFullAddressPeer(peer))
	host.Peers[GetFullAddressPeer(peer)] = peer
	return host
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
func AddToUnresolvedPeers(host *model.Host, newNodes []*model.Node) *model.Host {
	isMaxUnresolvedPeers := HasMaxUnresolvedPeers(host)
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
			// removing a peer at random if the UnresolvedPeers has reached max
			if isMaxUnresolvedPeers {
				peer := GetAnyUnresolvedPeer(host)
				if peer != nil {
					delete(host.UnresolvedPeers, GetFullAddressPeer(peer))
				}
			}
			host.UnresolvedPeers[GetFullAddressPeer(peer)] = peer
		}

		if isMaxUnresolvedPeers {
			break
		}
	}
	return host
}

// HasMaxUnresolvedPeers checks whether the unresolved peers max has been reached
func HasMaxUnresolvedPeers(host *model.Host) bool {
	return len(host.GetPeers())+len(host.GetUnresolvedPeers()) >= constant.MaxUnresolvedPeers
}

// HasMaxConnectedPeers checks whether the connected peers max has been reached
func HasMaxConnectedPeers(host *model.Host) bool {
	return len(host.GetPeers()) >= constant.MaxConnectedPeers
}

// PeerUnblacklist to update Peer state of peer
func PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	if peer.State == model.PeerState_BLACKLISTED {
		peer.State = model.PeerState_NON_CONNECTED
	}
	return peer
}

func GetTickerTime(duration uint32) *time.Ticker {
	return time.NewTicker(time.Duration(duration) * time.Second)
}

// DisconnectPeer moves connected peer to resolved peer
// if the unresolved peer is full (maybe) it should not go to the unresolved peer
func DisconnectPeer(host *model.Host, peer *model.Peer) {
	if peer != nil {
		delete(host.Peers, GetFullAddressPeer(peer))
	}

	if !HasMaxUnresolvedPeers(host) {
		host.UnresolvedPeers[GetFullAddressPeer(peer)] = peer
	}
}

// RemovePeer removes peer from unresolved peer list
func RemoveUnresolvedPeer(host *model.Host, peer *model.Peer) {
	if peer != nil {
		delete(host.UnresolvedPeers, GetFullAddressPeer(peer))
	}
}
