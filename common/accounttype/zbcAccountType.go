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

// ZbcAccountType the default account type
type ZbcAccountType struct {
	seed                               string
	privateKey, publicKey, fullAddress []byte
	publicKeyString, encodedAddress    string
}

func (acc *ZbcAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.publicKey = make([]byte, 0)
	}
	acc.publicKey = accountPublicKey
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
	return acc.publicKey
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

func (acc *ZbcAccountType) IsEqual(acc2 AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *ZbcAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_DefaultSignature
}

func (acc *ZbcAccountType) GetSignatureLength() uint32 {
	return constant.ZBCSignatureLength
}

func (acc *ZbcAccountType) GetEncodedAddress() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	return address.EncodeZbcID(acc.GetAccountPrefix(), acc.GetAccountPublicKey())
}

func (acc *ZbcAccountType) SetEncodedAccountAddress(encodedAccount string) {
	// acc.encodedAddress = encodedAccount
}

func (acc *ZbcAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	var (
		ed25519Signature = signaturetype.NewEd25519Signature()
		useSlip10, ok    bool
		err              error
	)
	if len(optionalParams) != 0 {
		useSlip10, ok = optionalParams[0].(bool)
		if !ok {
			return blocker.NewBlocker(blocker.AppErr, "failedAssertType")
		}
	}
	if useSlip10 {
		acc.privateKey, err = ed25519Signature.GetPrivateKeyFromSeedUseSlip10(seed)
		if err != nil {
			return err
		}
		acc.publicKey, err = ed25519Signature.GetPublicKeyFromPrivateKeyUseSlip10(acc.privateKey)
		if err != nil {
			return err
		}
	} else {
		acc.privateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
		acc.publicKey, err = ed25519Signature.GetPublicKeyFromPrivateKey(acc.privateKey)
		if err != nil {
			return err
		}
	}
	acc.publicKeyString, err = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcNodeAccount, acc.publicKey)
	if err != nil {
		return err
	}
	acc.encodedAddress, err = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcDefaultAccount, acc.publicKey)
	if err != nil {
		return err
	}
	return nil
}

func (acc *ZbcAccountType) GetAccountPublicKeyString() (string, error) {
	var (
		err error
	)
	if acc.publicKeyString != "" {
		return acc.publicKeyString, nil
	}
	if len(acc.publicKey) == 0 {
		return "", blocker.NewBlocker(blocker.AppErr, "EmptyAccountPublicKey")
	}
	acc.publicKeyString, err = signaturetype.NewEd25519Signature().GetAddressFromPublicKey(constant.PrefixZoobcNodeAccount, acc.publicKey)
	return acc.publicKeyString, err
}

func (acc *ZbcAccountType) GetAccountPrivateKey() ([]byte, error) {
	if len(acc.privateKey) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "AccountNotGenerated")
	}
	return acc.privateKey, nil
}

func (acc *ZbcAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	var (
		ed25519Signature  = signaturetype.NewEd25519Signature()
		accountPrivateKey []byte
		useSlip10, ok     bool
		err               error
		buffer            = bytes.NewBuffer([]byte{})
	)
	// optionalParams index 0 used for flag boolean slip10
	if len(optionalParams) != 0 {
		useSlip10, ok = optionalParams[0].(bool)
		if !ok {
			return nil, blocker.NewBlocker(blocker.AppErr, "failedAssertType")
		}
	}
	if useSlip10 {
		accountPrivateKey, err = ed25519Signature.GetPrivateKeyFromSeedUseSlip10(seed)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.AppErr, err.Error())
		}
		publicKey, err := ed25519Signature.GetPublicKeyFromPrivateKeyUseSlip10(accountPrivateKey)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.AppErr, err.Error())
		}
		accountPrivateKey = append(accountPrivateKey, publicKey...)
	} else {
		accountPrivateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
	}

	signature := ed25519Signature.Sign(accountPrivateKey, payload)
	buffer.Write(signature)
	return buffer.Bytes(), nil
}

func (acc *ZbcAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
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
