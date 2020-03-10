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
	// PriorityStrategyMaxStayedInUnresolvedPeers max time a peer can stay before being cycled out from unresolved peers
	PriorityStrategyMaxStayedInUnresolvedPeers int64 = 120
	// BlockchainsyncWaitingTime time, in seconds, to wait before start syncing the blockchain
	BlockchainsyncWaitingTime time.Duration = 5 * time.Second
	// BlockchainsyncCheckInterval time, in seconds, between checks if spine blocks have finished to be downloaded
	BlockchainsyncCheckInterval time.Duration = 3 * time.Second
	// BlockchainsyncSpineTimeout timeout, in seconds, for spine blocks to be downloaded from the network
	// download spine blocks and snapshot (if present) timeout
	BlockchainsyncSpineTimeout time.Duration = 3600 * time.Second
)
