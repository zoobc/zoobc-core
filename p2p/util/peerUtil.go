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
	// peer.State = NON_CONNECTED
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
			fmt.Println("Invalid port number in the passed knownPeers list")
			return nil, errors.New("Invalid port number in the passed knownPeers list")
		}
		knownPeers = append(knownPeers, NewKnownPeer(peerAddress, peerPort))
	}
	return knownPeers, nil
}

// GetFullAddressPeer is
func GetFullAddressPeer(peer *model.Peer) string {
	return peer.Info.Address + ":" + strconv.Itoa(int(peer.Info.Port))
}
