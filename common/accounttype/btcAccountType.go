package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/zoobc/zoobc-core/common/model"
)

// BTCAccountType a dummy account type
type BTCAccountType struct {
	accountPublicKey []byte
	encodedAddress   string
}

func (btcAcc *BTCAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	btcAcc.accountPublicKey = accountPublicKey
}

func (btcAcc *BTCAccountType) GetAccountAddress() ([]byte, error) {
	if btcAcc.GetAccountPublicKey() == nil {
		return nil, errors.New("AccountAddressPublicKeyEmpty")
	}
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(btcAcc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(btcAcc.GetAccountPublicKey())
	return buff.Bytes(), nil
}

func (btcAcc *BTCAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_BTCAccountType)
}

func (btcAcc *BTCAccountType) GetAccountPublicKey() []byte {
	return btcAcc.accountPublicKey
}

func (btcAcc *BTCAccountType) GetAccountPrefix() string {
	return "DUM"
}

func (btcAcc *BTCAccountType) GetName() string {
	return "Dummy"
}

func (btcAcc *BTCAccountType) GetAccountPublicKeyLength() uint32 {
	return btcec.PubKeyBytesLenCompressed
}

func (btcAcc *BTCAccountType) IsEqual(acc AccountType) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), btcAcc.GetAccountPublicKey()) && acc.GetTypeInt() == btcAcc.GetTypeInt()
}

// func (dAcc *BTCAccountType) GetSignatureTypeInterface() crypto.SignatureTypeInterface {
// 	return crypto.NewBitcoinSignature()
// }
//
//
// func (dAcc *BTCAccountType) GetAccountSignatureInterface() crypto.SignatureInterface {
// 	return crypto.NewSignature()
// }

func (btcAcc *BTCAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_BitcoinSignature
}

func (btcAcc *BTCAccountType) GetFormattedAccount() (string, error) {
	if btcAcc.encodedAddress != "" {
		return btcAcc.encodedAddress, nil
	}
	// TODO: for now, due to the high complexity of bringing in signing methods in this package,
	//  instead of calculating the encoded address, we return an error if this variable has not been calculated before
	return "", errors.New("EmptyAddress")
}

func (btcAcc *BTCAccountType) SetEncodedAccountAddress(encodedAccount string) {
	btcAcc.encodedAddress = encodedAccount
}
