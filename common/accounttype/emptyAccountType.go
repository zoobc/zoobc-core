package accounttype

import (
	"bytes"
	"encoding/binary"
	"github.com/zoobc/zoobc-core/common/model"
)

// EmptyAccountType the default account type
type EmptyAccountType struct {
	accountPublicKey []byte
	encodedAddress   string
}

func (acc *EmptyAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.accountPublicKey = make([]byte, 0)
	}
	// could be a zero-padded pub key
	acc.accountPublicKey = accountPublicKey
}

func (acc *EmptyAccountType) GetAccountAddress() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(acc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(acc.GetAccountPublicKey())
	return buff.Bytes(), nil
}

func (acc *EmptyAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_EmptyAccountType)
}

func (acc *EmptyAccountType) GetAccountPublicKey() []byte {
	if acc.accountPublicKey == nil {
		return make([]byte, 0)
	}
	return acc.accountPublicKey
}

func (acc *EmptyAccountType) GetAccountPrefix() string {
	return ""
}

func (acc *EmptyAccountType) GetName() string {
	return "Empty"
}

func (acc *EmptyAccountType) GetAccountPublicKeyLength() uint32 {
	return 0
}

func (acc *EmptyAccountType) IsEqual(acc2 AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *EmptyAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_DefaultSignature
}

func (acc *EmptyAccountType) GetSignatureLength() uint32 {
	return 0
}

func (acc *EmptyAccountType) GetFormattedAccount() (string, error) {
	return "", nil
}

func (acc *EmptyAccountType) SetEncodedAccountAddress(encodedAccount string) {
	acc.encodedAddress = ""
}
