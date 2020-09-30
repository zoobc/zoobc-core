package accounttype

import (
	"bytes"
	"encoding/binary"
)

// DummyAccountType a dummy account type
// TODO: this is only for the sake of having at least two account type.
//  as soon as we can add a real one this will be overridden by it
type DummyAccountType struct {
	accountPublicKey []byte
}

func (dAcc *DummyAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	dAcc.accountPublicKey = accountPublicKey
}

func (dAcc *DummyAccountType) GetAccountAddress() []byte {
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, dAcc.GetTypeInt())
	buff.Write(tmpBuf)
	buff.Write(dAcc.GetAccountPublicKey())
	return buff.Bytes()
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

func (dAcc *DummyAccountType) GetAccountPublicKeyLength() uint32 {
	return 32
}

func (dAcc *DummyAccountType) IsEqual(acc AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), dAcc.GetAccountPublicKey()) && acc.GetTypeInt() == dAcc.GetTypeInt()
}

// func (dAcc *DummyAccountType) GetFormattedAccount() (string, error) {
// 	return crypto.NewEd25519Signature().GetAddressFromPublicKey(dAcc.GetAccountPrefix(), dAcc.GetAccountPublicKey())
// }
