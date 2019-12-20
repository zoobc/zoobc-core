package chaintype

import "github.com/zoobc/zoobc-core/common/constant"

// SpineChain is struct should has methods in below
type SpineChain struct{}

// GetTypeInt return the value of the chain type in int
func (*SpineChain) GetTypeInt() int32 {
	return 1
}

// GetTablePrefix return the value of current chain table prefix in the database
func (*SpineChain) GetTablePrefix() string {
	return "spine"
}

// GetChainSmithingDelayTime return the value of chain smithing delay in second
func (*SpineChain) GetChainSmithingDelayTime() int64 {
	return 20
}

func (*SpineChain) GetSmithingPeriod() int64 {
	return 60
}

// GetName return the name of the chain : used in parsing chaintype across node
func (*SpineChain) GetName() string {
	return "Spinechain"
}

// GetGenesisBlockID return the block ID of genesis block in the chain
func (*SpineChain) GetGenesisBlockID() int64 {
	return constant.SpinechainGenesisBlockID
}

func (*SpineChain) GetGenesisBlockSeed() []byte {
	return constant.SpinechainGenesisBlockSeed
}

func (*SpineChain) GetGenesisNodePublicKey() []byte {
	return constant.SpinechainGenesisNodePublicKey
}

func (*SpineChain) GetGenesisBlockTimestamp() int64 {
	return constant.SpinechainGenesisBlockTimestamp
}

func (*SpineChain) GetGenesisBlockSignature() []byte {
	return constant.SpinechainGenesisBlockSignature
}
