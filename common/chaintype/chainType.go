package chaintype

// ChainType interface define the different behavior of each chain
type ChainType interface {
	// GetTypeInt return the value of the chain type in int
	GetTypeInt() int32
	// GetTablePrefix return the value of current chain table prefix in the database
	GetTablePrefix() string
	// GetChainSmithingDelayTime return the value of chain smithing delay in second
	GetChainSmithingDelayTime() int64
	// GetSmithingPeriod return the value of smithing period we want
	GetSmithingPeriod() int64
	// GetName return the name of the chain : used in parsing chaintype across node
	GetName() string
	// GetGenesisBlockID return the block ID of genesis block in the chain
	GetGenesisBlockID() int64
}
