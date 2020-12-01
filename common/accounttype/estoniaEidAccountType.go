package accounttype

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// EstoniaEidAccountType the default account type
type EstoniaEidAccountType struct {
	publicKey, fullAddress []byte
}

func (acc *EstoniaEidAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.publicKey = make([]byte, 0)
	}
	acc.publicKey = accountPublicKey
}

func (acc *EstoniaEidAccountType) GetAccountAddress() ([]byte, error) {
	if acc.fullAddress != nil {
		return acc.fullAddress, nil
	}
	if acc.GetAccountPublicKey() == nil {
		return nil, errors.New("AccountAddressPublicKeyEmpty")
	}
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(acc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(acc.GetAccountPublicKey())
	acc.fullAddress = buff.Bytes()
	return acc.fullAddress, nil
}

func (acc *EstoniaEidAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_EstoniaEidAccountType)
}

func (acc *EstoniaEidAccountType) GetAccountPublicKey() []byte {
	return acc.publicKey
}

func (acc *EstoniaEidAccountType) GetAccountPrefix() string {
	return ""
}

func (acc *EstoniaEidAccountType) GetName() string {
	return "EstoniaEid"
}

func (acc *EstoniaEidAccountType) GetAccountPublicKeyLength() uint32 {
	return 97
}

func (acc *EstoniaEidAccountType) IsEqual(acc2 AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *EstoniaEidAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_EstoniaEidSignature
}

func (acc *EstoniaEidAccountType) GetSignatureLength() uint32 {
	return constant.EstoniaEidSignatureLength
}

func (acc *EstoniaEidAccountType) GetEncodedAddress() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	return hex.EncodeToString(acc.GetAccountPublicKey()), nil
}

func (acc *EstoniaEidAccountType) DecodePublicKeyFromAddress(address string) ([]byte, error) {
	return hex.DecodeString(address)
}

func (acc *EstoniaEidAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	return errors.New("NoImplementation")
}

func (acc *EstoniaEidAccountType) GetAccountPublicKeyString() (string, error) {
	return acc.GetEncodedAddress()
}

func (acc *EstoniaEidAccountType) GetAccountPrivateKey() ([]byte, error) {
	return nil, nil
}

func (acc *EstoniaEidAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	return []byte{}, nil
}

func (acc *EstoniaEidAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
	publicKey := acc.loadPublicKeyFromDer(acc.GetAccountPublicKey())
	r, s, _ := acc.decodeSignatureNIST384RS(signature)
	if !ecdsa.Verify(&publicKey, payload, r, s) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidSignature",
		)
	}
	return nil
}

// source: https://github.com/warner/python-ecdsa/blob/master/src/ecdsa/util.py (sigdecode_string)
// return: r, s, error
func (acc *EstoniaEidAccountType) decodeSignatureNIST384RS(signature []byte) (r, s *big.Int, err error) {
	// curveOrder "39402006196394479212279040100143613805079739270465446667946905279627659399113263569398956308152294913554433653942643"
	curveOrderLen := 48
	if len(signature) != curveOrderLen*2 {
		return nil, nil, fmt.Errorf("error signature length: %d", len(signature))
	}
	rBytes := signature[:48]
	sBytes := signature[48:]
	r = new(big.Int).SetBytes(rBytes)
	s = new(big.Int).SetBytes(sBytes)
	return r, s, nil
}

func (acc *EstoniaEidAccountType) loadPublicKeyFromDer(publicKeyBytes []byte) (publicKey ecdsa.PublicKey) {
	curve := elliptic.P384()
	publicKey.Curve = curve
	publicKey.X, publicKey.Y = elliptic.Unmarshal(curve, publicKeyBytes)
	return publicKey
}
