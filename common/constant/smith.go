package constant

var (
	Two64                      = SetCheckVarString("Two64", "18446744073709551616")
	MaximumBalance             = SetCheckVarInt64("MaximumBalance", 10000000000)
	InitialSmithScale          = SetCheckVarInt64("InitialSmithScale", 153722867)
	MaxSmithScale              = SetCheckVarInt64("MaxSmithScale", InitialSmithScale*MaximumBalance)
	MaxSmithScale2             = SetCheckVarInt64("MaxSmithScale2", InitialSmithScale*50)
	MinSmithScale              = SetCheckVarInt64("MinSmithScale", InitialSmithScale*9/10)
	MaximumBlocktimeLimit      = SetCheckVarInt64("MaxBlocktimeLimit", 67)
	MinimumBlocktimeLimit      = SetCheckVarInt64("MinBlocktimeLimit", 53)
	SmithscaleGamma            = SetCheckVarInt64("MinSmithscaleGamma", 64)
	AverageSmithingBlockHeight = SetCheckVarUint32("AverageSmithingBlockHeight", 10)
	MaxNumBlocksmithRewards    = SetCheckVarInt("MaxNumBlocksmithRewards", 5)
	GenerateBlockTimeoutSec    = SetCheckVarInt64("GenerateBlockTimeoutSec", 15)
)
