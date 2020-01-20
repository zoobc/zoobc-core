package constant

var (
	// SnapshotInterval interval in seconds between snapshots (30 days)
	SnapshotInterval = int64(3 * 60) // 1 snapshot every 5 minutes (only for testing)
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	SnapshotGenerationTimeout = int64(2 * 60) // 4 minutes before including in spine block (only for testing)

	// TODO: in production use these values
	// SnapshotInterval     = int64(1440 * 60 * 30) // 30 days
	// SnapshotGenerationTimeout = int64(1440 * 60 * 3) // 3 days
)
