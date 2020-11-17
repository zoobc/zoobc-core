package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/util"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type (
	FeedbackStrategyInterface interface {
		StartSampling(samplingInterval time.Duration)
		IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		IsCPULimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		GetSuggestedActions() map[constant.FeedbackAction]bool
		SetFeedbackVar(k string, v interface{})
		GetFeedbackVar(k string) interface{}
	}

	// AntiSpamStrategy implements an anti spam filter and it is used to reduce or inhibit number of api requests when the app
	// reaches some hard limits on concurrent processes, memory and/or cpu
	AntiSpamStrategy struct {
		CPUPercentageSamples []float64
		MemUsageSamples      []float64
		GoRoutineSamples     []int
		// FeedbackVars variables relative to feedback system that can be used by the service where AntiSpamStrategy is injected into and/or
		// for internal calculations
		FeedbackVars         map[string]interface{}
		FeedbackVarsLock     sync.RWMutex
		FeedbackActionsLock  sync.RWMutex
		FeedbackSamplingLock sync.RWMutex
		Logger               *log.Logger
	}
)

// NewAntiSpamStrategy initialize system internal variables
func NewAntiSpamStrategy(
	logger *log.Logger,
) *AntiSpamStrategy {
	return &AntiSpamStrategy{
		Logger:               logger,
		CPUPercentageSamples: make([]float64, 0, constant.FeedbackTotalSamples),
		MemUsageSamples:      make([]float64, 0, constant.FeedbackTotalSamples),
		GoRoutineSamples:     make([]int, 0, constant.FeedbackTotalSamples),
		FeedbackVars: map[string]interface{}{
			"tpsReceived":  0,
			"tpsProcessed": 0,
			"txReceived":   0,
			"txProcessed":  0,
		},
	}
}

// GetSuggestedActions given some internal state, return a list of actions that the node should undertake
func (ass *AntiSpamStrategy) GetSuggestedActions() map[constant.FeedbackAction]bool {
	panic("implement me")
}

// StartSampling main feedback service loop to collect system stats
func (ass *AntiSpamStrategy) StartSampling(samplingInterval time.Duration) {
	ticker := time.NewTicker(constant.FeedbackThreadInterval)
	tickerResetPerSecondVars := time.NewTicker(time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go func() {
				ass.FeedbackSamplingLock.RLock()
				defer ass.FeedbackSamplingLock.RUnlock()
				cpuPercentage, vm, _ := util.GetHwStats(samplingInterval)
				if len(ass.CPUPercentageSamples) > 99 {
					ass.CPUPercentageSamples = append(ass.CPUPercentageSamples[1:], cpuPercentage)
				} else {
					ass.CPUPercentageSamples = append(ass.CPUPercentageSamples, cpuPercentage)
				}
				if len(ass.MemUsageSamples) > 99 {
					ass.MemUsageSamples = append(ass.MemUsageSamples[1:], vm.UsedPercent)
				} else {
					ass.MemUsageSamples = append(ass.MemUsageSamples, vm.UsedPercent)
				}
				if len(ass.GoRoutineSamples) > 99 {
					ass.GoRoutineSamples = append(ass.GoRoutineSamples[1:], util.GetGoRoutineStats())
				} else {
					ass.GoRoutineSamples = append(ass.GoRoutineSamples, util.GetGoRoutineStats())
				}
				// STEF to test only!
				// _, _ = ass.IsGoroutineLimitReached(constant.FeedbackMinGoroutineSamples)
			}()
		case <-tickerResetPerSecondVars.C:
			// Reset feedback variables that are sampled 'per second'
			func() {
				ass.FeedbackVarsLock.RLock()
				defer ass.FeedbackVarsLock.RUnlock()
				if ass.FeedbackVars["tpsReceived"] != nil {
					monitoring.SetTpsReceived(ass.FeedbackVars["tpsReceived"].(int))
				}
				if ass.FeedbackVars["tpsProcessed"] != nil {
					monitoring.SetTpsProcessed(ass.FeedbackVars["tpsProcessed"].(int))
				}
				ass.FeedbackVars["tpsReceived"] = 0
				ass.FeedbackVars["tpsProcessed"] = 0
			}()
		case <-sigs:
			ticker.Stop()
			ass.Logger.Info("resourceSampling thread stopped")
			return
		}
	}
}

// IsGoroutineLimitReached return true if one of the limits has been reached, together with the feedback limit level (from none to critical)
func (ass *AntiSpamStrategy) IsGoroutineLimitReached(numSamples int) (limitReached bool, limitLevel constant.FeedbackLimitLevel) {
	var (
		counter          int
		sumGoRoutines    int
		avg              int
		numQueuedSamples = len(ass.GoRoutineSamples)
	)
	ass.FeedbackSamplingLock.RLock()
	defer ass.FeedbackSamplingLock.RUnlock()

	// if there are less elements in queue that the number of samples we want to compute the average from, return false
	if numQueuedSamples < numSamples {
		return false, constant.FeedbackLimitNone
	}
	for n := numQueuedSamples; n > 0; n-- {
		counter++
		if counter >= numSamples {
			break
		}
		sumGoRoutines += ass.GoRoutineSamples[n-1]
	}
	switch avg = sumGoRoutines / counter; {
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitCriticalPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitCritical
		ass.Logger.Errorf("goroutine level critical! average count for last %d samples is %d", counter, avg)
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitHighPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitHigh
		ass.Logger.Errorf("goroutine level high! average count for last %d samples is %d", counter, avg)
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitMediumPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitMedium
		ass.Logger.Errorf("goroutine level medium! average count for last %d samples is %d", counter, avg)
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitLowPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitLow
	default:
		limitLevel = constant.FeedbackLimitNone
	}
	// STEF to test only!
	ass.Logger.Errorf("goroutines (last sample): %d", ass.GoRoutineSamples[len(ass.GoRoutineSamples)-1])
	ass.Logger.Errorf("goroutines (avg for %d samples): %d", counter, avg)
	ass.Logger.Errorf("limit level: %s", limitLevel)
	return limitReached, limitLevel
}

// IsCPULimitReached to be implemented
func (ass *AntiSpamStrategy) IsCPULimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	panic("implement me")
}

// IsMemoryLimitReached to be implemented
func (ass *AntiSpamStrategy) IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	panic("implement me")
}

// SetFeedbackVar set one of the variables useful to determine internal system state
func (ass *AntiSpamStrategy) SetFeedbackVar(k string, v interface{}) {
	ass.FeedbackVarsLock.Lock()
	defer ass.FeedbackVarsLock.Unlock()
	ass.FeedbackVars[k] = v
}

// GetFeedbackVar get one of the variables useful to determine internal system state. if not set, return nil
func (ass *AntiSpamStrategy) GetFeedbackVar(k string) interface{} {
	v, ok := ass.FeedbackVars[k]
	if !ok {
		return nil
	}
	return v
}
