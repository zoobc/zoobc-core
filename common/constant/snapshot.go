package constant

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	SnapshotGenerationTimeout int64 = 4 * 60 // 4 minutes before including in spine block (only for testing)
	// SnapshotInterval interval in mainchain blocks between snapshots
	SnapshotInterval uint32 = 10 // 1 snapshot every 10 mainchain blocks (only for testing)
)
