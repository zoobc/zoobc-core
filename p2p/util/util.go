package util

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

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
	return peer.Info.Address + ":" + strconv.Itoa(int(peer.Info.Port))
}

// GetFullAddress to get full address based on address and port
func GetFullAddress(address string, port uint32) string {
	return address + ":" + strconv.Itoa(int(port))
}

func ServerListener(port int) net.Listener {
	serv, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return serv
}
