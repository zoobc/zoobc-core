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
	"github.com/zoobc/zoobc-core/common/util"
)

type MultiSignatureAccountType struct {
	privateKey, publicKey, fullAddress []byte
	publicKeyString                    string
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
	return nil, blocker.NewBlocker(blocker.AppErr, "AccountNotGenerated")
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

func (acc *MultiSignatureAccountType) Sign([]byte, string, ...interface{}) ([]byte, error) {
	return nil, blocker.NewBlocker(blocker.AppErr, "NotAllowedSigning")
}

// VerifySignature especially for MultisignatureAccountType, the accountAddress passed variable is:
// [[count][accountAddresses]]
func (acc *MultiSignatureAccountType) VerifySignature(payload, signature, accountAddresses []byte) error {

	var (
		buff = bytes.NewBuffer(accountAddresses)
		b    []byte
		err  error
	)
	b = buff.Next(4)

	for i := 0; i <= int(util.ConvertBytesToUint32(b)); i++ {
		var accType AccountTypeInterface
		accType, err = ParseBytesToAccountType(buff)
		if err != nil {
			return err
		}
		err = accType.VerifySignature(payload, signature, accType.GetAccountPublicKey())
		if err != nil {
			return err
		}
	}
	return nil
}

func (acc *MultiSignatureAccountType) GenerateAccountFromSeed(string, ...interface{}) error {
	return blocker.NewBlocker(blocker.AppErr, "NotAllowedGenerateFromSeed")
}
