package chaintype

import "time"

// ChainType interface define the different behavior of each chain
type (
	ChainType interface {
		// GetTypeInt return the value of the chain type in int
		GetTypeInt() int32
		// GetTablePrefix return the value of current chain table prefix in the database
		GetTablePrefix() string
		// GetSmithingPeriod return the value of smithing period we want
		GetSmithingPeriod() int64
		// GetName return the name of the chain : used in parsing chaintype across node
		GetName() string
		// GetGenesisBlockID return the block ID of genesis block in the chain
		GetGenesisBlockID() int64

		GetGenesisBlockSeed() []byte
		GetGenesisNodePublicKey() []byte
		GetGenesisBlockTimestamp() int64
		GetGenesisBlockSignature() []byte
		// HasTransactions true if this chain type implements transactions (thus has a mempool)
		HasTransactions() bool
		// HasSnapshots true if this chain type implements snapshots
		HasSnapshots() bool
		// If HasSnapshot is true, this must return the interval, in blocks, the snapshot has to be taken
		// If HasSnapshot is false, this will return zero
		GetSnapshotInterval() uint32
		// If HasSnapshot is true, this returns the seconds to pass, from the snapshot's process start (a block's timestamp),
		// before considering the snapshot's expired (= snapshot's process timeout)
		// If HasSnapshot is false, this will return zero
		GetSnapshotGenerationTimeout() time.Duration
	}
)
