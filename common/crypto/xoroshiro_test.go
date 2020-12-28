// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
