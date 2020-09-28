package accounttype

// GetAccountType returns the appropriate AccountType object based on the account type index
func GetAccountType(accNum uint32) AccountType {
	switch accNum {
	case 0:
		return &ZbcAccountType{}
	default:
		return nil
	}
}

// GetAccountTypes returns all AccountType (useful for loops)
func GetAccountTypes() map[uint32]AccountType {
	var (
		zbcAccount   = &ZbcAccountType{}
		dummyAccount = &DummyAccountType{}
	)
	return map[uint32]AccountType{
		zbcAccount.GetTypeInt():   zbcAccount,
		dummyAccount.GetTypeInt(): dummyAccount,
	}
}

// IsZbcAccount validates whether an account type is a default account (ZBC)
func IsZbcAccount(at AccountType) bool {
	_, ok := at.(*ZbcAccountType)
	return ok
}
