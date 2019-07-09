package chaintype

// MainChain is struct should has methods in below
type MainChain struct{}

// GetTablePrefix return the value of current chain table prefix in the database
func (*MainChain) GetTablePrefix() string {
	return "main"
}

// GetChainSmithingDelayTime return the value of chain smithing delay in second
func (*MainChain) GetChainSmithingDelayTime() int64 {
	return 6
}

// GetName return the name of the chain : used in parsing chaintype across node
func (*MainChain) GetName() string {
	return ""
}

// GetGenesisBlockID return the block ID of genesis block in the chain
func (*MainChain) GetGenesisBlockID() int64 {
	return 8326926047637383911
}
