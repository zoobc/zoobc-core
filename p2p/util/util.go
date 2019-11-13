package util

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"

	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

const DefaultConnectionMetadata = "requster"

// NewHost to initialize new server node
func NewHost(address string, port uint32, knownPeers []*model.Peer) *model.Host {
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
			Address: address,
			Port:    port,
		},
		ResolvedPeers:    make(map[string]*model.Peer),
		UnresolvedPeers:  unresolvedPeersMap,
		KnownPeers:       knownPeersMap,
		BlacklistedPeers: nil,
		Stopped:          false,
	}
}

// NewKnownPeer to parse address & port into Peer structur
func NewKnownPeer(address string, port int) *model.Peer {
	return &model.Peer{
		Info: &model.Node{
			Address: address,
			Port:    uint32(port),
		},
	}
}

// ParseKnownPeers to parse list of string peers into list of Peer structur
func ParseKnownPeers(peers []string) ([]*model.Peer, error) {
	var (
		knownPeers []*model.Peer
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

// GetPriorityPeersByNodeFullAddress, get a list peer should connect in scramble node by providing the node
// full address
func GetPriorityPeersByNodeFullAddress(
	senderFullAddress string,
	scrambledNodes *model.ScrambledNodes,
) (map[string]*model.Peer, error) {
	var (
		priorityPeers = make(map[string]*model.Peer)
	)
	hostIndex := scrambledNodes.IndexNodes[senderFullAddress]
	if hostIndex == nil {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "senderNotInScrambledList")
	}
	startPeers := GetStartIndexPriorityPeer(*hostIndex, scrambledNodes)
	addedPosition := 0
	for addedPosition < constant.PriorityStrategyMaxPriorityPeers {
		var (
			peersPosition = (startPeers + addedPosition + 1) % (len(scrambledNodes.IndexNodes))
			peer          = scrambledNodes.AddressNodes[peersPosition]
			addressPeer   = GetFullAddressPeer(peer)
		)
		if priorityPeers[addressPeer] != nil {
			break
		}
		if addressPeer != senderFullAddress {
			priorityPeers[addressPeer] = peer
		}
		addedPosition++
	}
	return priorityPeers, nil
}
