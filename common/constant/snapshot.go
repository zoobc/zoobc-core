package constant

const (
	// SnapshotGenerationTimeout maximum time, in seconds, allowed for a node to generate a snapshot
	SnapshotGenerationTimeout int64 = 5 * 60 // 5 minutes before including in spine block (test only!)
	// MainchainSnapshotInterval interval in mainchain blocks between snapshots
	MainchainSnapshotInterval uint32 = 100 // 1 snapshot every 3000 mainchain blocks (test only!)
)
