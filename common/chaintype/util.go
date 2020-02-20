package chaintype

// GetChainType returns the appropriate chainType object based on the chain type number
func GetChainType(ctNum int32) ChainType {
	switch ctNum {
	case 0:
		return &MainChain{}
	case 1:
		return &SpineChain{}
	default:
		return nil
	}
}

// GetChainTypeCount util function to get the number of chain type (useful when looping through chain types)
func GetChainTypeCount() int {
	return 2
}
