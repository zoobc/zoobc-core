package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/util"
	"sync"
	"time"
)

type (
	FeedbackStrategyInterface interface {
		StartSampling(samplingInterval time.Duration)
		IsGoroutineLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		IsCpuLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel)
		GetSuggestedActions() map[constant.FeedbackAction]bool
	}

	// AntiSpamStrategy implements an anti spam filter and it is used to reduce or inhibit number of api requests when the app
	// reaches some hard limits on concurrent processes, memory and/or cpu
	AntiSpamStrategy struct {
		CpuPercentageSamples []float64
		MemUsageSamples      []float64
		GoRoutineSamples     []int
		FeedbackActionsLock  sync.RWMutex
		Logger               *log.Logger
	}
)

func NewAntiSpamStrategy(
	logger *log.Logger,
) *AntiSpamStrategy {
	return &AntiSpamStrategy{
		Logger:               logger,
		CpuPercentageSamples: make([]float64, 0, 100),
		MemUsageSamples:      make([]float64, 0, 100),
		GoRoutineSamples:     make([]int, 0, 100),
	}
}

func (ass *AntiSpamStrategy) GetSuggestedActions() map[constant.FeedbackAction]bool {
	panic("implement me")
}

func (ass *AntiSpamStrategy) StartSampling(samplingInterval time.Duration) {
	defer func() {
		ass.Logger.Info("resourceSampling thread stopped")
	}()

	for {
		cpuPercentage, vm, _ := util.GetHwStats(samplingInterval)
		if len(ass.CpuPercentageSamples) > 99 {
			ass.CpuPercentageSamples = append(ass.CpuPercentageSamples[1:], cpuPercentage)
		} else {
			ass.CpuPercentageSamples = append(ass.CpuPercentageSamples, cpuPercentage)
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
		_, _ = ass.IsGoroutineLimitReached(5)
		time.Sleep(samplingInterval)
	}
}

func (ass *AntiSpamStrategy) IsGoroutineLimitReached(numSamples int) (limitReached bool, limitLevel constant.FeedbackLimitLevel) {
	var (
		counter          int
		sumGoRoutines    int
		avg              int
		numQueuedSamples = len(ass.GoRoutineSamples)
	)
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
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitHighPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitHigh
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitMediumPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitMedium
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitLowPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitLow
	default:
		limitLevel = constant.FeedbackLimitNone
	}
	// FIXME: for testing only!
	log.Errorf("goroutine average in last %d samples: %d", counter, avg)
	log.Errorf("goroutine limit reached: %v, level: %d", limitReached, limitLevel)
	return limitReached, limitLevel
}

func (ass *AntiSpamStrategy) IsCpuLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	panic("implement me")
}

func (ass *AntiSpamStrategy) IsMemoryLimitReached(numSamples int) (bool, constant.FeedbackLimitLevel) {
	panic("implement me")
}
