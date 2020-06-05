package fee

import "github.com/zoobc/zoobc-core/common/constant"

const (
	// SendMoneyFeeConstant value of initial / constant send money fee
	SendMoneyFeeConstant     = constant.OneZBC / 100
	FeeScaleLowerConstraints = 0.5
	FeeScaleUpperConstraints = 2.0
)
