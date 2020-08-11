package constant

import "time"

const (
	// GetMoreBlocksDelay returns delay between GetMoreBlocksThread in seconds
	GetMoreBlocksDelay               time.Duration = 10
	BlockDownloadSegSize             uint32        = 36
	MaxResponseTime                                = 1 * time.Minute
	DefaultNumberOfForkConfirmations int32         = 1
	PeerGetBlocksLimit               uint32        = 1440
	CommonMilestoneBlockIdsLimit     int32         = 10
	SafeBlockGap                                   = MinRollbackBlocks / 2
	// @iltoga change this to 2 for testing snapshots to maintain functionality keep this value above PriorityStrategyBuildScrambleNodesGap
	MinRollbackBlocks              uint32 = 720 // production 720
	MaxCommonMilestoneRequestTrial        = MinRollbackBlocks/uint32(CommonMilestoneBlockIdsLimit) + 1
	MinimumPeersBlocksToDownload   int32  = 2
)
