package constant

import "time"

type (
	// FeedbackLimitLevel limit level reached by the system (None, Low, Medium, High, Critical)
	FeedbackLimitLevel int
	// FeedbackAction action suggested by the feedback system
	FeedbackAction int
)

const (
	// FeedbackLimitLowPerc percentage of hard limit to be considered low level
	FeedbackLimitLowPerc = 30
	// FeedbackLimitLowPerc percentage of hard limit to be considered medium level
	FeedbackLimitMediumPerc = 60
	// FeedbackLimitLowPerc percentage of hard limit to be considered high level
	FeedbackLimitHighPerc = 80
	// FeedbackLimitLowPerc percentage of hard limit to be considered critical level
	FeedbackLimitCriticalPerc = 100

	// FeedbackSamplingInterval interval between sampling system metrics for feedback system
	FeedbackSamplingInterval = 2 * time.Second
	// FeedbackThreadInterval interval between a new feedback sampling is triggered (must be higher than FeedbackSamplingInterval)
	FeedbackThreadInterval = 4 * time.Second
	// FeedbackMinSamples min number of samples to calculate average of a FeedbackVar (eg. goroutines or P2PRequests) currently spawned
	FeedbackMinSamples = 4
	// FeedbackCPUSampleTime CPU usage sampling time interval
	FeedbackCPUSampleTime = 10 * time.Second
	// FeedbackTotalSamples total number of samples kept im memory
	FeedbackTotalSamples = 50
	// GoRoutineHardLimit max number of concurrent goroutine allowed
	GoRoutineHardLimit = 700
	// P2PRequestHardLimit max number of opened (running) incoming P2P api requests (tx broadcast by other peers)
	P2PRequestHardLimit = 1000
	// FeedbackLimitCPUPercentage max CPU percentage, sampled in FeedbackCPUSampleTime to trigger anti-spam filter
	FeedbackLimitCPUPercentage = 98

	FeedbackLimitNone FeedbackLimitLevel = iota
	FeedbackLimitLow
	FeedbackLimitMedium
	FeedbackLimitHigh
	FeedbackLimitCritical

	FeedbackActionAllowAll FeedbackAction = iota
	FeedbackActionLimitAPIRequests
	FeedbackActionInhibitAPIRequests
	FeedbackActionLimitP2PRequests
	FeedbackActionInhibitP2PRequests
	FeedbackActionLimitGoroutines
)
