package util

import (
	"crypto/rand"
	"math"
	"math/big"
)

// GetSecureRandom generates a int64 secure random
// TODO: implement the real function to generate a secure random. for now we generate a pseudo-secure number
func GetSecureRandom() int64 {
	max := *big.NewInt(math.MaxInt64)
	newNumber, _ := rand.Int(rand.Reader, &max)
	return newNumber.Int64()
}
