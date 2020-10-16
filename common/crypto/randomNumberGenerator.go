package crypto

import (
	"bytes"
	"golang.org/x/crypto/sha3"
	"math/big"
)

type (
	RandomNumberGenerator struct {
		rand *Rng128P
	}
)

func NewRandomNumberGenerator() *RandomNumberGenerator {
	return &RandomNumberGenerator{
		rand: &Rng128P{},
	}
}

func (r *RandomNumberGenerator) Reset(prefix string, seed []byte) error {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write([]byte(prefix))
	buffer.Write(seed)
	randSeedHash := sha3.Sum256(buffer.Bytes())
	randSeedBigInt := new(big.Int).SetBytes(randSeedHash[:])
	r.rand.Seed(randSeedBigInt.Int64())
	return nil
}

func (r *RandomNumberGenerator) Next() uint64 {
	return r.rand.Uint64()
}
