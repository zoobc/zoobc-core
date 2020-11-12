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
	FeedbackLimitHighPerc = 90
	// FeedbackLimitLowPerc percentage of hard limit to be considered critical level
	FeedbackLimitCriticalPerc = 100

	// FeedbackSamplingInterval interval between sampling system metrics for feedback system
	FeedbackSamplingInterval = 5 * time.Second
	// GoRoutineHardLimit max number of concurrent goroutine allowed
	GoRoutineHardLimit = 10000

	FeedbackLimitNone FeedbackLimitLevel = iota
	FeedbackLimitLow
	FeedbackLimitMedium
	FeedbackLimitHigh
	FeedbackLimitCritical

	FeedbackActionAllowAll FeedbackAction = iota
	FeedbackActionLimitApiRequests
	FeedbackActionInhibitApiRequests
	FeedbackActionLimitP2PRequests
	FeedbackActionInhibitP2PRequests
	FeedbackActionLimitGoroutines
)
