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
	// FeedbackMinGoroutineSamples min number of samples to calculate average of goroutine currently spawned
	FeedbackMinGoroutineSamples = 4
	// FeedbackTotalSamples total number of samples kept im memory
	FeedbackTotalSamples = 50
	// GoRoutineHardLimit max number of concurrent goroutine allowed
	GoRoutineHardLimit = 700
	// P2PRequestHardLimit max number of opened (running) P2P api requests, both incoming (server) and outgoing (client)
	P2PRequestHardLimit = 400

	FeedbackLimitCPUPercentage = 95

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
