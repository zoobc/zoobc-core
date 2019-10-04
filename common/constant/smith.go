package constant

import "github.com/spf13/viper"

var (
	Two64                      = setTwo64()
	MaximumBalance             = setMaximumBalance()
	InitialSmithScale          = setInitialSmithScale()
	MaxSmithScale              = setMaxSmithScale()
	MaxSmithScale2             = setMaxSmithScale2()
	MinSmithScale              = setMinSmithScale()
	MaximumBlocktimeLimit      = setMaximumBlocktimeLimit()
	MinimumBlocktimeLimit      = setMinimumBlocktimeLimit()
	SmithscaleGamma            = setSmithscaleGamma()
	AverageSmithingBlockHeight = setAverageSmithingBlockHeight()
	MaxNumBlocksmithRewards    = setMaxNumBlocksmithRewards()
)

func setTwo64() string {
	var Two64 string
	if viper.GetString("Two64") != "" {
		Two64 = viper.GetString("Two64")
	} else {
		Two64 = "18446744073709551616"
	}

	return Two64
}

func setMaximumBalance() int64 {
	var MaximumBalance int64
	if viper.GetInt64("MaximumBalance") != 0 {
		MaximumBalance = viper.GetInt64("MaximumBalance")
	} else {
		MaximumBalance = 10000000000
	}

	return MaximumBalance
}

func setInitialSmithScale() int64 {
	var InitialSmithScale int64
	if viper.GetInt64("InitialSmithScale") != 0 {
		InitialSmithScale = viper.GetInt64("InitialSmithScale")
	} else {
		InitialSmithScale = 153722867
	}

	return InitialSmithScale
}

func setMaximumBlocktimeLimit() int64 {
	var MaximumBlocktimeLimit int64
	if viper.GetInt64("MaximumBlocktimeLimit") != 0 {
		MaximumBlocktimeLimit = viper.GetInt64("MaximumBlocktimeLimit")
	} else {
		MaximumBlocktimeLimit = 67
	}

	return MaximumBlocktimeLimit
}

func setMinimumBlocktimeLimit() int64 {
	var MinimumBlocktimeLimit int64
	if viper.GetInt64("MinimumBlocktimeLimit") != 0 {
		MinimumBlocktimeLimit = viper.GetInt64("MinimumBlocktimeLimit")
	} else {
		MinimumBlocktimeLimit = 53
	}

	return MinimumBlocktimeLimit
}

func setSmithscaleGamma() int64 {
	var SmithscaleGamma int64
	if viper.GetInt64("SmithscaleGamma") != 0 {
		SmithscaleGamma = viper.GetInt64("SmithscaleGamma")
	} else {
		SmithscaleGamma = 64
	}

	return SmithscaleGamma
}

func setAverageSmithingBlockHeight() uint32 {
	var AverageSmithingBlockHeight uint32
	if viper.GetUint32("AverageSmithingBlockHeight") != 0 {
		AverageSmithingBlockHeight = viper.GetUint32("AverageSmithingBlockHeight")
	} else {
		AverageSmithingBlockHeight = 10
	}

	return AverageSmithingBlockHeight
}

func setMaxNumBlocksmithRewards() int {
	var MaxNumBlocksmithRewards int
	if viper.GetInt("MaxNumBlocksmithRewards") != 0 {
		MaxNumBlocksmithRewards = viper.GetInt("MaxNumBlocksmithRewards")
	} else {
		MaxNumBlocksmithRewards = 5
	}

	return MaxNumBlocksmithRewards
}

func setMaxSmithScale() int64 {
	var MaxSmithScale int64
	if viper.GetInt64("MaxSmithScale") != 0 {
		MaxSmithScale = viper.GetInt64("MaxSmithScale")
	} else {
		MaxSmithScale = InitialSmithScale * MaximumBalance
	}

	return MaxSmithScale
}

func setMaxSmithScale2() int64 {
	var MaxSmithScale2 int64
	if viper.GetInt64("MaxSmithScale2") != 0 {
		MaxSmithScale2 = viper.GetInt64("MaxSmithScale2")
	} else {
		MaxSmithScale2 = InitialSmithScale * 50
	}

	return MaxSmithScale2
}

func setMinSmithScale() int64 {
	var MinSmithScale int64
	if viper.GetInt64("MinSmithScale") != 0 {
		MinSmithScale = viper.GetInt64("MinSmithScale")
	} else {
		MinSmithScale = InitialSmithScale * 9 / 10
	}

	return MinSmithScale
}
