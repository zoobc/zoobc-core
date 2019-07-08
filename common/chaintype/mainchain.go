package chaintype

type MainChain struct{}

// GetChainTablePrefix return the value of current chain table prefix in the database
func (*MainChain) GetTablePrefix() string {
	return "main"
}

// GetChainSmithingDelayTime return the value of chain smithing delay in second
func (*MainChain) GetChainSmithingDelayTime() int64 {
	return 0
}

// GetName return the name of the chain : used in parsing chaintype across node
func (*MainChain) GetName() string {
	return ""
}

// GetGenesisBlockID return the block ID of genesis block in the chain
func (*MainChain) GetGenesisBlockID() int64 {
	return 0
}
