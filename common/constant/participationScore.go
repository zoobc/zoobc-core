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
package constant

const (
	// ScalarReceiptScore the converter score to avoid calculation in float number, this value is following OneZBC to
	// maintain the number scale like balance does.
	ScalarReceiptScore = float32(OneZBC)
	// LinkedReceiptScore the score for each receipt that proved have relation with previous published receipt via merkle root
	LinkedReceiptScore float32 = 2
	// LinkedReceiptScore the score for each receipt that can't proved have relation with previous published receipt via merkle root
	UnlinkedReceiptScore float32 = 0.5
	// MaxScoreChange the maximum score that node wll get.
	// note that in small networks if this value is too high it will lead to nodes being expelled from registry quickly
	// in production 100000000 * int64(ScalarReceiptScore). reduce to 10 * int64(ScalarReceiptScore) to test with less than 10 nodes
	MaxScoreChange = 10000 * int64(ScalarReceiptScore)
	// punishment amount
	ParticipationScorePunishAmount = -1 * MaxScoreChange / 2
	// MaxParticipationScore maximum achievable score, this will be important to maintain smithing process so it doesn't
	// smith too fast
	MaxParticipationScore int64 = 10000000000 * int64(ScalarReceiptScore)
	// Starting score for newly registered nodes
	DefaultParticipationScore int64 = MaxParticipationScore / 5
	// Starting score for pre seed nodes (registered at genesis)
	GenesisParticipationScore int64 = MaxParticipationScore
	// BetaBlockBonus beta only additional participation score
	BetaBlockBonus = MaxParticipationScore / 420
	// BetaBlockBonusSkipMultiplier beta only additional participation score decrease
	BetaBlockBonusSkipMultiplier = 5
)
