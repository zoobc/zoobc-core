package feedbacksystem

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	tickerResetPerSecondVars := time.NewTicker(time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-tickerResetPerSecondVars.C:
			// Reset feedback variables that are sampled 'per second'
			func() {
				if tpsReceivedTmp := dfs.GetFeedbackVar("tpsReceivedTmp"); tpsReceivedTmp != nil {
					dfs.SetFeedbackVar("tpsReceived", tpsReceivedTmp)
					monitoring.SetTpsReceived(tpsReceivedTmp.(int))
				}
				if tpsProcessedTmp := dfs.GetFeedbackVar("tpsProcessedTmp"); tpsProcessedTmp != nil {
					dfs.SetFeedbackVar("tpsProcessed", tpsProcessedTmp)
					monitoring.SetTpsProcessed(tpsProcessedTmp.(int))
				}
				// Reset the temporary tps received/processed every second
				dfs.SetFeedbackVar("tpsReceivedTmp", 0)
				dfs.SetFeedbackVar("tpsProcessedTmp", 0)
			}()
		case <-sigs:
			tickerResetPerSecondVars.Stop()
			return
		}
	}
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

func (dfs *DummyFeedbackStrategy) IsCPULimitReached(sampleTime time.Duration) (bool, constant.FeedbackLimitLevel) {
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
