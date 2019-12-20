package constant

var (
	MaxNumBlocksmithRewards = 5
	GenerateBlockTimeoutSec = int64(15)
	// SmithingStartTime smithing initial delay from last block
	SmithingStartTime           = int64(15) // second
	SmithingStartTimeSpine      = int64(30) // second
	SmithingBlockCreationTime   = int64(30)
	SmithingNetworkTolerance    = int64(15)
	CumulativeDifficultyDivisor = int64(1000000)
)
