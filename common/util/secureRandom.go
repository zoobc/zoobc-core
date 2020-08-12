package util

import (
	"crypto/rand"
	"math"
	"math/big"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/tyler-smith/go-bip39"
)

// GetSecureRandom generates a int64 secure random
// TODO: implement the real function to generate a secure random. for now we generate a pseudo-secure number
func GetSecureRandom() int64 {
	max := *big.NewInt(math.MaxInt64)
	newNumber, _ := rand.Int(rand.Reader, &max)
	return newNumber.Int64()
}

// GetSecureRandomSeed generates a new random seed, a mnemonic that can be used to derive a private key
func GetSecureRandomSeed() string {
	entropy, _ := bip39.NewEntropy(constant.SecureRandomSeedBitSize)
	seed, _ := bip39.NewMnemonic(entropy)
	return seed
}
