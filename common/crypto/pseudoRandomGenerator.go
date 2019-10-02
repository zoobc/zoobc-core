package crypto

import "math/rand"

// PseudoRandomGenerator uses xoroshiro128+ PRG to generate random uint64 numbers
func PseudoRandomGenerator(id, offset uint64) uint64 {
	seed := int64(id ^ offset)
	src := Rng128P{}
	src.Seed(seed)
	rng := rand.New(&src)
	return rng.Uint64()
}
