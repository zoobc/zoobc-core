// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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

func CheckPeerCompatibility(host, peer *model.Node) error {
	if peer.GetCodeName() != host.GetCodeName() {
		return blocker.NewBlocker(blocker.P2PPeerError, "peer code name does not match")
	}
	if strings.Split(peer.GetVersion(), ".")[0] != strings.Split(host.GetVersion(), ".")[0] {
		return blocker.NewBlocker(blocker.P2PPeerError, "peer version does not match")
	}
	return nil
}
