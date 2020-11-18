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
}

func (dfs *DummyFeedbackStrategy) GetSuggestedActions() map[constant.FeedbackAction]bool {
	return nil
}

func (dfs *DummyFeedbackStrategy) IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsCPULimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) SetFeedbackVar(k string, v interface{}) {
}

func (dfs *DummyFeedbackStrategy) GetFeedbackVar(k string) interface{} {
	return nil
}

func (dfs *DummyFeedbackStrategy) IncrementVarCount(k string) interface{} {
	return 0
}

func (dfs *DummyFeedbackStrategy) DecrementVarCount(k string) interface{} {
	return 0
}
