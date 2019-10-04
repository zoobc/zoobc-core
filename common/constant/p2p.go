package constant

import "github.com/spf13/viper"

var (
	// MaxUnresolvedPeers : Max number of unresolved peers
	MaxUnresolvedPeers = setMaxUnresolvedPeers() // 1000
	// MaxResolvedPeers : Max number of connected/resolved peers stored in a host
	MaxResolvedPeers = setMaxResolvedPeers() // 100
	// SecondsToUpdatePeersConnection : Minimum time period in second to update a peer
	SecondsToUpdatePeersConnection = setSecondsToUpdatePeersConnection() // 3600
	// ResolvePeersGap : interval of peer thread trying to resolve a peer (in second)
	ResolvePeersGap = setResolvePeersGap()
	// UpdateBlacklistedStatusGap : interval of a tread that will update the status of blacklisted node
	UpdateBlacklistedStatusGap = setUpdateBlacklistedStatusGap()
	// BlacklistingPeriod : how long a peer in blaclisting status
	BlacklistingPeriod = setBlacklistingPeriod()
	// ConnectPriorityPeersGap : interval of peer thread trying connect to priority peer (in second)
	ConnectPriorityPeersGap = setConnectPriorityPeersGap()
	// NumberOfPriorityPeersToBeAdded : how many priority peers we want to add at once
	NumberOfPriorityPeersToBeAdded = setNumberOfPriorityPeersToBeAdded()
)

func setMaxUnresolvedPeers() int32 {
	var MaxUnresolvedPeers int32
	if viper.GetInt32("MaxUnresolvedPeers") != 0 {
		MaxUnresolvedPeers = viper.GetInt32("MaxUnresolvedPeers")
	} else {
		MaxUnresolvedPeers = 10
	}

	return MaxUnresolvedPeers
}

func setMaxResolvedPeers() int32 {
	var MaxResolvedPeers int32
	if viper.GetInt32("MaxResolvedPeers") != 0 {
		MaxResolvedPeers = viper.GetInt32("MaxResolvedPeers")
	} else {
		MaxResolvedPeers = 5
	}

	return MaxResolvedPeers
}

func setSecondsToUpdatePeersConnection() int64 {
	var SecondsToUpdatePeersConnection int64
	if viper.GetInt64("SecondsToUpdatePeersConnection") != 0 {
		SecondsToUpdatePeersConnection = viper.GetInt64("SecondsToUpdatePeersConnection")
	} else {
		SecondsToUpdatePeersConnection = 10
	}

	return SecondsToUpdatePeersConnection
}

func setResolvePeersGap() uint {
	var ResolvePeersGap uint
	if viper.GetUint("ResolvePeersGap") != 0 {
		ResolvePeersGap = viper.GetUint("ResolvePeersGap")
	} else {
		ResolvePeersGap = 5
	}

	return ResolvePeersGap
}

func setUpdateBlacklistedStatusGap() uint {
	var UpdateBlacklistedStatusGap uint
	if viper.GetUint("UpdateBlacklistedStatusGap") != 0 {
		UpdateBlacklistedStatusGap = viper.GetUint("UpdateBlacklistedStatusGap")
	} else {
		UpdateBlacklistedStatusGap = 60
	}

	return UpdateBlacklistedStatusGap
}

func setBlacklistingPeriod() uint64 {
	var BlacklistingPeriod uint64
	if viper.GetUint64("BlacklistingPeriod") != 0 {
		BlacklistingPeriod = viper.GetUint64("BlacklistingPeriod")
	} else {
		BlacklistingPeriod = 3600
	}

	return BlacklistingPeriod
}

func setConnectPriorityPeersGap() uint {
	var ConnectPriorityPeersGap uint
	if viper.GetUint("ConnectPriorityPeersGap") != 0 {
		ConnectPriorityPeersGap = viper.GetUint("ConnectPriorityPeersGap")
	} else {
		ConnectPriorityPeersGap = 60
	}

	return ConnectPriorityPeersGap
}

func setNumberOfPriorityPeersToBeAdded() int {
	var NumberOfPriorityPeersToBeAdded int
	if viper.GetInt("NumberOfPriorityPeersToBeAdded") != 0 {
		NumberOfPriorityPeersToBeAdded = viper.GetInt("NumberOfPriorityPeersToBeAdded")
	} else {
		NumberOfPriorityPeersToBeAdded = 10
	}

	return NumberOfPriorityPeersToBeAdded
}
