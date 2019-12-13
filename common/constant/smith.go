package constant

import "math"

var (
	Two64                 = "18446744073709551616"
	MaximumBalance        = int64(10000000000)
	MaximumBlocktimeLimit = int64(67)
	MinimumBlocktimeLimit = int64(53)
	// AverageSmithingBlockHeight todo: inspect this number, this is for test only
	AverageSmithingBlockHeight = uint32(math.MaxInt32)
	MaxNumBlocksmithRewards    = int(5)
	GenerateBlockTimeoutSec    = int64(15)
	// SmithingStartTime smithing initial delay from last block
	SmithingStartTime           = int64(15) // second
	SmithingBlockCreationTime   = int64(30)
	SmithingNetworkTolerance    = int64(15)
	CumulativeDifficultyDivisor = int64(1000000)
)
