package constant

import "time"

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	// STEF reduce to 1 for testing locally
	MainchainSnapshotGenerationTimeout time.Duration = 1 * time.Minute // 10 minutes before including in spine block
	// MainchainSnapshotInterval interval in mainchain blocks between snapshots
	// STEF reduce to 5 for testing locally
	MainchainSnapshotInterval uint32 = 5 // 720 mainchain blocks (= MinRollbackHeight)
	// STEF reduce to 1 for testing locally
	SnapshotChunkSize int = int(1 * 1024) // 1 KB
	// DownloadSnapshotNumberOfRetries number of times to retry downloading failed snapshot file chunks from other peers
	DownloadSnapshotNumberOfRetries uint32 = 3
)
