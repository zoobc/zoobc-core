package service

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"time"
)

type (
	// DummyFeedbackStrategy implements FeedbackStrategyInterface and it is used to switch AntiSpam filter off
	DummyFeedbackStrategy struct {
	}
)

func (dfs *DummyFeedbackStrategy) StartSampling(samplingInterval time.Duration) {
	return
}

func (dfs *DummyFeedbackStrategy) GetSuggestedActions() map[constant.FeedbackAction]bool {
	return nil
}

func (dfs *DummyFeedbackStrategy) IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsCpuLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}
