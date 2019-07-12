package util

import (
	"encoding/base64"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

// CreateAccountIDFromAddress return the account ID byte which is the hash of
// account type (int32) and the account address (default: base64(public key))
// for type 0
func CreateAccountIDFromAddress(accountType int32, address string) []byte {
	accountTypeByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(accountTypeByte, uint32(accountType))
	digest := sha3.New256()
	_, _ = digest.Write(accountTypeByte)
	_, _ = digest.Write([]byte(address))
	accountID := digest.Sum([]byte{})
	return accountID
}

// GetAccountIDByPublicKey return the account ID byte which is the hash of
// account type (int32) and the account address (default: base64(public key))
// for type 0
func GetAccountIDByPublicKey(accountType int32, publicKey []byte) []byte {
	accountTypeByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(accountTypeByte, uint32(accountType))
	var address string
	if accountType == 0 { // default account type: zoobc
		rawAddress := make([]byte, 33)
		copy(rawAddress, publicKey)

		// Add Checksum Byte to end
		rawAddress[32] = GetChecksumByte(publicKey)
		address = base64.URLEncoding.EncodeToString(publicKey)
	}
	digest := sha3.New256()
	_, _ = digest.Write(accountTypeByte)
	_, _ = digest.Write([]byte(address))
	accountID := digest.Sum([]byte{})
	return accountID
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
