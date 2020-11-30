package feedbacksystem

import (
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
)

type (
	FeedbackStrategyInterface interface {
		StartSampling(samplingInterval time.Duration)
		IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		IsP2PRequestLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		IsCPULimitReached(sampleTime time.Duration) (bool, constant.FeedbackLimitLevel)
		IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		GetSuggestedActions() map[constant.FeedbackAction]bool
		SetFeedbackVar(k string, v interface{})
		GetFeedbackVar(k string) interface{}
		IncrementVarCount(k string) interface{}
		DecrementVarCount(k string) interface{}
	}
)
