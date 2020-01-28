package constant

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	SnapshotGenerationTimeout int64 = 10 * 60 // 10 minutes before including in spine block
	// MainchainSnapshotInterval interval in mainchain blocks between snapshots
	MainchainSnapshotInterval uint32 = 3000 // 1 snapshot every 3000 mainchain blocks
)
