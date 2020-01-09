package constant

import (
	"time"
)

var (
	MaxNumBlocksmithRewards     = 5
	GenerateBlockTimeoutSec     = int64(15)
	SmithingBlockCreationTime   = int64(30)
	SmithingNetworkTolerance    = int64(15)
	CumulativeDifficultyDivisor = int64(1000000)
	CheckTimedOutBlock          = time.Duration(ConnectPriorityPeersGap) * time.Second
)
