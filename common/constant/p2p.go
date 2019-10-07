package constant

var (
	// MaxUnresolvedPeers : Max number of unresolved peers
	MaxUnresolvedPeers = SetCheckVarInt32("MaxUnresolvedPeers", 10) // 1000
	// MaxResolvedPeers : Max number of connected/resolved peers stored in a host
	MaxResolvedPeers = SetCheckVarInt32("MaxResolvedPeers", 5) // 100
	// SecondsToUpdatePeersConnection : Minimum time period in second to update a peer
	SecondsToUpdatePeersConnection = SetCheckVarInt64("SecondsToUpdatePeersConnection", 10) // 3600
	// ResolvePeersGap : interval of peer thread trying to resolve a peer (in second)
	ResolvePeersGap = SetCheckVarUint("ResolvePeersGap", 5)
	// UpdateBlacklistedStatusGap : interval of a tread that will update the status of blacklisted node
	UpdateBlacklistedStatusGap = SetCheckVarUint("UpdateBlacklistedStatusGap", 60)
	// BlacklistingPeriod : how long a peer in blaclisting status
	BlacklistingPeriod = SetCheckVarUint64("BlacklistingPeriod", 3600)
	// ConnectPriorityPeersGap : interval of peer thread trying connect to priority peer (in second)
	ConnectPriorityPeersGap = SetCheckVarUint("ConnectPriorityPeersGap", 60)
	// NumberOfPriorityPeersToBeAdded : how many priority peers we want to add at once
	NumberOfPriorityPeersToBeAdded = SetCheckVarInt("NumberOfPriorityPeersToBeAdded", 10)
	// PriorityStrategyBuildScrambleNodesGap : interval of scramble thread to build scramble node (in block height)
	PriorityStrategyBuildScrambleNodesGap = SetCheckVarUint32("PriorityStrategyBuildScrambleNodesGap", 10)
	// PriorityStrategyMaxPriorityPeers : max priority peers will have
	PriorityStrategyMaxPriorityPeers = SetCheckVarInt("PriorityStrategyMaxPriorityPeers", 5)
)
