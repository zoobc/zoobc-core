package monitoring

import (
	"sync"
	"time"
)

// setting a big number to avoid losing count of important process
const limitProcessOwnerQueue = 10000000

var (
	processOwnerQueueHighPriority = []int{}
	processOwnerQueueLowPriority  = []int{}
	processOwnerQueueMutex        sync.Mutex
)

func startDBLockOwnerMetricsLoggingRoutine() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				logProcessOwnerQueue(0)
				logProcessOwnerQueue(1)
			}
		}
	}()

}

func getDbLockType(priorityLock int) string {
	if priorityLock > 0 {
		return "highPriority"
	}

	return "lowPriority"
}

func updateProcessOwnerQueue(priorityLock, processOwner int) {
	name := getDbLockType(priorityLock)
	if priorityLock > 0 {
		if len(processOwnerQueueHighPriority) < limitProcessOwnerQueue {
			processOwnerQueueHighPriority = append(processOwnerQueueHighPriority, processOwner)
		}

		if len(processOwnerQueueHighPriority) == 0 {
			SetDbLockBlockingOwner(name, -1)
		} else {
			SetDbLockBlockingOwner(name, processOwnerQueueHighPriority[0])
		}
	} else {
		if len(processOwnerQueueLowPriority) < limitProcessOwnerQueue {
			processOwnerQueueLowPriority = append(processOwnerQueueLowPriority, processOwner)
		}

		if len(processOwnerQueueLowPriority) == 0 {
			SetDbLockBlockingOwner(name, -1)
		} else {
			SetDbLockBlockingOwner(name, processOwnerQueueLowPriority[0])
		}
	}
}

func popProcessOwnerQueue(priorityLock int) {
	if priorityLock > 0 {
		if len(processOwnerQueueHighPriority) < limitProcessOwnerQueue {
			processOwnerQueueHighPriority = processOwnerQueueHighPriority[1:]
		}
	} else {
		if len(processOwnerQueueLowPriority) < limitProcessOwnerQueue {
			processOwnerQueueLowPriority = processOwnerQueueLowPriority[1:]
		}
	}
}

func logProcessOwnerQueue(priorityLock int) {
	name := getDbLockType(priorityLock)
	if priorityLock > 0 {
		if len(processOwnerQueueHighPriority) == 0 {
			SetDbLockBlockingOwner(name, -1)
		} else {
			SetDbLockBlockingOwner(name, processOwnerQueueHighPriority[0])
		}
	} else {
		if len(processOwnerQueueLowPriority) == 0 {
			SetDbLockBlockingOwner(name, -1)
		} else {
			SetDbLockBlockingOwner(name, processOwnerQueueLowPriority[0])
		}
	}
}
