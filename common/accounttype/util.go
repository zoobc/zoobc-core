package accounttype

// GetAccountType returns the appropriate AccountType object based on the account type index
func NewAccountType(accTypeInt uint32, accPubKey []byte) AccountType {
	switch accTypeInt {
	case 0:
		return &ZbcAccountType{
			accountPublicKey: accPubKey,
		}
	case 1:
		return &DummyAccountType{
			accountPublicKey: accPubKey,
		}
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
