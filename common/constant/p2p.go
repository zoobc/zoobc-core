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
package constant

import "time"

const (
	// PriorityStrategyMaxPriorityPeers, max priority peers will have
	PriorityStrategyMaxPriorityPeers = 5
	// Max number of connected/resolved peers stored in a host
	MaxResolvedPeers int32 = PriorityStrategyMaxPriorityPeers * 4 // 100
	// Max number of unresolved peers stored in a host
	MaxUnresolvedPeers int32 = MaxResolvedPeers * 2 // 1000
	// Minimum time period in second to update a peer
	SecondsToUpdatePeersConnection int64 = 15 // 3600
	// ResolvePeersGap, interval of peer thread trying to resolve a peer (in second)
	ResolvePeersGap uint = 10
	// ResolvePendingPeersGap, interval of peer thread trying to resolve a peer (in second)
	ResolvePendingPeersGap uint = 60
	// UnresolvedPendingPeerExpirationTimeOffset max time in seconds a node should try to connect to/resolve a pending node address (one hour)
	UnresolvedPendingPeerExpirationTimeOffset int64 = 3600
	// UpdateNodeAddressGap, interval in seconds of peer thread to update and broadcast node dynamic address
	UpdateNodeAddressGap uint = 120
	// SyncNodeAddressGap, interval in minutes of peer thread to sync node address info table
	SyncNodeAddressGap uint = 30 // every 30 min
	// ScrambleNodesSafeHeight height before which scramble nodes are always recalculated (
	// this is to allow first nodes that bootstrap the network to update their priority peers till every node has exchanged all peer
	// node addresses)
	ScrambleNodesSafeHeight uint32 = 10
	// SyncNodeAddressDelay, delay in millis to execute send/get address info api call,
	// to make sure even if many nodes start at the same time they won't execute requests at the same time
	SyncNodeAddressDelay int = 10000
	// UpdateBlacklistedStatusGap, interval of a tread that will update the status of blacklisted node
	UpdateBlacklistedStatusGap uint = 60
	// BlacklistingPeriod, how long a peer in blaclisting status
	BlacklistingPeriod uint64 = 3600
	// ConnectPriorityPeersGapScale, the gap scale of conneting priority schedule
	ConnectPriorityPeersGapScale uint32 = 1
	// ConnectPriorityPeersGap, interval of peer thread trying connect to priority peer (in second)
	ConnectPriorityPeersGap uint32 = PriorityStrategyBuildScrambleNodesGap / ConnectPriorityPeersGapScale
	// NumberOfPriorityPeersToBeAdded how many priority peers we want to add at once
	NumberOfPriorityPeersToBeAdded int = PriorityStrategyMaxPriorityPeers / 2
	// PriorityStrategyBuildScrambleNodesGap, interval of scramble thread to build scramble node (in block height)
	PriorityStrategyBuildScrambleNodesGap uint32 = 40
	// MaxScrambleCacheRound
	MaxScrambleCacheRound = (MinRollbackBlocks / PriorityStrategyBuildScrambleNodesGap) * 2
	// PriorityStrategyMaxStayedInUnresolvedPeers max time a peer can stay before being cycled out from unresolved peers
	PriorityStrategyMaxStayedInUnresolvedPeers int64 = 120
	// BlockchainsyncWaitingTime time, in seconds, to wait before start syncing the blockchain
	BlockchainsyncWaitingTime time.Duration = 5 * time.Second
	// BlockchainsyncCheckInterval time, in seconds, between checks if spine blocks have finished to be downloaded
	BlockchainsyncCheckInterval time.Duration = 3 * time.Second
	// BlockchainsyncSpineTimeout timeout, in seconds, for spine blocks to be downloaded from the network
	// download spine blocks and snapshot (if present) timeout
	BlockchainsyncSpineTimeout time.Duration = 3600 * time.Second
	// ProofOfOriginExpirationOffset expiration offset in seconds for proof of origin response
	ProofOfOriginExpirationOffset = 10
	// P2PClientConnShortTimeout timeout in seconds for a gRpc client (p2p) connection
	P2PClientConnShortTimeout   = 2 * time.Second
	P2PClientConnDefaultTimeout = 10 * time.Second
	P2PClientConnLongTimeout    = 25 * time.Second
	P2PClientKeepAliveInterval  = 20 * time.Second
	P2PClientKeepAliveTimeout   = 10 * time.Second
	P2PServerKeepAliveInterval  = 60 * time.Second
	P2PServerKeepAliveTimeout   = 10 * time.Second
	// MaxSeverConnectionIdle is a duration for the amount of time after which an idle connection would be closed by sending a GoAway.
	// Idleness duration is defined since the most recent time the number of outstanding RPCs became zero or the connection establishment.
	MaxSeverConnectionIdle = 5 * time.Minute
)
