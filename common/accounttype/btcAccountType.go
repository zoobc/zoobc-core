package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// BTCAccountType a dummy account type
type BTCAccountType struct {
	accountPublicKey []byte
	encodedAddress   string
}

func (acc *BTCAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	acc.accountPublicKey = accountPublicKey
}

func (acc *BTCAccountType) GetAccountAddress() ([]byte, error) {
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

func (acc *BTCAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_BTCAccountType)
}

func (acc *BTCAccountType) GetAccountPublicKey() []byte {
	return acc.accountPublicKey
}

func (acc *BTCAccountType) GetAccountPrefix() string {
	return "BTC"
}

func (acc *BTCAccountType) GetName() string {
	return "BTCAccount"
}

func (acc *BTCAccountType) GetAccountPublicKeyLength() uint32 {
	return btcec.PubKeyBytesLenCompressed
}

func (acc *BTCAccountType) IsEqual(acc2 AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *BTCAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_BitcoinSignature
}

func (acc *BTCAccountType) GetSignatureLength() uint32 {
	return constant.BTCECDSASignatureLength
}

func (acc *BTCAccountType) GetFormattedAccount() (string, error) {
	if acc.encodedAddress != "" {
		return acc.encodedAddress, nil
	}
	// TODO: for now, due to the high complexity of bringing in signing methods in this package,
	//  instead of calculating the encoded address, we return an error if this variable has not been calculated before
	return "", errors.New("EmptyAddress")
}

func (acc *BTCAccountType) SetEncodedAccountAddress(encodedAccount string) {
	acc.encodedAddress = encodedAccount
}
