package constant

import "time"

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	MainchainSnapshotGenerationTimeout time.Duration = 5 * time.Minute // 5 minutes before including in spine block (test only!)
	// MainchainSnapshotInterval interval in mainchain blocks between snapshots
	MainchainSnapshotInterval uint32 = 1440   // 1140 mainchain blocks (test only!)
	SnapshotChunkLengthBytes  int64  = 10 ^ 6 // 1MB
)
