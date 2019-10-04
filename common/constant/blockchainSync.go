package constant

import (
	"time"

	"github.com/spf13/viper"
)

var (
	// GetMoreBlocksDelay returns delay between GetMoreBlocksThread in seconds
	GetMoreBlocksDelay               = time.Duration(setGetMoreBlockDelay())
	BlockDownloadSegSize             = setBlockDownloadSegSize()
	MaxResponseTime                  = time.Duration(setMaxResponseTime())
	DefaultNumberOfForkConfirmations = setDefaultNumberOfForkConfirmations()
	PeerGetBlocksLimit               = setPeerGetBlocksLimit()
	CommonMilestoneBlockIdsLimit     = setCommonMilestoneBlockIdsLimit()
	SafeBlockGap                     = setSafeBlockGap()
	MinRollbackBlocks                = setMinRollbackBlocks()
)

func setGetMoreBlockDelay() int64 {
	var GetMoreBlockDelay int64
	if viper.GetInt64("GetMoreBlockDelay") != 0 {
		GetMoreBlockDelay = viper.GetInt64("GetMoreBlockDelay")
	} else {
		GetMoreBlockDelay = 10
	}

	return GetMoreBlockDelay
}

func setBlockDownloadSegSize() uint32 {
	var BlockDownloadSegSize uint32
	if viper.GetUint32("BlockDownloadSegSize") != 0 {
		BlockDownloadSegSize = viper.GetUint32("BlockDownloadSegSize")
	} else {
		BlockDownloadSegSize = 36
	}

	return BlockDownloadSegSize
}

func setMaxResponseTime() int64 {
	var MaxResponseTime int64
	if viper.GetInt64("MaxResponseTime") != 0 {
		MaxResponseTime = viper.GetInt64("MaxResponseTime")
	} else {
		MaxResponseTime = int64(1 * time.Minute)
	}

	return MaxResponseTime
}

func setDefaultNumberOfForkConfirmations() int32 {
	var DefaultNumberOfForkConfirmations int32
	if viper.GetInt32("DefaultNumberOfForkConfirmations") != 0 {
		DefaultNumberOfForkConfirmations = viper.GetInt32("DefaultNumberOfForkConfirmations")
	} else {
		DefaultNumberOfForkConfirmations = 1
	}

	return DefaultNumberOfForkConfirmations
}

func setPeerGetBlocksLimit() uint32 {
	var PeerGetBlocksLimit uint32
	if viper.GetUint32("PeerGetBlocksLimit ") != 0 {
		PeerGetBlocksLimit = viper.GetUint32("PeerGetBlocksLimit")
	} else {
		PeerGetBlocksLimit = 1440
	}

	return PeerGetBlocksLimit
}

func setCommonMilestoneBlockIdsLimit() int32 {
	var CommonMilestoneBlockIdsLimit int32
	if viper.GetInt32("CommonMilestoneBlockIdsLimit ") != 0 {
		CommonMilestoneBlockIdsLimit = viper.GetInt32("CommonMilestoneBlockIdsLimit")
	} else {
		CommonMilestoneBlockIdsLimit = 10
	}

	return CommonMilestoneBlockIdsLimit
}

func setSafeBlockGap() uint32 {
	var SafeBlockGap uint32
	if viper.GetUint32("SafeBlockGap ") != 0 {
		SafeBlockGap = viper.GetUint32("SafeBlockGap")
	} else {
		SafeBlockGap = 1440
	}

	return SafeBlockGap
}

func setMinRollbackBlocks() uint32 {
	var MinRollbackBlocks uint32
	if viper.GetUint32("MinRollbackBlocks ") != 0 {
		MinRollbackBlocks = viper.GetUint32("MinRollbackBlocks")
	} else {
		MinRollbackBlocks = 720
	}

	return MinRollbackBlocks
}
