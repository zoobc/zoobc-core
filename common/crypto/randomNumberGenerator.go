package crypto

import (
	"bytes"
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
	blockSeedBigInt := new(big.Int).SetBytes(buffer.Bytes())
	r.rand.Seed(blockSeedBigInt.Int64())
	return nil
}

func (r *RandomNumberGenerator) Next() uint64 {
	return r.rand.Uint64()
}
