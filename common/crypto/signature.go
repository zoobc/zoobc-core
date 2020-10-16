package crypto

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zed25519/zed"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// SignatureInterface represent interface of signature
	SignatureInterface interface {
		Sign(payload []byte, accountType model.AccountType, seed string, optionalParams ...interface{}) ([]byte, error)
		SignByNode(payload []byte, nodeSeed string) []byte
		VerifySignature(payload, signature, accountAddress []byte) error
		VerifyNodeSignature(payload, signature []byte, nodePublicKey []byte) bool
		GenerateAccountFromSeed(accountType accounttype.AccountTypeInterface, seed string, optionalParams ...interface{}) (
			privateKey, publicKey []byte,
			publicKeyString, encodedAddress string,
			fullAccountAddress []byte,
			err error,
		)
		GenerateBlockSeed(payload []byte, nodeSeed string) []byte
	}

	// Signature object handle signing and verifying different signature
	Signature struct {
	}
)

// NewSignature create new instance of signature object
func NewSignature() *Signature {
	return &Signature{}
}

// Sign accept account ID and payload to be signed then return the signature byte based on the
// signature method associated with account.Type
func (*Signature) Sign(
	payload []byte,
	accountTypeEnum model.AccountType,
	seed string,
	optionalParams ...interface{},
) ([]byte, error) {
	accountType, err := accounttype.NewAccountType(int32(accountTypeEnum), nil)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	switch accountType.GetSignatureType() {
	case model.SignatureType_DefaultSignature:
		var (
			ed25519Signature  = NewEd25519Signature()
			accountPrivateKey []byte
			useSlip10, ok     bool
			err               error
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
	case model.SignatureType_BitcoinSignature:
		var (
			bitcoinSignature       = NewBitcoinSignature(DefaultBitcoinNetworkParams(), DefaultBitcoinCurve())
			accountPrivateKey, err = bitcoinSignature.GetPrivateKeyFromSeed(seed, DefaultBitcoinPrivateKeyLength())
		)
		if err != nil {
			return nil, err
		}
		accountPublicKey, err := bitcoinSignature.GetPublicKeyFromPrivateKey(
			accountPrivateKey,
			DefaultBitcoinPublicKeyFormat(),
		)
		if err != nil {
			return nil, err
		}
		// Add public key into signature bytes
		accountPublicKeyLength := util.ConvertUint16ToBytes(uint16(len(accountPublicKey)))
		buffer.Write(accountPublicKeyLength)
		buffer.Write(accountPublicKey)
		signature, err := bitcoinSignature.Sign(accountPrivateKey, payload)
		if err != nil {
			return nil, err
		}

		buffer.Write(signature)
		return buffer.Bytes(), nil
	default:
		return nil, blocker.NewBlocker(
			blocker.AppErr,
			"InvalidSignatureType",
		)
	}
}

// SignByNode special method for signing block only, there will be no multiple signature options
func (*Signature) SignByNode(payload []byte, nodeSeed string) []byte {
	var (
		buffer           = bytes.NewBuffer([]byte{})
		ed25519Signature = NewEd25519Signature()
		nodePrivateKey   = ed25519Signature.GetPrivateKeyFromSeed(nodeSeed)
		signature        = ed25519Signature.Sign(nodePrivateKey, payload)
	)
	buffer.Write(signature)
	return buffer.Bytes()
}

// VerifySignature accept payload (before without signature), signature and the account id
// then verify the signature + public key against the payload based on the
func (*Signature) VerifySignature(payload, signature, accountAddress []byte) error {
	var (
		accountTypeInt = int32(util.ConvertBytesToUint32(accountAddress[:4]))
	)
	accountType, err := accounttype.NewAccountType(accountTypeInt, accountAddress[4:])
	if err != nil {
		return err
	}
	// TODO: move this logic into AccountTypeInterface interface (remove switch cases for every account/signature type)
	switch accountType.GetSignatureType() {
	case model.SignatureType_DefaultSignature: // zoobc
		accType, err := accounttype.NewAccountTypeFromAccount(accountAddress)
		if err != nil {
			return err
		}
		ed25519Signature := NewEd25519Signature()
		if !ed25519Signature.Verify(accType.GetAccountPublicKey(), payload, signature) {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"InvalidSignature",
			)
		}
		return nil
	case model.SignatureType_BitcoinSignature: // bitcoin
		var (
			bitcoinSignature = NewBitcoinSignature(DefaultBitcoinNetworkParams(), DefaultBitcoinCurve())
			// first 2 bytes are the public key length
			pubKeyFirstBytesIndex    = 2
			pubKeyBytesLength        = util.ConvertBytesToUint16(signature[:pubKeyFirstBytesIndex])
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
		accType, err := accounttype.ParseBytesToAccountType(bytes.NewBuffer(accountAddress))
		if err != nil {
			return err
		}
		accPubKey := accType.GetAccountPublicKey()
		accountAddress, err := bitcoinSignature.GetAddressFromPublicKey(accPubKey)
		if err != nil {
			return err
		}
		if accountAddress != signaturePubKeyAddress {
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
	default:
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidSignatureType",
		)
	}
}

// VerifyNodeSignature Verify a signature of a block or message signed with a node private key
// Note: this function is a wrapper around the ed25519 algorithm
func (*Signature) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	var result = NewEd25519Signature().Verify(nodePublicKey, payload, signature)
	return result
}

