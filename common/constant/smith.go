package constant

var (
	Two64                      = "18446744073709551616"
	MaximumBalance             = int64(10000000000)
	InitialSmithScale          = int64(153722867)
	MaxSmithScale              = InitialSmithScale * MaximumBalance
	MaxSmithScale2             = InitialSmithScale * 50
	MinSmithScale              = InitialSmithScale * 9 / 10
	MaximumBlocktimeLimit      = int64(67)
	MinimumBlocktimeLimit      = int64(53)
	SmithscaleGamma            = int64(64)
	AverageSmithingBlockHeight = uint32(10)
	MaxNumBlocksmithRewards    = int(5)
	GenerateBlockTimeoutSec    = int64(15)
	CoinbaseConstant           = int64(100) * OneZBC
)
