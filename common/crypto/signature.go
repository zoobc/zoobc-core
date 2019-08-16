package crypto

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/ed25519"
)

type (
	SignatureInterface interface {
		Sign(payload []byte, accountAddress, seed string) []byte
		SignByNode(payload []byte, nodeSeed string) []byte
		VerifySignature(payload, signature []byte, accountAddress string) bool
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
func (sig *Signature) Sign(payload []byte, accountAddress, seed string) []byte {
	accountPrivateKey := ed25519GetPrivateKeyFromSeed(seed)
	signature := ed25519.Sign(accountPrivateKey, payload)
	return signature
}

// SignByNode special method for signing block only, there will be no multiple signature options
func (*Signature) SignByNode(payload []byte, nodeSeed string) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(constant.NodeSignatureTypeDefault))
	nodePrivateKey := ed25519GetPrivateKeyFromSeed(nodeSeed)
	signature := ed25519.Sign(nodePrivateKey, payload)
	buffer.Write(signature)
	return buffer.Bytes()
}

// VerifySignature accept payload (before without signature), signature and the account id
// then verify the signature + public key against the payload based on the
func (*Signature) VerifySignature(payload, signature []byte, accountAddress string) bool {
	accountType := signature[:4]
	switch util.ConvertBytesToUint32(accountType) {
	case 0: // zoobc
		accountPublicKey, _ := util.GetPublicKeyFromAddress(accountAddress)
		result := ed25519.Verify(accountPublicKey, payload, signature[4:])
		return result
	default:
		accountPublicKey, _ := util.GetPublicKeyFromAddress(accountAddress)
		result := ed25519.Verify(accountPublicKey, payload, signature[4:])
		return result
	}
}

// VerifyNodeSignature Verify a signature of a block or message signed with a node private key
// Note: this function is a wrapper around the ed25519 algorithm
func (*Signature) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	result := ed25519.Verify(nodePublicKey, payload, signature[4:])
	return result
}
