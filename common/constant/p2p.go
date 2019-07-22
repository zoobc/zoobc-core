package constant

const (
	MaxUnresolvedPeers             = 1 // 1000
	MaxConnectedPeers              = 1 // 100
	SecondsToUpdatePeersConnection = 3 // 3600
	// ResolvePeersGap, interval of peer thread trying resolve a peer (in second)
	ResolvePeersGap uint = 10
	// BlacklistingPeriod, how long a peer in blaclisting status
	BlacklistingPeriod uint32 = 5
)
