package constant

import "time"

var (
	MaxNumBlocksmithRewards     = 5
	GenerateBlockTimeoutSec     = int64(15)
	SmithingBlockCreationTime   = int64(30)
	SmithingNetworkTolerance    = int64(15)
	SmithingBlocksmithTimeGap   = int64(10)
	CumulativeDifficultyDivisor = int64(1000000)
	// BlockPoolScanPeriod define the periodic time to scan the whole block pool for legal block to persist to the chain
	BlockPoolScanPeriod = 5 * time.Second
)
