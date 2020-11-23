package feedbacksystem

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"sync"
	"time"
)

type (
	// DummyFeedbackStrategy implements FeedbackStrategyInterface and it is used to switch AntiSpam filter off
	DummyFeedbackStrategy struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
)

// DummyFeedbackStrategy initialize system internal variables
func NewDummyFeedbackStrategy() *DummyFeedbackStrategy {
	return &DummyFeedbackStrategy{
		FeedbackVars: map[string]interface{}{
			"tpsReceived":         0,
			"tpsReceivedTmp":      0,
			"tpsProcessed":        0,
			"tpsProcessedTmp":     0,
			"txReceived":          0,
			"txProcessed":         0,
			"P2PIncomingRequests": 0,
			"P2POutgoingRequests": 0,
		},
	}
}

func (dfs *DummyFeedbackStrategy) StartSampling(samplingInterval time.Duration) {
}

func (dfs *DummyFeedbackStrategy) GetSuggestedActions() map[constant.FeedbackAction]bool {
	return nil
}

func (dfs *DummyFeedbackStrategy) IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsP2PRequestLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsCPULimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	return false, constant.FeedbackLimitNone
}

func (dfs *DummyFeedbackStrategy) SetFeedbackVar(k string, v interface{}) {
	dfs.FeedbackVarsLock.Lock()
	defer dfs.FeedbackVarsLock.Unlock()
	dfs.FeedbackVars[k] = v
}

func (dfs *DummyFeedbackStrategy) GetFeedbackVar(k string) interface{} {
	dfs.FeedbackVarsLock.RLock()
	defer dfs.FeedbackVarsLock.RUnlock()
	v, ok := dfs.FeedbackVars[k]
	if !ok {
		return nil
	}
	return v
}

func (dfs *DummyFeedbackStrategy) IncrementVarCount(k string) interface{} {
	var (
		v        = dfs.GetFeedbackVar(k)
		newCount = 1
	)
	if v != nil {
		newCount = v.(int) + 1
		dfs.SetFeedbackVar(k, newCount)
	}
	return newCount
}

func (dfs *DummyFeedbackStrategy) DecrementVarCount(k string) interface{} {
	var (
		v        = dfs.GetFeedbackVar(k)
		newCount = 0
	)
	if v != nil {
		newCount = v.(int) - 1
		dfs.SetFeedbackVar(k, newCount)
	}
	return newCount
}
