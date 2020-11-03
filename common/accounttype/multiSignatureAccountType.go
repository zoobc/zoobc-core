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
	privateKey, publicKey, fullAddress []byte
	publicKeyString, encodedAddress    string
}

func (acc *MultiSignatureAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.publicKey = make([]byte, 0)
	}
	acc.publicKey = accountPublicKey
}

func (acc *MultiSignatureAccountType) GetAccountAddress() ([]byte, error) {
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

func (acc *MultiSignatureAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_MultiSignatureAccountType)
}

func (acc *MultiSignatureAccountType) GetAccountPublicKey() []byte {
	return acc.publicKey
}

func (acc *MultiSignatureAccountType) GetAccountPrefix() string {
	return model.PrefixConstant_ZMS.String()
}

func (acc *MultiSignatureAccountType) GetName() string {
	return "ZooMS"
}

func (acc *MultiSignatureAccountType) GetAccountPublicKeyLength() uint32 {
	return 32
}

func (acc *MultiSignatureAccountType) GetEncodedAddress() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	return address.EncodeZbcID(acc.GetAccountPrefix(), acc.GetAccountPublicKey())
}

func (acc *MultiSignatureAccountType) GetAccountPublicKeyString() (string, error) {
	var (
		err error
	)
	if acc.publicKeyString != "" {
		return acc.publicKeyString, nil
	}
	if len(acc.publicKey) == 0 {
		return "", blocker.NewBlocker(blocker.AppErr, "EmptyAccountPublicKey")
	}
	acc.publicKeyString, err = signaturetype.NewEd25519Signature().GetAddressFromPublicKey(model.PrefixConstant_ZMS.String(), acc.publicKey)
	return acc.publicKeyString, err
}

func (acc *MultiSignatureAccountType) GetAccountPrivateKey() ([]byte, error) {
	if len(acc.privateKey) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "AccountNotGenerated")
	}
	return acc.privateKey, nil
}

func (acc *MultiSignatureAccountType) IsEqual(accType AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), accType.GetAccountPublicKey()) && acc.GetTypeInt() == accType.GetTypeInt()

}

func (acc *MultiSignatureAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_MultisigSignature
}

func (acc *MultiSignatureAccountType) GetSignatureLength() uint32 {
	return constant.ZBCSignatureLength
}

func (acc *MultiSignatureAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	var (
		ed25519Signature  = signaturetype.NewEd25519Signature()
		err               error
		buffer            = bytes.NewBuffer([]byte{})
		accountPrivateKey []byte
	)
	err = acc.GenerateAccountFromSeed(seed, optionalParams...)
	if err != nil {
		return nil, err
	}
	accountPrivateKey, err = acc.GetAccountPrivateKey()
	if err != nil {
		return nil, err
	}
	signature := ed25519Signature.Sign(accountPrivateKey, payload)
	buffer.Write(signature)
	return buffer.Bytes(), nil
}

func (acc *MultiSignatureAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
	accType, err := NewAccountTypeFromAccount(accountAddress)
	if err != nil {
		return err
	}
	ed25519Signature := signaturetype.NewEd25519Signature()
	accPubKey := accType.GetAccountPublicKey()
	if !ed25519Signature.Verify(accPubKey, payload, signature) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidSignature",
		)
	}
	return nil
}

func (acc *MultiSignatureAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	return blocker.NewBlocker(blocker.AppErr, "NotAllowedGenerateFromSeed")
}
