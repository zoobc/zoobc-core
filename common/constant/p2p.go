package constant

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
)
