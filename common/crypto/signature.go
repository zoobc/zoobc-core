package crypto

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zed25519/zed"

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
	return accountType.Sign(payload, seed, optionalParams...)
}

// SignByNode special method for signing block only, there will be no multiple signature options
func (*Signature) SignByNode(payload []byte, nodeSeed string) []byte {
	var (
		buffer           = bytes.NewBuffer([]byte{})
		ed25519Signature = signaturetype.NewEd25519Signature()
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
	return accountType.VerifySignature(payload, signature, accountAddress)
}

// VerifyNodeSignature Verify a signature of a block or message signed with a node private key
// Note: this function is a wrapper around the ed25519 algorithm
func (*Signature) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	var result = signaturetype.NewEd25519Signature().Verify(nodePublicKey, payload, signature)
	return result
}

// GenerateAccountFromSeed to generate account based on provided seed
func (*Signature) GenerateAccountFromSeed(accountType accounttype.AccountTypeInterface, seed string, optionalParams ...interface{}) (
	privateKey, publicKey []byte,
	publicKeyString, encodedAddress string,
	fullAccountAddress []byte,
	err error,
) {
	err = accountType.GenerateAccountFromSeed(seed, optionalParams...)
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	privateKey, err = accountType.GetAccountPrivateKey()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	publicKey = accountType.GetAccountPublicKey()
	publicKeyString, err = accountType.GetAccountPublicKeyString()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	encodedAddress, err = accountType.GetEncodedAddress()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	fullAccountAddress, err = accountType.GetAccountAddress()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	return
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
