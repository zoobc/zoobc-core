package constant

import "time"

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	MainchainSnapshotGenerationTimeout time.Duration = 10 * time.Minute // 10 minutes before including in spine block
	// MainchainSnapshotInterval interval in mainchain blocks between snapshots
	MainchainSnapshotInterval uint32 = 720             // 720 mainchain blocks (= MinRollbackHeight)
	SnapshotChunkSize         int    = int(100 * 1024) // 100 KB
)
