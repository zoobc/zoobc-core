package contract

// ChainType interface define the different behaviour of each chain
type ChainType interface {
	// GetChainTablePrefix return the value of current chain table prefix in the database
	GetTablePrefix() string
	// GetChainForgingDelayTime return the value of chain forging delay in second
	GetChainSmithingDelayTime() uint32
	// GetName return the name of the chain : used in parsing chaintype across node
	GetName() string
	// GetGenesisBlockID return the block ID of genesis block in the chain
	GetGenesisBlockID() int64
}
