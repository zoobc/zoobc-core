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

// GetChainTypes returns all chainType (useful for loops)
func GetChainTypes() map[int32]ChainType {
	var (
		mainchain  = &MainChain{}
		spinechain = &SpineChain{}
	)
	return map[int32]ChainType{
		mainchain.GetTypeInt():  mainchain,
		spinechain.GetTypeInt(): spinechain,
	}
}

// IsSpineChain validates whether a chaintype is a spinechain
func IsSpineChain(ct ChainType) bool {
	_, ok := ct.(*SpineChain)
	return ok
}
