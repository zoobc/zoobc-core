package monitoring

import (
	"sync"
	"time"
)

const (
	MigrationApplyOwnerProcess                   = 1
	AddGenesisNextNodeAdmissionOwnerProcess      = 2
	AddGenesisAccountOwnerProcess                = 3
	MainPushBlockOwnerProcess                    = 4
	SpinePushBlockOwnerProcess                   = 5
	SpinePopOffToBlockOwnerProcess               = 6
	BackupMempoolsOwnerProcess                   = 7
	ProcessMempoolLaterOwnerProcess              = 8
	PostTransactionServiceOwnerProcess           = 9
	RestoreMempoolsBackupOwnerProcess            = 10
	ReceivedTransactionOwnerProcess              = 11
	DeleteExpiredMempoolTransactionsOwnerProcess = 12
	InsertAddressInfoOwnerProcess                = 13
	UpdateAddrressInfoOwnerProcess               = 14
	ConfirmNodeAddressInfoOwnerProcess           = 15
	DeletePendingNodeAddressInfoOwnerProcess     = 16
	ExpiringPendingTransactionsOwnerProcess      = 17
	GenerateReceiptsMerkleRootOwnerProcess       = 18
	InsertSnapshotPayloadToDBOwnerProcess        = 19
	CreateSpineBlockManifestOwnerProcess         = 20
	ExpiringEscrowTransactionsOwnerProcess       = 21
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
		for range ticker.C {
			logProcessOwnerQueue(0)
			logProcessOwnerQueue(1)
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
