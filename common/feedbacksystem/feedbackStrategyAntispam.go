package feedbacksystem

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
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
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
		Logger           *log.Logger
	}
)

// NewAntiSpamStrategy initialize system internal variables
func NewAntiSpamStrategy(
	logger *log.Logger,
) *AntiSpamStrategy {
	return &AntiSpamStrategy{
		Logger:                      logger,
		CPUPercentageSamples:        make([]float64, 0, constant.FeedbackTotalSamples),
		MemUsageSamples:             make([]float64, 0, constant.FeedbackTotalSamples),
		GoRoutineSamples:            make([]int, 0, constant.FeedbackTotalSamples),
		RunningServerP2PAPIRequests: make([]int, 0, constant.FeedbackTotalSamples),
		RunningCliP2PAPIRequests:    make([]int, 0, constant.FeedbackTotalSamples),
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
				cpuPercentage, vm, _ := util.GetHwStats(samplingInterval)
				if len(ass.CPUPercentageSamples) >= constant.FeedbackTotalSamples {
					ass.CPUPercentageSamples = append(ass.CPUPercentageSamples[1:], cpuPercentage)
				} else {
					ass.CPUPercentageSamples = append(ass.CPUPercentageSamples, cpuPercentage)
				}
				if len(ass.MemUsageSamples) >= constant.FeedbackTotalSamples {
					ass.MemUsageSamples = append(ass.MemUsageSamples[1:], vm.UsedPercent)
				} else {
					ass.MemUsageSamples = append(ass.MemUsageSamples, vm.UsedPercent)
				}
				nGoroutines := util.GetGoRoutineStats()
				if len(ass.GoRoutineSamples) >= constant.FeedbackTotalSamples {
					ass.GoRoutineSamples = append(ass.GoRoutineSamples[1:], nGoroutines)
				} else {
					ass.GoRoutineSamples = append(ass.GoRoutineSamples, nGoroutines)
				}
				cliRequests := ass.GetFeedbackVar("P2POutgoingRequests")
				if cliRequests == nil {
					cliRequests = 0
				}
				if len(ass.RunningCliP2PAPIRequests) >= constant.FeedbackTotalSamples {
					ass.RunningCliP2PAPIRequests = append(ass.RunningCliP2PAPIRequests[1:], cliRequests.(int))
				} else {
					ass.RunningCliP2PAPIRequests = append(ass.RunningCliP2PAPIRequests, cliRequests.(int))
				}
				P2PRequests := ass.GetFeedbackVar("P2PIncomingRequests")
				if P2PRequests == nil {
					P2PRequests = 0
				}
				if len(ass.RunningServerP2PAPIRequests) >= constant.FeedbackTotalSamples {
					ass.RunningServerP2PAPIRequests = append(ass.RunningServerP2PAPIRequests[1:], P2PRequests.(int))
				} else {
					ass.RunningServerP2PAPIRequests = append(ass.RunningServerP2PAPIRequests, P2PRequests.(int))
				}
			}()
		case <-tickerResetPerSecondVars.C:
			// Reset feedback variables that are sampled 'per second'
			func() {
				if tpsReceivedTmp := ass.GetFeedbackVar("tpsReceivedTmp"); tpsReceivedTmp != nil {
					ass.SetFeedbackVar("tpsReceived", tpsReceivedTmp)
					monitoring.SetTpsReceived(tpsReceivedTmp.(int))
				}
				if tpsProcessedTmp := ass.GetFeedbackVar("tpsProcessedTmp"); tpsProcessedTmp != nil {
					ass.SetFeedbackVar("tpsProcessed", tpsProcessedTmp)
					monitoring.SetTpsProcessed(tpsProcessedTmp.(int))
				}
				// Reset the temporary tps received/processed every second
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
		sumGoRoutines    int
		avg              int
		numQueuedSamples = len(ass.GoRoutineSamples)
	)

	// if there are less elements in queue that the number of samples we want to compute the average from, return false
	if numQueuedSamples < numSamples {
		return false, constant.FeedbackLimitNone
	}
	for n := 1; n <= numSamples; n++ {
		sumGoRoutines += ass.GoRoutineSamples[len(ass.GoRoutineSamples)-n]
	}
	switch avg = sumGoRoutines / numSamples; {
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitCriticalPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitCritical
		ass.Logger.Infof("goroutine level critical! average count for last %d samples is %d", numSamples, avg)
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitHighPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitHigh
		ass.Logger.Infof("goroutine level high! average count for last %d samples is %d", numSamples, avg)
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitMediumPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitMedium
	case avg >= constant.GoRoutineHardLimit*constant.FeedbackLimitLowPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitLow
	default:
		limitLevel = constant.FeedbackLimitNone
	}
	return limitReached, limitLevel
}

