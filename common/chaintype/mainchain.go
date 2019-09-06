package chaintype

import "github.com/zoobc/zoobc-core/common/constant"

// MainChain is struct should has methods in below
type MainChain struct{}

// GetTypeInt return the value of the chain type in int
func (*MainChain) GetTypeInt() int32 {
	return 0
}

// GetTablePrefix return the value of current chain table prefix in the database
func (*MainChain) GetTablePrefix() string {
	return "main"
}

// GetChainSmithingDelayTime return the value of chain smithing delay in second
func (*MainChain) GetChainSmithingDelayTime() int64 {
	return 60
}

// GetName return the name of the chain : used in parsing chaintype across node
func (*MainChain) GetName() string {
	return "Mainchain"
}

// GetGenesisBlockID return the block ID of genesis block in the chain
func (*MainChain) GetGenesisBlockID() int64 {
	return constant.GenesisBlockID
}
