package accounttype

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/constant"
)

// ZbcAccountType the default account type
type ZbcAccountType struct {
	accountPublicKey []byte
}

func (zAcc *ZbcAccountType) GetAccount() (uint32, []byte) {
	return zAcc.GetTypeInt(), zAcc.GetAccountPublicKey()
}

func (zAcc *ZbcAccountType) GetTypeInt() uint32 {
	return 0
}

func (zAcc *ZbcAccountType) GetAccountPublicKey() []byte {
	return zAcc.accountPublicKey
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

func (zAcc *ZbcAccountType) IsEqual(acc AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), zAcc.GetAccountPublicKey()) && acc.GetTypeInt() == zAcc.GetTypeInt()
}

// func (zAcc *ZbcAccountType) GetFormattedAccount() (string, error) {
// 	return crypto.NewEd25519Signature().GetAddressFromPublicKey(zAcc.GetAccountPrefix(), zAcc.GetAccountPublicKey())
// }
