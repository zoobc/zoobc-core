package constant

import "time"

const (
	// GetMoreBlocksDelay returns delay between GetMoreBlocksThread in seconds
	GetMoreBlocksDelay               time.Duration = 10
	BlockDownloadSegSize             uint32        = 36
	MaxResponseTime                  time.Duration = 1 * time.Minute
	DefaultNumberOfForkConfirmations int32         = 1
	PeerGetBlocksLimit               uint32        = 1440
	CommonMilestoneBlockIdsLimit     int32         = 10
	SafeBlockGap                     uint32        = 1440
	// @iltoga change this to 2 for testing snapshots
	MinRollbackBlocks              uint32 = 720 // production 720
	MaxCommonMilestoneRequestTrial uint32 = MinRollbackBlocks/uint32(CommonMilestoneBlockIdsLimit) + 1
)
