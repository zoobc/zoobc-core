package util

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

const DefaultConnectionMetadata = "requester"

// NewHost to initialize new server node
func NewHost(address string, port uint32, knownPeers []*model.Peer, version, codeName string) *model.Host {
	knownPeersMap := make(map[string]*model.Peer)
	unresolvedPeersMap := make(map[string]*model.Peer)
	for _, peer := range knownPeers {
		knownPeersMap[GetFullAddressPeer(peer)] = peer
		// so that known peers and unresolved peer will have different reference of object
		newPeer := *peer
		unresolvedPeersMap[GetFullAddressPeer(peer)] = &newPeer
	}

	return &model.Host{
		Info: &model.Node{
			Address:  address,
			Port:     port,
			Version:  version,
			CodeName: codeName,
		},
		ResolvedPeers:    make(map[string]*model.Peer),
		UnresolvedPeers:  unresolvedPeersMap,
		KnownPeers:       knownPeersMap,
		BlacklistedPeers: make(map[string]*model.Peer),
		Stopped:          false,
	}
}

// NewPeer to parse address & port into Peer structur
func NewPeer(address string, port int) *model.Peer {
	return &model.Peer{
		Info: &model.Node{
			Address: address,
			Port:    uint32(port),
		},
	}
}

// ParsePeer to parse an address to a peer model
func ParsePeer(peerStr string) (*model.Peer, error) {
	peerInfo := strings.Split(peerStr, ":")
	if len(peerInfo) != 2 {
		return nil, errors.New("peer address must be provided in address:port format")
	}
	peerAddress := peerInfo[0]
	peerPort, err := strconv.Atoi(peerInfo[1])
	if err != nil {
		return nil, errors.New("invalid port number in the passed knownPeers list")
	}
	return NewPeer(peerAddress, peerPort), nil
}

// ParseKnownPeers to parse list of string peers into list of Peer structure
func ParseKnownPeers(peers []string) ([]*model.Peer, error) {
	var (
		knownPeers []*model.Peer
	)

	for _, peerData := range peers {
		peer, err := ParsePeer(peerData)
		if err != nil {
			return nil, err
		}
		knownPeers = append(knownPeers, peer)
	}
	return knownPeers, nil
}

// GetFullAddressPeer to get full address of peers
func GetFullAddressPeer(peer *model.Peer) string {
	return peer.GetInfo().GetAddress() + ":" + strconv.Itoa(int(peer.GetInfo().GetPort()))
}

// GetFullAddress to get full address of node
func GetFullAddress(node *model.Node) string {
	return node.GetAddress() + ":" + strconv.Itoa(int(node.GetPort()))
}

func GetNodeInfo(fullAddress string) *model.Node {
	var (
		splittedAddress = strings.Split(fullAddress, ":")
		node            = &model.Node{
			Address: splittedAddress[0],
			Port:    uint32(viper.GetInt("peerPort")),
		}
	)
	if len(splittedAddress) != 1 {
		port, err := strconv.ParseUint(splittedAddress[1], 10, 32)
		if err == nil {
			node.Port = uint32(port)
		}
	}
	return node
}

func ServerListener(port int) net.Listener {
	serv, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return serv
}

// GetStartIndexPriorityPeer, get first index of priority peers in scramble node
func GetStartIndexPriorityPeer(
	nodeIndex int,
	scrambledNodes *model.ScrambledNodes,
) int {
	return (nodeIndex * constant.PriorityStrategyMaxPriorityPeers) % (len(scrambledNodes.IndexNodes))
}

// GetPriorityPeersByNodeID extract a list of scrambled nodes by nodeID
func GetPriorityPeersByNodeID(
	senderPeerID int64,
	scrambledNodes *model.ScrambledNodes,
) (map[string]*model.Peer, error) {
	var (
		priorityPeers = make(map[string]*model.Peer)
		nodeIDStr     = fmt.Sprintf("%d", senderPeerID)
	)
	hostIndex := scrambledNodes.IndexNodes[nodeIDStr]
	if hostIndex == nil {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "senderNotInScrambledList")
	}
	startPeers := GetStartIndexPriorityPeer(*hostIndex, scrambledNodes)
	addedPosition := 0
	for addedPosition < constant.PriorityStrategyMaxPriorityPeers {
		var (
			peersPosition = (startPeers + addedPosition + 1) % (len(scrambledNodes.IndexNodes))
			peer          = scrambledNodes.AddressNodes[peersPosition]
			peerIDStr     = fmt.Sprintf("%d", peer.GetInfo().ID)
		)
		if priorityPeers[peerIDStr] != nil {
			break
		}
		if peerIDStr != nodeIDStr {
			priorityPeers[peerIDStr] = peer
		}
		addedPosition++
	}
	return priorityPeers, nil
}

// GetSortedPriorityPeersByNodeID extract a list of scrambled nodes by nodeID
func GetSortedPriorityPeersByNodeID(
	senderPeerID int64,
	scrambledNodes *model.ScrambledNodes,
) ([]*model.Peer, error) {
	var (
		priorityPeers       = make(map[string]*model.Peer)
		sortedPriorityPeers = make([]*model.Peer, 0)
		nodeIDStr           = fmt.Sprintf("%d", senderPeerID)
	)
	hostIndex := scrambledNodes.IndexNodes[nodeIDStr]
	if hostIndex == nil {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "senderNotInScrambledList")
	}
	startPeers := GetStartIndexPriorityPeer(*hostIndex, scrambledNodes)
	addedPosition := 0
	for addedPosition < constant.PriorityStrategyMaxPriorityPeers {
		var (
			peersPosition = (startPeers + addedPosition + 1) % (len(scrambledNodes.IndexNodes))
			peer          = scrambledNodes.AddressNodes[peersPosition]
			peerIDStr     = fmt.Sprintf("%d", peer.GetInfo().ID)
		)
		if priorityPeers[peerIDStr] != nil {
			break
		}
		if peerIDStr != nodeIDStr {
			priorityPeers[peerIDStr] = peer
			sortedPriorityPeers = append(sortedPriorityPeers, peer)
		}
		addedPosition++
	}
	return sortedPriorityPeers, nil
}

func CheckPeerCompatibility(host, peer *model.Node) error {
	if peer.GetCodeName() != host.GetCodeName() {
		return blocker.NewBlocker(blocker.P2PPeerError, "peer code name does not match")
	}
	if strings.Split(peer.GetVersion(), ".")[0] != strings.Split(host.GetVersion(), ".")[0] {
		return blocker.NewBlocker(blocker.P2PPeerError, "peer version does not match")
	}
	return nil
}
