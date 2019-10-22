package constant

import "time"

const (
	// SchedulerInterval interval of each job in the scheduler
	SchedulerInterval = time.Duration(500) * time.Millisecond
)
