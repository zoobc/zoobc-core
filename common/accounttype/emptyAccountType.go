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

func (zAcc *EmptyAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		zAcc.accountPublicKey = make([]byte, 0)
	}
	// could be a zero-padded pub key
	zAcc.accountPublicKey = accountPublicKey
}

func (zAcc *EmptyAccountType) GetAccountAddress() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(zAcc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(zAcc.GetAccountPublicKey())
	return buff.Bytes(), nil
}

func (zAcc *EmptyAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_EmptyAccountType)
}

func (zAcc *EmptyAccountType) GetAccountPublicKey() []byte {
	if zAcc.accountPublicKey == nil {
		return make([]byte, 0)
	}
	return zAcc.accountPublicKey
}

func (zAcc *EmptyAccountType) GetAccountPrefix() string {
	return ""
}

func (zAcc *EmptyAccountType) GetName() string {
	return "Empty"
}

func (zAcc *EmptyAccountType) GetAccountPublicKeyLength() uint32 {
	return 0
}

func (zAcc *EmptyAccountType) IsEqual(acc AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), zAcc.GetAccountPublicKey()) && acc.GetTypeInt() == zAcc.GetTypeInt()
}

func (zAcc *EmptyAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_DefaultSignature
}

func (zAcc *EmptyAccountType) GetFormattedAccount() (string, error) {
	return "", nil
}

func (zAcc *EmptyAccountType) SetEncodedAccountAddress(encodedAccount string) {
	zAcc.encodedAddress = ""
}
