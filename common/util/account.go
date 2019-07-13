package util

import (
	"encoding/base64"
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/ed25519"

	"golang.org/x/crypto/sha3"
)

// CreateAccountIDFromAddress return the account ID byte which is the hash of
// account type (int32) and the account address (default: base64(public key))
// for type 0
func CreateAccountIDFromAddress(accountType uint32, address string) []byte {
	accountTypeByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(accountTypeByte, uint32(accountType))
	digest := sha3.New256()
	_, _ = digest.Write(accountTypeByte)
	_, _ = digest.Write([]byte(address))
	accountID := digest.Sum([]byte{})
	return accountID
}

// GetAccountIDByPublicKey return the account ID byte which is the hash of
// account type (uint32) and the account address (default: base64(public key))
// for type 0
func GetAccountIDByPublicKey(accountType uint32, publicKey []byte) []byte {
	accountTypeByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(accountTypeByte, accountType)
	var address string
	if accountType == 0 { // default account type: zoobc
		address, _ = GetAddressFromPublicKey(publicKey)
	}
	digest := sha3.New256()
	_, _ = digest.Write(accountTypeByte[:2])
	_, _ = digest.Write([]byte(address))
	accountID := digest.Sum([]byte{})
	return accountID
}

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

// GetAddressFromPublicKey Get the formatted address from a raw public key
func GetAddressFromPublicKey(publicKey []byte) (string, error) {
	// public key should be 32 long
	if len(publicKey) != 32 {
		return "", errors.New("ErrInvalidPublicKeyLength")
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
