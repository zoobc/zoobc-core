package util

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
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
	if host.UnresolvedPeers[GetFullAddressPeer(peer)] != nil {
		delete(host.UnresolvedPeers, GetFullAddressPeer(peer))
	}
	host.Peers[GetFullAddressPeer(peer)] = peer
	return host
}

// AddToUnresolvedPeers to add incoming peers to UnresolvedPeers list
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

// PeerUnblacklist to update Peer state of peer
func PeerUnblacklist(peer *model.Peer) *model.Peer {
	peer.BlacklistingCause = ""
	peer.BlacklistingTime = 0
	if peer.State == model.PeerState_BLACKLISTED {
		peer.State = model.PeerState_NON_CONNECTED
	}
	return peer
}

func GrpcDialer(destinationPeer *model.Peer) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(GetFullAddressPeer(destinationPeer), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ServerListener(port int) net.Listener {
	serv, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return serv
}

func GetTickerTime(duration uint) *time.Ticker {
	return time.NewTicker(time.Duration(duration) * time.Second)
}
