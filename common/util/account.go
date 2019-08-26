package util

import (
	"encoding/base64"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"

	"golang.org/x/crypto/ed25519"

	"golang.org/x/crypto/sha3"
)

// GetPrivateKeyFromSeed Get the raw private key corresponding to a seed (secret phrase)
func GetPrivateKeyFromSeed(seed string) ([]byte, error) {
	// Convert seed (secret phrase) to byte array
	seedBuffer := []byte(seed)

	// Compute SHA3-256 hash of seed (secret phrase)
	seedHash := sha3.Sum256(seedBuffer)

	// Generate a private key from the hash of the seed
	privateKey := ed25519.NewKeyFromSeed(seedHash[:])

	return privateKey, nil
}

// GetAddressFromSeed Get the address corresponding to a seed (secret phrase)
func GetAddressFromSeed(seed string) string {
	result, _ := GetAddressFromPublicKey(GetPublicKeyFromSeed(seed))
	return result
}

// GetPublicKeyFromSeed Get the raw public key corresponding to a seed (secret phrase)
func GetPublicKeyFromSeed(seed string) []byte {

	// Get the private key from the seed
	privateKey, _ := GetPrivateKeyFromSeed(seed)

	// Get the public key from the private key
	return privateKey[32:]
}

// GetAddressFromPublicKey Get the formatted address from a raw public key
func GetAddressFromPublicKey(publicKey []byte) (string, error) {
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
	rawAddress[32] = GetChecksumByte(publicKey)

	// Convert the raw address (public key + checksum) to Base64 notation
	address := base64.URLEncoding.EncodeToString(rawAddress)

	return address, nil
}

// GetPublicKeyFromAddress Get the raw public key from a formatted address
func GetPublicKeyFromAddress(address string) ([]byte, error) {
	// decode base64 back to byte
	publicKey, err := base64.URLEncoding.DecodeString(address)
	if err != nil {
		return nil, err
	}
	// Needs to check the checksum bit at the end, and if valid,
	if publicKey[32] != GetChecksumByte(publicKey[:32]) {
		return nil, fmt.Errorf("address checksum failed")
	}
	return publicKey[:32], nil
}

// GetChecksumByte Calculate a checksum byte from a collection of bytes
// checksum 255 = 255, 256 = 0, 257 = 1 and so on.
func GetChecksumByte(bytes []byte) byte {
	n := len(bytes)
	var a byte
	for i := 0; i < n; i++ {
		a += bytes[i]
	}
	return a
}
