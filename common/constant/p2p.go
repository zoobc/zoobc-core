package constant

const (
	// Max number of unresolved peers stored in a host
	MaxUnresolvedPeers = 2 // 1000
	// Max number of connected/resolved peers stored in a host
	MaxResolvedPeers = 1 // 100
	// Minimum time period in second to update a peer
	SecondsToUpdatePeersConnection = 10 // 3600
	// ResolvePeersGap, interval of peer thread trying resolve a peer (in second)
	ResolvePeersGap uint = 5
	// BlacklistingPeriod, how long a peer in blaclisting status
	BlacklistingPeriod uint64 = 3600
)
