package crypto

import (
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

func ed25519GetPrivateKeyFromSeed(seed string) []byte {
	// Convert seed (secret phrase) to byte array
	seedBuffer := []byte(seed)
	// Compute SHA3-256 hash of seed (secret phrase)
	seedHash := sha3.Sum256(seedBuffer)
	// Generate a private key from the hash of the seed
	return ed25519.NewKeyFromSeed(seedHash[:])
}
