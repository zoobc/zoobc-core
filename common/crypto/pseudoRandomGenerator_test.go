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

import "testing"

func TestPseudoRandomGenerator(t *testing.T) {
	type args struct {
		id     uint64
		offset uint64
		algo   int
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomSha3256-NoOffset}",
			args: args{
				offset: 0,
				id:     3014845244095079110,
				algo:   PseudoRandomSha3256,
			},
			want: 18041622792886434681,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomSha3256-NoOffset-2}",
			args: args{
				offset: 0,
				id:     1941309198183084506,
				algo:   PseudoRandomSha3256,
			},
			want: 3953548740169852696,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomSha3256-Offset}",
			args: args{
				offset: 132553774296354339,
				id:     3014845244095079110,
				algo:   PseudoRandomSha3256,
			},
			want: 14251496166035092223,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomXoroshiro128-NoOffset}",
			args: args{
				offset: 0,
				id:     3014845244095079110,
				algo:   PseudoRandomXoroshiro128,
			},
			want: 17061115035045365337,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomXoroshiro128-NoOffset-2}",
			args: args{
				offset: 0,
				id:     1941309198183084506,
				algo:   PseudoRandomXoroshiro128,
			},
			want: 2623596903506267843,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomXoroshiro128-Offset}",
			args: args{
				offset: 132553774296354339,
				id:     3014845244095079110,
				algo:   PseudoRandomXoroshiro128,
			},
			want: 8913237946701621685,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PseudoRandomGenerator(tt.args.id, tt.args.offset, tt.args.algo); got != tt.want {
				t.Errorf("PseudoRandomGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}
