package constant

import (
	"time"
)

var (
	MaxNumBlocksmithRewards     = 5
	GenerateBlockTimeoutSec     = int64(15)
	SmithingBlockCreationTime   = int64(30)
	SmithingNetworkTolerance    = int64(15)
	SmithingBlocksmithTimeGap   = int64(10)
	CumulativeDifficultyDivisor = int64(1000000)
	// BlockPoolScanPeriod define the periodic time to scan the whole block pool for legal block to persist to the chain
	BlockPoolScanPeriod = 5 * time.Second
	// TimeOutBlockWaitingTransactions is the timeout of block while waiting transactions
	TimeOutBlockWaitingTransactions = int64(2 * 60) // 2 minute
	// CheckTimedOutBlock to use in scheduler to check timedout block while waiting transaction
	CheckTimedOutBlock = 30 * time.Second
)
