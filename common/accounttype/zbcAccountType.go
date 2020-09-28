package accounttype

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
)

// ZbcAccountType the default account type
type ZbcAccountType struct {
	accBytes []byte
}

func (zAcc *ZbcAccountType) GetAccount() (uint32, []byte) {
	return zAcc.GetTypeInt(), zAcc.GetAccountBytes()
}

func (zAcc *ZbcAccountType) GetTypeInt() uint32 {
	return 0
}

func (zAcc *ZbcAccountType) GetAccountBytes() []byte {
	return zAcc.accBytes
}

func (zAcc *ZbcAccountType) GetAccountPrefix() string {
	return constant.PrefixZoobcDefaultAccount
}

func (zAcc *ZbcAccountType) GetName() string {
	return "ZooBC"
}

func (zAcc *ZbcAccountType) GetAccountLength() uint32 {
	return 32
}

func (zAcc *ZbcAccountType) GetFormattedAccount() (string, error) {
	return crypto.NewEd25519Signature().GetAddressFromPublicKey(zAcc.GetAccountPrefix(), zAcc.GetAccountBytes())
}
