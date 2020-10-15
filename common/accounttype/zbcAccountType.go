package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// ZbcAccountType the default account type
type ZbcAccountType struct {
	accountPublicKey []byte
	encodedAddress   string
}

func (acc *ZbcAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	acc.accountPublicKey = accountPublicKey
}

func (acc *ZbcAccountType) GetAccountAddress() ([]byte, error) {
	if acc.GetAccountPublicKey() == nil {
		return nil, errors.New("AccountAddressPublicKeyEmpty")
	}
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(acc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(acc.GetAccountPublicKey())
	return buff.Bytes(), nil
}

func (acc *ZbcAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_ZbcAccountType)
}

func (acc *ZbcAccountType) GetAccountPublicKey() []byte {
	return acc.accountPublicKey
}

func (acc *ZbcAccountType) GetAccountPrefix() string {
	return constant.PrefixZoobcDefaultAccount
}

func (acc *ZbcAccountType) GetName() string {
	return "ZooBC"
}

func (acc *ZbcAccountType) GetAccountPublicKeyLength() uint32 {
	return 32
}

func (acc *ZbcAccountType) IsEqual(acc2 AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *ZbcAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_DefaultSignature
}

func (acc *ZbcAccountType) GetSignatureLength() uint32 {
	return constant.ZBCSignatureLength
}

func (acc *ZbcAccountType) GetFormattedAccount() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	return address.EncodeZbcID(acc.GetAccountPrefix(), acc.GetAccountPublicKey())
}

func (acc *ZbcAccountType) SetEncodedAccountAddress(encodedAccount string) {
	acc.encodedAddress = encodedAccount
}
