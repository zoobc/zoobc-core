package constant

import (
	"time"
)

var (
	CoinbaseTotalDistribution        int64   = 3000000 * OneZBC // 3 million * 10^8 in production
	CoinbaseTime                     int64   = OneYear          // 15 years in production
	CoinbaseSigmoidStart             float64 = 3
	CoinbaseSigmoidEnd               float64 = 6
	CoinbaseNumberRewardsPerSecond   int64   = 1 // probably this will always be 1
	CoinbaseMaxNumberRewardsPerBlock int64   = 600

	GenerateBlockTimeoutSec     = int64(15)
	CumulativeDifficultyDivisor = int64(1000000)
	// BlockPoolScanPeriod define the periodic time to scan the whole block pool for legal block to persist to the chain
	BlockPoolScanPeriod = 5 * time.Second
	// TimeOutBlockWaitingTransactions is the timeout of block while waiting transactions
	TimeOutBlockWaitingTransactions = int64(2 * 60) // 2 minute
	// CheckTimedOutBlock to use in scheduler to check timedout block while waiting transaction
	CheckTimedOutBlock        = 30 * time.Second
	SpineChainSmithIdlePeriod = 500 * time.Millisecond
	// SpineChainSmithingPeriod intervals between spine blocks in seconds
	// reduce to 60 for testing locally (300 in production)
	SpineChainSmithingPeriod = int64(300)
	MainChainSmithIdlePeriod = 500 * time.Millisecond
	// MainChainSmithingPeriod one main block every 15 seconds + block pool delay (max +30 seconds)
	MainChainSmithingPeriod = int64(15)
	// EmptyBlockSkippedBlocksmithLimit state the number of allowed skipped blocksmith until only empty block can be generated
	// 0 will set node to always create empty block
	EmptyBlockSkippedBlocksmithLimit = int64(2) // 10 in production
	/*
		Mainchain smithing
	*/

	MainSmithingBlockCreationTime = int64(30)
	MainSmithingNetworkTolerance  = int64(15)
	MainSmithingBlocksmithTimeGap = int64(10)

	/*
		Spinechain smithing
	*/

	SpineSmithingBlockCreationTime = int64(30)
	SpineSmithingNetworkTolerance  = int64(15)
	SpineSmithingBlocksmithTimeGap = int64(10)
)
