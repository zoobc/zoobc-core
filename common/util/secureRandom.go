package util

import (
	"math/rand"
)

// GetSecureRandom generates a int64 secure random
// TODO: implement the real function to generate a secure random. for now we generate a pseudo-secure number
func GetSecureRandom() int64 {
	return rand.Int63()
}
