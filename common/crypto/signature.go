package crypto

import (
	"bytes"

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
		Sign(payload []byte, signatureType model.SignatureType, seed string, optionalParams ...interface{}) ([]byte, error)
		SignByNode(payload []byte, nodeSeed string) []byte
		VerifySignature(payload, signature []byte, accountAddress string) error
		VerifyNodeSignature(payload, signature []byte, nodePublicKey []byte) bool
		GenerateAccountFromSeed(signatureType model.SignatureType, seed string, optionalParams ...interface{}) (
			privateKey, publicKey []byte,
			publicKeyString, address string,
			err error,
		)
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
	signatureType model.SignatureType,
	seed string,
	optionalParams ...interface{},
) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(signatureType)))
	switch signatureType {
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

// VerifySignature accept payload (before without signature), signature and the account id
// then verify the signature + public key against the payload based on the
func (*Signature) VerifySignature(payload, signature []byte, accountAddress string) error {
	var (
		signatureType = util.ConvertBytesToUint32(signature[:4])
	)
	switch model.SignatureType(signatureType) {
	case model.SignatureType_DefaultSignature: // zoobc
		var (
			ed25519Signature      = NewEd25519Signature()
			accountPublicKey, err = ed25519Signature.GetPublicKeyFromAddress(accountAddress)
		)
		if err != nil {
			return err
		}
		if !ed25519Signature.Verify(accountPublicKey, payload, signature[4:]) {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"InvalidSignature",
			)
		}
		return nil
	case model.SignatureType_BitcoinSignature: // bitcoin
		var (
			bitcoinSignature = NewBitcoinSignature(DefaultBitcoinNetworkParams(), DefaultBitcoinCurve())
			// 2 bytes after signature type bytes is length of public key
			pubKeyFirstBytesIndex    = 6
			pubKeyBytesLength        = util.ConvertBytesToUint16(signature[4:pubKeyFirstBytesIndex])
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
		// check sender account address to address from public key in signature
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
func (*Signature) GenerateAccountFromSeed(signatureType model.SignatureType, seed string, optionalParams ...interface{}) (
	privateKey, publicKey []byte,
	publicKeyString, address string,
	err error,
) {
	switch signatureType {
	case model.SignatureType_DefaultSignature:
		var (
			ed25519Signature = NewEd25519Signature()
			useSlip10, ok    bool
		)
		if len(optionalParams) != 0 {
			useSlip10, ok = optionalParams[0].(bool)
			if !ok {
				return nil, nil, "", "", blocker.NewBlocker(blocker.AppErr, "failedAssertType")
			}
		}
		if useSlip10 {
			privateKey, err = ed25519Signature.GetPrivateKeyFromSeedUseSlip10(seed)
			if err != nil {
				return nil, nil, "", "", err
			}
			publicKey, err = ed25519Signature.GetPublicKeyFromPrivateKeyUseSlip10(privateKey)
			if err != nil {
				return nil, nil, "", "", err
			}
		} else {
			privateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
			publicKey, err = ed25519Signature.GetPublicKeyFromPrivateKey(privateKey)
			if err != nil {
				return nil, nil, "", "", err
			}
		}
		publicKeyString, err = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcNodeAccount, publicKey)
		if err != nil {
			return nil, nil, "", "", err
		}
		address, err = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcDefaultAccount, publicKey)
		if err != nil {
			return nil, nil, "", "", err
		}
		return privateKey, publicKey, publicKeyString, address, nil
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
				return nil, nil, "", "", blocker.NewBlocker(blocker.AppErr, "failedAssertPrivateKeyLengthType")
			}
			publicKeyFormat, ok = optionalParams[1].(model.BitcoinPublicKeyFormat)
			if !ok {
				return nil, nil, "", "", blocker.NewBlocker(blocker.AppErr, "failedAssertPublicKeyFormatType")
			}
		}
		privKey, err := bitcoinSignature.GetPrivateKeyFromSeed(seed, privateKeyLength)
		if err != nil {
			return nil, nil, "", "", err
		}
		privateKey = privKey.Serialize()
		publicKey, err = bitcoinSignature.GetPublicKeyFromSeed(
			seed,
			publicKeyFormat,
			privateKeyLength,
		)
		if err != nil {
			return nil, nil, "", "", err
		}
		address, err = bitcoinSignature.GetAddressFromPublicKey(publicKey)
		if err != nil {
			return nil, nil, "", "", err
		}
		publicKeyString, err = bitcoinSignature.GetPublicKeyString(publicKey)
		if err != nil {
			return nil, nil, "", "", err
		}
		return privateKey, publicKey, publicKeyString, address, nil
	default:
		return nil, nil, "", "", blocker.NewBlocker(
			blocker.AppErr,
			"InvalidSignatureType",
		)
	}
}
