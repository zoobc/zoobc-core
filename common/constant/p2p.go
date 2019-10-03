package constant

const (
	// Max number of unresolved peers stored in a host
	MaxUnresolvedPeers int32 = 10 // 1000
	// Max number of connected/resolved peers stored in a host
	MaxResolvedPeers int32 = 5 // 100
	// Minimum time period in second to update a peer
	SecondsToUpdatePeersConnection int64 = 10 // 3600
	// ResolvePeersGap, interval of peer thread trying to resolve a peer (in second)
	ResolvePeersGap uint = 5
	// UpdateBlacklistedStatusGap, interval of a tread that will update the status of blacklisted node
	UpdateBlacklistedStatusGap uint = 60
	// BlacklistingPeriod, how long a peer in blaclisting status
	BlacklistingPeriod uint64 = 3600
	// ConnectPriorityPeersGap, interval of peer thread trying connect to priority peer (in second)
	ConnectPriorityPeersGap uint = 60
	// NumberOfPriorityPeersToBeAdded how many priority peers we want to add at once
	NumberOfPriorityPeersToBeAdded int = 10
	// PriorityStrategyBuildScrambleNodesGap, interval of scramble thread to build scramble node (in block height)
	PriorityStrategyBuildScrambleNodesGap uint32 = 10
	// PriorityStrategyMaxPriorityPeers, max priority peers will have
	PriorityStrategyMaxPriorityPeers = 10
)
