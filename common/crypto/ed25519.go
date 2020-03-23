package crypto

import (
	"encoding/base64"

	slip10 "github.com/zoobc/zoo-slip10"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

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

// GetPrivateKeyFromSeedUseSlip10 generate privite key form seed using slip10
func (*Ed25519Signature) GetPrivateKeyFromSeedUseSlip10(seed string) ([]byte, error) {
	var (
		seedBytes      = slip10.NewSeed(seed, slip10.DefaultPassword)
		slip10Key, err = slip10.DeriveForPath(slip10.StellarPrimaryAccountPath, seedBytes)
	)
	if err != nil {
		return nil, err
	}
	return slip10Key.Key, nil

}

// GetPublicKeyFromSeed Get the raw public key corresponding to a seed (secret phrase)
func (es *Ed25519Signature) GetPublicKeyFromSeed(seed string) []byte {
	// Get the private key from the seed
	privateKey := es.GetPrivateKeyFromSeed(seed)
	// Get the public key from the private key
	return privateKey[32:]
}

// GetAddressFromSeed Get the address corresponding to a seed (secret phrase)
func (es *Ed25519Signature) GetAddressFromSeed(seed string) string {
	result, _ := es.GetAddressFromPublicKey(es.GetPublicKeyFromSeed(seed))
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

// GetPublicKeyFromAddress Get the raw public key from a formatted address
func (*Ed25519Signature) GetPublicKeyFromAddress(address string) ([]byte, error) {
	// decode base64 back to byte
	publicKey, err := base64.URLEncoding.DecodeString(address)
	if err != nil {

		return nil, blocker.NewBlocker(blocker.AppErr, err.Error())
	}
	// Needs to check the checksum bit at the end, and if valid,
	if publicKey[32] != util.GetChecksumByte(publicKey[:32]) {
		return nil, blocker.NewBlocker(blocker.AppErr, "address checksum failed")
	}
	return publicKey[:32], nil
}

// GetAddressFromPublicKey Get the formatted address from a raw public key
func (*Ed25519Signature) GetAddressFromPublicKey(publicKey []byte) (string, error) {
	// public key should be 32 long
	if len(publicKey) != 32 {
		return "", blocker.NewBlocker(
			blocker.ServerError,
			"invalid public key length",
		)
	}
	// Make 33 byte buffer for Public Key + Checksum Byte
	rawAddress := make([]byte, 33)
	copy(rawAddress, publicKey)

	// Add Checksum Byte to end
	rawAddress[32] = util.GetChecksumByte(publicKey)

	// Convert the raw address (public key + checksum) to Base64 notation
	address := base64.URLEncoding.EncodeToString(rawAddress)
	return address, nil
}
