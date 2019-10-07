package constant

import (
	"time"
)

var (
	// GetMoreBlocksDelay returns delay between GetMoreBlocksThread in seconds
	GetMoreBlocksDelay               = time.Duration(SetCheckVarInt64("GetMoreBlockDelay", 10))
	BlockDownloadSegSize             = SetCheckVarUint32("BlockDownloadSegSize", 36)
	MaxResponseTime                  = time.Duration(SetCheckVarInt64("MaxResponseTime", int64(1*time.Minute)))
	DefaultNumberOfForkConfirmations = SetCheckVarInt32("DefaultNumberOfForkConfirmations", 1)
	PeerGetBlocksLimit               = SetCheckVarUint32("PeerGetBlockLimit", 1440)
	CommonMilestoneBlockIdsLimit     = SetCheckVarInt32("CommonMilestoneBlockIdsLimit", 10)
	SafeBlockGap                     = SetCheckVarUint32("SafeBlockGap", 1440)
	MinRollbackBlocks                = SetCheckVarUint32("MinRollbackBlocks", 720)
)