func (ass *AntiSpamStrategy) IsP2PRequestLimitReached(numSamples int) (limitReached bool, limitLevel constant.FeedbackLimitLevel) {
	var (
		avg, sumIncoming,
		avgOutgoing, avgIncoming int
		// STEF test limiting by only considering incoming p2p requests (other nodes spamming)
		// sumOutgoing int
		// numQueuedSamplesOutGoing = len(ass.RunningCliP2PAPIRequests)
		numQueuedSamplesIncoming = len(ass.RunningServerP2PAPIRequests)
	)

	// if there are less elements in queue that the number of samples we want to compute the average from, return false
	if numQueuedSamplesIncoming < numSamples {
		// if numQueuedSamplesOutGoing < numSamples || numQueuedSamplesIncoming < numSamples {
		return false, constant.FeedbackLimitNone
	}
	// STEF test limiting by only considering incoming p2p requests (other nodes spamming)
	// for n := 1; n <= numSamples; n++ {
	// 	sumOutgoing += ass.RunningCliP2PAPIRequests[len(ass.RunningCliP2PAPIRequests)-n]
	// }
	// avgOutgoing = sumOutgoing / numSamples
	for n := 1; n <= numSamples; n++ {
		sumIncoming += ass.RunningServerP2PAPIRequests[len(ass.RunningServerP2PAPIRequests)-n]
	}
	avgIncoming = sumIncoming / numSamples
	switch avg = avgOutgoing + avgIncoming; {
	case avg >= constant.P2PRequestHardLimit*constant.FeedbackLimitCriticalPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitCritical
		ass.Logger.Errorf("P2PRequests level critical! average count for last %d samples is %d", numSamples, avg)
	case avg >= constant.P2PRequestHardLimit*constant.FeedbackLimitHighPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitHigh
		ass.Logger.Errorf("P2PRequests level high! average count for last %d samples is %d", numSamples, avg)
	case avg >= constant.P2PRequestHardLimit*constant.FeedbackLimitMediumPerc/100:
		limitReached = true
		limitLevel = constant.FeedbackLimitMedium
		ass.Logger.Errorf("P2PRequests level medium! average count for last %d samples is %d", numSamples, avg)
	case avg >= constant.P2PRequestHardLimit*constant.FeedbackLimitLowPerc/100:
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
	ass.FeedbackVarsLock.Lock()
	defer ass.FeedbackVarsLock.Unlock()
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

// IncrementVarCount increment k feedback map element (int) by one
func (ass *AntiSpamStrategy) IncrementVarCount(k string) interface{} {
	var (
		v        = ass.GetFeedbackVar(k)
		newCount = 1
	)
	if v != nil {
		newCount = v.(int) + 1
		ass.SetFeedbackVar(k, newCount)
	}
	return newCount
}

// DecrementVarCount decrement k feedback map element (int) by one
func (ass *AntiSpamStrategy) DecrementVarCount(k string) interface{} {
	var (
		v        = ass.GetFeedbackVar(k)
		newCount = 0
	)
	if v != nil {
		newCount = v.(int) - 1
		ass.SetFeedbackVar(k, newCount)
	}
	return newCount
}
