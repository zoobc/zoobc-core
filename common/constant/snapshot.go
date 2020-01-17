package constant

var (
	// SnapshotInterval interval in seconds between snapshots (30 days)
	SnapshotInterval = int64(10 * 60) // 1 snapshot every 10 minutes (only for testing)
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	SnapshotGenerationTimeout = int64(30) // 30 seconds timeout before including in spine block (only for testing)

	// TODO: in production use these values
	// SnapshotInterval     = int64(1440 * 60 * 30) // 30 days
	// SnapshotGenerationTimeout = int64(1440 * 60 * 3) // 3 days
)
