package crypto

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// SignatureInterface represent interface of signature
	SignatureInterface interface {
		Sign(payload []byte, signatureType model.SignatureType, seed string) []byte
		SignByNode(payload []byte, nodeSeed string) []byte
		VerifySignature(payload, signature []byte, accountAddress string) bool
		VerifyNodeSignature(payload, signature []byte, nodePublicKey []byte) bool
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
func (sig *Signature) Sign(payload []byte, signatureType model.SignatureType, seed string) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(signatureType)))
	switch signatureType {
	case model.SignatureType_DefaultSignature:
		var (
			ed25519Signature  = NewEd25519Signature()
			accountPrivateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
			signature         = ed25519Signature.Sign(accountPrivateKey, payload)
		)
		buffer.Write(signature)
		return buffer.Bytes()
	case model.SignatureType_BitcoinSignature:
		var (
			bitcoinSignature  = NewBitcoinSignature(DefaultBitcoinNetworkParams(), DefaultBitcoinCurve())
			accountPrivateKey = bitcoinSignature.GetPrivateKeyFromSeed(seed)
			signature, err    = bitcoinSignature.Sign(accountPrivateKey, payload)
		)
		if err != nil {
			// TODO: need catch err into log
			return nil
		}
		buffer.Write(signature)
		return buffer.Bytes()
	default:
		var (
			ed25519Signature  = NewEd25519Signature()
			accountPrivateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
			signature         = ed25519Signature.Sign(accountPrivateKey, payload)
		)
		buffer.Write(signature)
		return buffer.Bytes()
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
func (*Signature) VerifySignature(payload, signature []byte, accountAddress string) bool {
	var (
		signatureType      = util.ConvertBytesToUint32(signature[:4])
		signatureTypeInt32 = util.ConvertUint32ToInt32(signatureType)
	)
	switch model.SignatureType(signatureTypeInt32) {
	case model.SignatureType_DefaultSignature: // zoobc
		var (
			ed25519Signature      = NewEd25519Signature()
			accountPublicKey, err = ed25519Signature.GetPublicKeyFromAddress(accountAddress)
		)
		// fmt.Println(ed25519Signature.Sign())
		if err != nil {
			// TODO: need catch err into log
			return false
		}
		return ed25519Signature.Verify(accountPublicKey, payload, signature[4:])
	case model.SignatureType_BitcoinSignature: // bitcoin
		var (
			bitcoinSignature = NewBitcoinSignature(DefaultBitcoinNetworkParams(), DefaultBitcoinCurve())
			publicKey, err   = bitcoinSignature.GetPublicKeyFromAddress(accountAddress)
		)
		if err != nil {
			// TODO: need catch err into log
			return false
		}
		sig, err := bitcoinSignature.GetSignatureFromBytes(signature[4:])
		if err != nil {
			// TODO: need catch err into log
			return false
		}
		return bitcoinSignature.Verify(payload, sig, publicKey)
	default:
		var (
			ed25519Signature      = NewEd25519Signature()
			accountPublicKey, err = ed25519Signature.GetPublicKeyFromAddress(accountAddress)
		)
		if err != nil {
			// TODO: need catch err into log
			return false
		}
		return ed25519Signature.Verify(accountPublicKey, payload, signature[4:])
	}
}

// VerifyNodeSignature Verify a signature of a block or message signed with a node private key
// Note: this function is a wrapper around the ed25519 algorithm
func (*Signature) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	var result = NewEd25519Signature().Verify(nodePublicKey, payload, signature)
	return result
}
