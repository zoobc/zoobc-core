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
		IncrementVarCount(k string) interface{}
		DecrementVarCount(k string) interface{}
	}

	// AntiSpamStrategy implements an anti spam filter and it is used to reduce or inhibit number of api requests when the app
	// reaches some hard limits on concurrent processes, memory and/or cpu
	AntiSpamStrategy struct {
		CPUPercentageSamples []float64
		MemUsageSamples      []float64
		GoRoutineSamples     []int
		// RunningCliP2PAPIRequests number of running client p2p api requests (outgoing p2p requests)
		RunningCliP2PAPIRequests []int
		// RunningServerP2PAPIRequests number of running server p2p api requests (incoming p2p requests)
		RunningServerP2PAPIRequests []int
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
				if cliRequests := ass.GetFeedbackVar("P2POutgoingRequests"); cliRequests != nil {
					if len(ass.RunningCliP2PAPIRequests) > 99 {
						ass.RunningCliP2PAPIRequests = append(ass.RunningCliP2PAPIRequests[1:], cliRequests.(int))
					} else {
						ass.GoRoutineSamples = append(ass.GoRoutineSamples, cliRequests.(int))
					}
				}
				if P2PRequests := ass.GetFeedbackVar("P2PIncomingRequests"); P2PRequests != nil {
					if len(ass.RunningCliP2PAPIRequests) > 99 {
						ass.RunningCliP2PAPIRequests = append(ass.RunningServerP2PAPIRequests[1:], P2PRequests.(int))
					} else {
						ass.GoRoutineSamples = append(ass.GoRoutineSamples, P2PRequests.(int))
					}
				}
			}()
		case <-tickerResetPerSecondVars.C:
			// Reset feedback variables that are sampled 'per second'
			func() {
				ass.SetFeedbackVar("tpsReceived", ass.FeedbackVars["tpsReceivedTmp"])
				ass.SetFeedbackVar("tpsProcessed", ass.FeedbackVars["tpsProcessedTmp"])
				monitoring.SetTpsReceived(ass.FeedbackVars["tpsReceived"].(int))
				monitoring.SetTpsProcessed(ass.FeedbackVars["tpsProcessed"].(int))
				ass.SetFeedbackVar("tpsReceivedTmp", 0)
				ass.SetFeedbackVar("tpsProcessedTmp", 0)
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
		ass.Logger.Debugf("goroutine level medium! average count for last %d samples is %d", counter, avg)
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitLowPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitLow
	default:
		limitLevel = constant.FeedbackLimitNone
	}
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
	ass.FeedbackVarsLock.RLock()
	defer ass.FeedbackVarsLock.RUnlock()
	ass.FeedbackVars[k] = v
}

// GetFeedbackVar get one of the variables useful to determine internal system state. if not set, return nil
func (ass *AntiSpamStrategy) GetFeedbackVar(k string) interface{} {
	ass.FeedbackVarsLock.RLock()
	defer ass.FeedbackVarsLock.RUnlock()
	v, ok := ass.FeedbackVars[k]
	if !ok {
		return nil
	}
	return v
}

func (ass *AntiSpamStrategy) IncrementVarCount(k string) interface{} {
	v := ass.GetFeedbackVar(k)
	if v != nil {
		return v.(int) + 1
	}
	return 1
}

func (ass *AntiSpamStrategy) DecrementVarCount(k string) interface{} {
	v := ass.GetFeedbackVar(k)
	if v != nil {
		return v.(int) - 1
	}
	return 1
}