// GenerateAccountFromSeed to generate account based on provided seed
func (*Signature) GenerateAccountFromSeed(accountType accounttype.AccountTypeInterface, seed string, optionalParams ...interface{}) (
	privateKey, publicKey []byte,
	publicKeyString, encodedAddress string,
	fullAccountAddress []byte,
	err error,
) {
	switch accountType.GetSignatureType() {
	case model.SignatureType_DefaultSignature:
		var (
			ed25519Signature = NewEd25519Signature()
			useSlip10, ok    bool
		)
		if len(optionalParams) != 0 {
			useSlip10, ok = optionalParams[0].(bool)
			if !ok {
				return nil, nil, "", "", nil, blocker.NewBlocker(blocker.AppErr, "failedAssertType")
			}
		}
		if useSlip10 {
			privateKey, err = ed25519Signature.GetPrivateKeyFromSeedUseSlip10(seed)
			if err != nil {
				return nil, nil, "", "", nil, err
			}
			publicKey, err = ed25519Signature.GetPublicKeyFromPrivateKeyUseSlip10(privateKey)
			if err != nil {
				return nil, nil, "", "", nil, err
			}
		} else {
			privateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
			publicKey, err = ed25519Signature.GetPublicKeyFromPrivateKey(privateKey)
			if err != nil {
				return nil, nil, "", "", nil, err
			}
		}
		publicKeyString, err = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcNodeAccount, publicKey)
		if err != nil {
			return nil, nil, "", "", nil, err
		}
		encodedAddress, err = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcDefaultAccount, publicKey)
		if err != nil {
			return nil, nil, "", "", nil, err
		}
	case model.SignatureType_BitcoinSignature:
		var (
			bitcoinSignature = NewBitcoinSignature(DefaultBitcoinNetworkParams(), DefaultBitcoinCurve())
			privateKeyLength = DefaultBitcoinPrivateKeyLength()
			publicKeyFormat  = DefaultBitcoinPublicKeyFormat()
			ok               bool
		)
		if len(optionalParams) >= 2 {
			privateKeyLength, ok = optionalParams[0].(model.PrivateKeyBytesLength)
			if !ok {
				return nil, nil, "", "", nil, blocker.NewBlocker(blocker.AppErr, "failedAssertPrivateKeyLengthType")
			}
			publicKeyFormat, ok = optionalParams[1].(model.BitcoinPublicKeyFormat)
			if !ok {
				return nil, nil, "", "", nil, blocker.NewBlocker(blocker.AppErr, "failedAssertPublicKeyFormatType")
			}
		}
		privKey, err := bitcoinSignature.GetPrivateKeyFromSeed(seed, privateKeyLength)
		if err != nil {
			return nil, nil, "", "", nil, err
		}
		privateKey = privKey.Serialize()
		publicKey, err = bitcoinSignature.GetPublicKeyFromSeed(
			seed,
			publicKeyFormat,
			privateKeyLength,
		)
		if err != nil {
			return nil, nil, "", "", nil, err
		}
		encodedAddress, err = bitcoinSignature.GetAddressFromPublicKey(publicKey)
		if err != nil {
			return nil, nil, "", "", nil, err
		}
		publicKeyString, err = bitcoinSignature.GetPublicKeyString(publicKey)
		if err != nil {
			return nil, nil, "", "", nil, err
		}
	default:
		return nil, nil, "", "", nil, blocker.NewBlocker(
			blocker.AppErr,
			"InvalidSignatureType",
		)
	}
	accountType.SetAccountPublicKey(publicKey)
	accountType.SetEncodedAccountAddress(encodedAddress)
	fullAccountAddress, err = accountType.GetAccountAddress()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	return privateKey, publicKey, publicKeyString, encodedAddress, fullAccountAddress, nil
}

// GenerateBlockSeed special method for generating block seed using zed
func (*Signature) GenerateBlockSeed(payload []byte, nodeSeed string) []byte {
	var (
		buffer       = bytes.NewBuffer([]byte{})
		seedBuffer   = []byte(nodeSeed)
		seedHash     = sha3.Sum256(seedBuffer)
		seedByte     = seedHash[:]
		zedSecret    = zed.SecretFromSeed(seedByte)
		zedSignature = zedSecret.Sign(payload)
	)
	buffer.Write(zedSignature[:])
	return buffer.Bytes()
}
