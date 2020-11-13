package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/signaturetype"
)

type MultiSignatureAccountType struct {
	publicKey, fullAddress []byte
	publicKeyString        string
}

func (m *MultiSignatureAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if len(accountPublicKey) == 0 {
		m.publicKey = make([]byte, 0)
	}
	m.publicKey = accountPublicKey
}

func (m *MultiSignatureAccountType) GetAccountAddress() ([]byte, error) {
	if m.fullAddress != nil {
		return m.fullAddress, nil
	}
	if m.GetAccountPublicKey() == nil {
		return nil, errors.New("AccountAddressPublicKeyEmpty")
	}
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(m.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(m.GetAccountPublicKey())
	m.fullAddress = buff.Bytes()
	return m.fullAddress, nil
}

func (m *MultiSignatureAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_MultiSignatureAccountType)
}

func (m *MultiSignatureAccountType) GetAccountPublicKey() []byte {
	return m.publicKey
}

func (m *MultiSignatureAccountType) GetAccountPrefix() string {
	return constant.PrefixMultiSignatureAccount
}

func (m *MultiSignatureAccountType) GetName() string {
	return "ZooMS"
}

func (m *MultiSignatureAccountType) GetAccountPublicKeyLength() uint32 {
	return 32
}

func (m *MultiSignatureAccountType) GetEncodedAddress() (string, error) {
	if m.GetAccountPublicKey() == nil || bytes.Equal(m.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	return address.EncodeZbcID(m.GetAccountPrefix(), m.GetAccountPublicKey())
}

func (m *MultiSignatureAccountType) GetAccountPublicKeyString() (string, error) {
	var (
		err error
	)
	if m.publicKeyString != "" {
		return m.publicKeyString, nil
	}
	if len(m.publicKey) == 0 {
		return "", blocker.NewBlocker(blocker.AppErr, "EmptyAccountPublicKey")
	}
	m.publicKeyString, err = signaturetype.NewEd25519Signature().GetAddressFromPublicKey(constant.PrefixZoobcNodeAccount, m.publicKey)
	return m.publicKeyString, err
}

func (m *MultiSignatureAccountType) GetAccountPrivateKey() ([]byte, error) {
	return []byte{}, blocker.NewBlocker(blocker.AppErr, "PrivateDoesNotGeneratePrivateKey")
}

func (m *MultiSignatureAccountType) IsEqual(acc AccountTypeInterface) bool {
	return bytes.Equal(m.GetAccountPublicKey(), acc.GetAccountPublicKey()) && acc.GetTypeInt() == acc.GetTypeInt()
}

func (m *MultiSignatureAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_MultisigSignature
}

func (m *MultiSignatureAccountType) GetSignatureLength() uint32 {
	return 0
}

func (m *MultiSignatureAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	return []byte{}, blocker.NewBlocker(blocker.AppErr, "NotAllowedSigning")
}

func (m *MultiSignatureAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
	return blocker.NewBlocker(blocker.AppErr, "NotAllowedVerifying")
}

func (m *MultiSignatureAccountType) GenerateAccountFromSeed(string, ...interface{}) error {
	return blocker.NewBlocker(blocker.AppErr, "NotAlloweedGenerateAccountFromSeed")
}
