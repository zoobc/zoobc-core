package crypto

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"math/rand"
	"testing"
	"time"
)

func TestRandomNumberGenerator(t *testing.T) {
	rngInstance := NewRandomNumberGenerator()
	randomSeed := make([]byte, 32)
	rand.Seed(time.Now().Unix())
	rand.Read(randomSeed)
	err := rngInstance.Reset(constant.BlocksmithSelectionSeedPrefix, randomSeed)
	if err != nil {
		t.Errorf("fail to reset rng seed: %v\n", err)
	}
	var result = make([]uint64, 10)
	for i := 0; i < len(result); i++ {
		result[i] = rngInstance.Next()
	}
	// check for consistency
	err = rngInstance.Reset(constant.BlocksmithSelectionSeedPrefix, randomSeed)
	if err != nil {
		t.Errorf("fail to reset rng seed: %v\n", err)
	}
	for i := 0; i < len(result); i++ {
		if result[i] != rngInstance.Next() {
			t.Errorf("same seed produce different random sequence")
		}
	}
}
