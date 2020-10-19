package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/signaturetype"
)

// BTCAccountType a dummy account type
type BTCAccountType struct {
	privateKey, publicKey, fullAddress []byte
	publicKeyString, encodedAddress    string
}

func (acc *BTCAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.publicKey = make([]byte, 0)
	}
	acc.publicKey = accountPublicKey
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
	return acc.publicKey
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

func (acc *BTCAccountType) IsEqual(acc2 AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *BTCAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_BitcoinSignature
}

func (acc *BTCAccountType) GetSignatureLength() uint32 {
	return constant.BTCECDSASignatureLength
}

func (acc *BTCAccountType) GetEncodedAddress() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	bitcoinSignature := signaturetype.NewBitcoinSignature(signaturetype.DefaultBitcoinNetworkParams(), signaturetype.DefaultBitcoinCurve())
	return bitcoinSignature.GetAddressFromPublicKey(acc.GetAccountPublicKey())
}

func (acc *BTCAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	var (
		bitcoinSignature = signaturetype.NewBitcoinSignature(signaturetype.DefaultBitcoinNetworkParams(), signaturetype.DefaultBitcoinCurve())
		privateKeyLength = signaturetype.DefaultBitcoinPrivateKeyLength()
		publicKeyFormat  = signaturetype.DefaultBitcoinPublicKeyFormat()
		ok               bool
	)
	if len(optionalParams) >= 2 {
		privateKeyLength, ok = optionalParams[0].(model.PrivateKeyBytesLength)
		if !ok {
			return blocker.NewBlocker(blocker.AppErr, "failedAssertPrivateKeyLengthType")
		}
		publicKeyFormat, ok = optionalParams[1].(model.BitcoinPublicKeyFormat)
		if !ok {
			return blocker.NewBlocker(blocker.AppErr, "failedAssertPublicKeyFormatType")
		}
	}
	privKey, err := bitcoinSignature.GetPrivateKeyFromSeed(seed, privateKeyLength)
	if err != nil {
		return err
	}
	acc.privateKey = privKey.Serialize()
	acc.publicKey, err = bitcoinSignature.GetPublicKeyFromSeed(
		seed,
		publicKeyFormat,
		privateKeyLength,
	)
	if err != nil {
		return err
	}
	acc.encodedAddress, err = bitcoinSignature.GetAddressFromPublicKey(acc.publicKey)
	if err != nil {
		return err
	}
	acc.publicKeyString, err = bitcoinSignature.GetPublicKeyString(acc.publicKey)
	if err != nil {
		return err
	}
	return nil
}

func (acc *BTCAccountType) GetAccountPublicKeyString() (string, error) {
	var (
		err error
	)
	if acc.publicKeyString != "" {
		return acc.publicKeyString, nil
	}
	if len(acc.publicKey) == 0 {
		return "", blocker.NewBlocker(blocker.AppErr, "EmptyAccountPublicKey")
	}
	acc.publicKeyString, err = signaturetype.NewBitcoinSignature(
		signaturetype.DefaultBitcoinNetworkParams(),
		signaturetype.DefaultBitcoinCurve()).GetPublicKeyString(acc.publicKey)
	return acc.publicKeyString, err
}

func (acc *BTCAccountType) GetAccountPrivateKey() ([]byte, error) {
	if len(acc.privateKey) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "AccountNotGenerated")
	}
	return acc.privateKey, nil
}

func (acc *BTCAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	var (
		bitcoinSignature       = signaturetype.NewBitcoinSignature(signaturetype.DefaultBitcoinNetworkParams(), signaturetype.DefaultBitcoinCurve())
		accountPrivateKey, err = bitcoinSignature.GetPrivateKeyFromSeed(seed, signaturetype.DefaultBitcoinPrivateKeyLength())
		buffer                 = bytes.NewBuffer([]byte{})
	)
	if err != nil {
		return nil, err
	}
	accountPublicKey, err := bitcoinSignature.GetPublicKeyFromPrivateKey(
		accountPrivateKey,
		signaturetype.DefaultBitcoinPublicKeyFormat(),
	)
	if err != nil {
		return nil, err
	}
	// Add public key into signature bytes
	accountPublicKeyLength := convertUint16ToBytes(uint16(len(accountPublicKey)))
	buffer.Write(accountPublicKeyLength)
	buffer.Write(accountPublicKey)
	signature, err := bitcoinSignature.Sign(accountPrivateKey, payload)
	if err != nil {
		return nil, err
	}

	buffer.Write(signature)
	return buffer.Bytes(), nil
}

func (acc *BTCAccountType) VerifySignature(payload, signature, fullAccountAddress []byte) error {
	var (
		bitcoinSignature = signaturetype.NewBitcoinSignature(signaturetype.DefaultBitcoinNetworkParams(), signaturetype.DefaultBitcoinCurve())
		// first 2 bytes are the public key length
		pubKeyFirstBytesIndex    = 2
		pubKeyBytesLength        = convertBytesToUint16(signature[:pubKeyFirstBytesIndex])
		signatureFirstBytesIndex = pubKeyFirstBytesIndex + int(pubKeyBytesLength)
		signaturePubKeyBytes     = signature[pubKeyFirstBytesIndex:signatureFirstBytesIndex]
		signaturePubKey, err     = bitcoinSignature.GetPublicKeyFromBytes(signaturePubKeyBytes)
	)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			err.Error(),
		)
	}
	signaturePubKeyAddress, err := bitcoinSignature.GetAddressFromPublicKey(signaturePubKeyBytes)
	if err != nil {
		return err
	}
	accType, err := ParseBytesToAccountType(bytes.NewBuffer(fullAccountAddress))
	if err != nil {
		return err
	}
	accPubKey := accType.GetAccountPublicKey()
	encodedAddress, err := bitcoinSignature.GetAddressFromPublicKey(accPubKey)
	if err != nil {
		return err
	}
	if encodedAddress != signaturePubKeyAddress {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"invalidAccountAddressOrSignaturePublicKey",
		)
	}
	sig, err := bitcoinSignature.GetSignatureFromBytes(signature[signatureFirstBytesIndex:])
	if err != nil {
		return err

	}
	if !bitcoinSignature.Verify(payload, sig, signaturePubKey) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidSignature",
		)
	}
	return nil
}

// TODO: refactor this. we can't use the same function in utils package because of a circular dependency
func convertUint16ToBytes(number uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, number)
	return buf
}

// TODO: refactor this. we can't use the same function in utils package because of a circular dependency
func convertBytesToUint16(dataBytes []byte) uint16 {
	return binary.LittleEndian.Uint16(dataBytes)
}
