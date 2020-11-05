package constant

import "time"

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	MainchainSnapshotGenerationTimeout time.Duration = 10 * time.Minute // 10 minutes before including in spine block
	// MainchainSnapshotInterval interval in mainchain blocks between snapshots
	MainchainSnapshotInterval uint32 = 720             // 720 mainchain blocks (= MinRollbackHeight)
	SnapshotChunkSize         int    = int(100 * 1024) // 10 KB
	// DownloadSnapshotNumberOfRetries number of times to retry downloading failed snapshot file chunks from other peers
	DownloadSnapshotNumberOfRetries = uint32(MaxResolvedPeers)

	ShardBitLength                              = 8
	SnapshotSchedulerUnmaintainedChunksPeriod   = 3 * time.Hour // TODO: snapshotV2 will update on production
	SnapshotSchedulerUnmaintainedChunksAtHeight = 3 * MainchainSnapshotInterval
)
