package crypto

import (
	"encoding/binary"
	"math/rand"

	"golang.org/x/crypto/sha3"
)

const (
	PseudoRandomXoroshiro128 = iota
	PseudoRandomSha3256
)

// PseudoRandomGenerator using multple algorithms
func PseudoRandomGenerator(id, offset uint64, algo int) (pr uint64) {
	seed := uint64(id ^ offset)
	switch algo {
	case PseudoRandomXoroshiro128:
		src := Rng128P{}
		src.Seed(int64(seed))
		rng := rand.New(&src)
		return rng.Uint64()
	case PseudoRandomSha3256:
		seedBuffer := make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBuffer, seed)
		seedHash := sha3.Sum256(seedBuffer)
		pr = binary.LittleEndian.Uint64(seedHash[:])
		return pr
	}
	return 0
}
