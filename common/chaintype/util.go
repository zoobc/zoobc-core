package chaintype

import "github.com/zoobc/zoobc-core/common/contract"

// GetChainType returns the appropriate chainType object based on the chain type number
func GetChainType(ctNum int32) contract.ChainType {
	switch ctNum {
	case 0:
		return &MainChain{}
	default:
		return &MainChain{}
	}
}

func test() {

}
