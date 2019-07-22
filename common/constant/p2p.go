package constant

const (
	// Max number of unresolved peers stored in a host
	MaxUnresolvedPeers = 1 // 1000
	// Max number of connected/resolved peers stored in a host
	MaxConnectedPeers = 1 // 100
	// Minimum time period in second to update a peer
	SecondsToUpdatePeersConnection = 3 // 3600
	// ResolvePeersGap, interval of peer thread trying resolve a peer (in second)
	ResolvePeersGap uint = 1
	// BlacklistingPeriod, how long a peer in blaclisting status
	BlacklistingPeriod uint32 = 5
)
