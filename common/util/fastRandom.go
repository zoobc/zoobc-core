package util

import (
	"math/rand"
	"time"
)

// GetFastRandom generates a int64 random number
func GetFastRandom(seed *rand.Rand, max int) int64 {
	return int64(seed.Intn(max))
}

// GetFastRandomSeed generates a new randome seed, a mnemonic that can be used to derive a private key
func GetFastRandomSeed() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().Unix()))
}
