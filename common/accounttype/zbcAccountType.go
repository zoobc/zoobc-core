package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// ZbcAccountType the default account type
type ZbcAccountType struct {
	accountPublicKey []byte
	encodedAddress   string
}

func (zAcc *ZbcAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	zAcc.accountPublicKey = accountPublicKey
}

func (zAcc *ZbcAccountType) GetAccountAddress() ([]byte, error) {
	if zAcc.GetAccountPublicKey() == nil {
		return nil, errors.New("AccountAddressPublicKeyEmpty")
	}
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(zAcc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(zAcc.GetAccountPublicKey())
	return buff.Bytes(), nil
}

func (zAcc *ZbcAccountType) GetTypeInt() int32 {
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

func (zAcc *ZbcAccountType) GetAccountPublicKeyLength() uint32 {
	return 32
}

func (zAcc *ZbcAccountType) IsEqual(acc AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), zAcc.GetAccountPublicKey()) && acc.GetTypeInt() == zAcc.GetTypeInt()
}

// func (zAcc *ZbcAccountType) GetSignatureTypeInterface() crypto.SignatureTypeInterface {
// 	return crypto.NewEd25519Signature()
// }
//
// func (zAcc *ZbcAccountType) GetAccountSignatureInterface() crypto.SignatureInterface {
// 	return crypto.NewSignature()
// }

func (zAcc *ZbcAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_DefaultSignature
}

func (zAcc *ZbcAccountType) GetFormattedAccount() (string, error) {
	if zAcc.encodedAddress != "" {
		return zAcc.encodedAddress, nil
	}

	// TODO: for now, due to the high complexity of bringing in signing methods in this package,
	//  instead of calculating the encoded address, we return an error if this variable has not been calculated before
	return "", errors.New("EmptyAddress")
}

func (zAcc *ZbcAccountType) SetEncodedAccountAddress(encodedAccount string) {
	zAcc.encodedAddress = encodedAccount
}
