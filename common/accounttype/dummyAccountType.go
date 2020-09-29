package accounttype

import (
	"github.com/zoobc/zoobc-core/common/crypto"
)

// DummyAccountType a dummy account type
// TODO: this is only for the sake of having at least two account type.
//  as soon as we can add a real one this will be overridden by it
type DummyAccountType struct {
	accountPublicKey []byte
}

func (dAcc *DummyAccountType) GetAccount() (uint32, []byte) {
	return dAcc.GetTypeInt(), dAcc.GetAccountPublicKey()
}

func (dAcc *DummyAccountType) GetTypeInt() uint32 {
	return 1
}

func (dAcc *DummyAccountType) GetAccountPublicKey() []byte {
	return dAcc.accountPublicKey
}

func (dAcc *DummyAccountType) GetAccountPrefix() string {
	return "DUM"
}

func (dAcc *DummyAccountType) GetName() string {
	return "Dummy"
}

func (dAcc *DummyAccountType) GetAccountLength() uint32 {
	return 32
}

func (dAcc *DummyAccountType) GetFormattedAccount() (string, error) {
	return crypto.NewEd25519Signature().GetAddressFromPublicKey(dAcc.GetAccountPrefix(), dAcc.GetAccountPublicKey())
}
