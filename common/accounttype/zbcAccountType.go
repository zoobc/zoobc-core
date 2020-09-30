package accounttype

import (
	"bytes"
	"encoding/binary"
)

// ZbcAccountType the default account type
type ZbcAccountType struct {
	accountPublicKey []byte
}

func (zAcc *ZbcAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	zAcc.accountPublicKey = accountPublicKey
}

func (zAcc *ZbcAccountType) GetAccountAddress() []byte {
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, zAcc.GetTypeInt())
	buff.Write(tmpBuf)
	buff.Write(zAcc.GetAccountPublicKey())
	return buff.Bytes()
}

func (zAcc *ZbcAccountType) GetTypeInt() uint32 {
	return 0
}

func (zAcc *ZbcAccountType) GetAccountPublicKey() []byte {
	return zAcc.accountPublicKey
}

func (zAcc *ZbcAccountType) GetAccountPrefix() string {
	return "ZBC"
}

func (zAcc *ZbcAccountType) GetName() string {
	return "ZooBC"
}

func (zAcc *ZbcAccountType) GetAccountPublicKeyLength() uint32 {
	return 32
}

func (zAcc *ZbcAccountType) IsEqual(acc AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), zAcc.GetAccountPublicKey()) && acc.GetTypeInt() == zAcc.GetTypeInt()
}

// func (zAcc *ZbcAccountType) GetFormattedAccount() (string, error) {
// 	return crypto.NewEd25519Signature().GetAddressFromPublicKey(zAcc.GetAccountPrefix(), zAcc.GetAccountPublicKey())
// }
