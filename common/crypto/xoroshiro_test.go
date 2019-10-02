package crypto

import (
	"fmt"
	"math/rand"
)

const (
	SEED2 = 1387366483214
)

func ExampleRng128P() {
	src := Rng128P{}
	src.Seed(SEED2)
	rng := rand.New(&src)
	for i := 0; i < 4; i++ {
		fmt.Printf(" %d", rng.Uint32())
	}
	fmt.Println("")
	for i := 0; i < 4; i++ {
		fmt.Printf(" %d", rng.Uint64())
	}
	fmt.Println("")
	// Play craps
	for i := 0; i < 10; i++ {
		fmt.Printf(" %d%d", rng.Intn(6)+1, rng.Intn(6)+1)
	}

	// Output:
	// 3672052799 776653214 1122818236 1139848352
	//  14850484681238877506 7018105211938886447 5908230704518956940 2042158984393296588
	//  65 53 21 56 44 16 23 42 55 41
}

func ExampleRng128SS() {
	src := Rng128SS{}
	src.Seed(SEED2)
	rng := rand.New(&src)
	for i := 0; i < 4; i++ {
		fmt.Printf(" %d", rng.Uint32())
	}
	fmt.Println("")
	for i := 0; i < 4; i++ {
		fmt.Printf(" %d", rng.Uint64())
	}
	fmt.Println("")
	// Play craps
	for i := 0; i < 10; i++ {
		fmt.Printf(" %d%d", rng.Intn(6)+1, rng.Intn(6)+1)
	}

	// Output:
	// 901646676 398979522 1208087553 1093404254
	//  17905646702528074117 5693647338227160345 1089260090730707711 12276528025967720504
	//  41 35 56 61 56 35 31 12 63 54
}
