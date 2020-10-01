package crypto

import (
	"bytes"
	"encoding/base64"
	"github.com/zoobc/lib/address"
	slip10 "github.com/zoobc/zoo-slip10"
	"github.com/zoobc/zoobc-core/common/blocker"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

// SignatureTypeInterface implements all signature types methods
// TODO: some of these signatures are specific to some account type and should be separated from the main interface
type SignatureTypeInterface interface {
	Sign(accountPrivateKey, payload []byte) []byte
	Verify(accountPublicKey, payload, signature []byte) bool
	GetPrivateKeyFromSeed(seed string) []byte
	GetPrivateKeyFromSeedUseSlip10(seed string) ([]byte, error)
	GetPublicKeyFromPrivateKeyUseSlip10(privateKey []byte) ([]byte, error)
	GetPublicKeyFromSeed(seed string) []byte
	GetAddressFromSeed(prefix, seed string) string
	GetPublicKeyFromPrivateKey(privateKey []byte) ([]byte, error)
	GetPublicKeyString(publicKey []byte) string
	GetPublicKeyFromEncodedAddress(addr string) ([]byte, error)
	GetAddressFromPublicKey(prefix string, publicKey []byte) (string, error)
}

// Ed25519Signature represent of ed25519 signature
type Ed25519Signature struct{}

// NewEd25519Signature is new instance of ed25519 signature
func NewEd25519Signature() *Ed25519Signature {
	return &Ed25519Signature{}
}

// Sign to generates an ed25519 signature for the provided payload
func (*Ed25519Signature) Sign(accountPrivateKey, payload []byte) []byte {
	return ed25519.Sign(accountPrivateKey, payload)
}

// Verify to verify the signature of payload using provided account public key
func (*Ed25519Signature) Verify(accountPublicKey, payload, signature []byte) bool {
	return ed25519.Verify(accountPublicKey, payload, signature)
}

// GetPrivateKeyFromSeed to get private key form seed
func (*Ed25519Signature) GetPrivateKeyFromSeed(seed string) []byte {
	// Convert seed (secret phrase) to byte array
	seedBuffer := []byte(seed)
	// Compute SHA3-256 hash of seed (secret phrase)
	seedHash := sha3.Sum256(seedBuffer)
	// Generate a private key from the hash of the seed
	return ed25519.NewKeyFromSeed(seedHash[:])
}

// GetPrivateKeyFromSeedUseSlip10 generate private key form seed using slip10, this private used by hdwallet
// NOTE: currently this private cannot use to sign message using golang ed25519,
// The output private key is first 32 bytes from private key golang ed25519
func (*Ed25519Signature) GetPrivateKeyFromSeedUseSlip10(seed string) ([]byte, error) {
	var (
		seedBytes      = slip10.NewSeed(seed, slip10.DefaultPassword)
		slip10Key, err = slip10.DeriveForPath(slip10.ZoobcPrimaryAccountPath, seedBytes)
	)
	if err != nil {
		return nil, err
	}
	return slip10Key.Key, nil
}

// GetPublicKeyFromPrivateKeyUseSlip10 get pubic key from slip10 private key
func (*Ed25519Signature) GetPublicKeyFromPrivateKeyUseSlip10(privateKey []byte) ([]byte, error) {
	var (
		reader            = bytes.NewReader(privateKey)
		publicKey, _, err = ed25519.GenerateKey(reader)
	)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

// GetPublicKeyFromSeed Get the raw public key corresponding to a seed (secret phrase)
func (es *Ed25519Signature) GetPublicKeyFromSeed(seed string) []byte {
	// Get the private key from the seed
	privateKey := es.GetPrivateKeyFromSeed(seed)
	// Get the public key from the private key
	return privateKey[32:]
}

// GetAddressFromSeed Get the address corresponding to a seed (secret phrase)
func (es *Ed25519Signature) GetAddressFromSeed(prefix, seed string) string {
	result, _ := es.GetAddressFromPublicKey(prefix, es.GetPublicKeyFromSeed(seed))
	return result
}

// GetPublicKeyFromPrivateKey get public key bytes from private key
func (*Ed25519Signature) GetPublicKeyFromPrivateKey(privateKey []byte) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, blocker.NewBlocker(blocker.AppErr, "invalid ed25519 private key")
	}
	return privateKey[32:], nil
}

// GetPublicKeyString will return string of row bytes public key
func (*Ed25519Signature) GetPublicKeyString(publicKey []byte) string {
	return base64.StdEncoding.EncodeToString(publicKey)
}

// GetPublicKeyFromEncodedAddress Get the raw public key from a formatted address
func (*Ed25519Signature) GetPublicKeyFromEncodedAddress(addr string) ([]byte, error) {
	// decode base64 back to byte
	var (
		publicKey = make([]byte, 32)
		err       error
	)
	if err != nil {
		return nil, err
	}
	err = address.DecodeZbcID(addr, publicKey)
	return publicKey, err
}

// GetAddressFromPublicKey Get the formatted address from a raw public key
func (*Ed25519Signature) GetAddressFromPublicKey(prefix string, publicKey []byte) (string, error) {
	// public key should be 32 long
	id, err := address.EncodeZbcID(prefix, publicKey)
	return id, err
}
