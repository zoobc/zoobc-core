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
package configure

var (
	target string
	beta   = []string{
		"n0.beta.proofofparticipation.network:8002",
		"n1.beta.proofofparticipation.network:8002",
		"n2.beta.proofofparticipation.network:8002",
		"n3.beta.proofofparticipation.network:8002",
		"n4.beta.proofofparticipation.network:8002",
		"n5.beta.proofofparticipation.network:8002",
		"n6.beta.proofofparticipation.network:8002",
		"n7.beta.proofofparticipation.network:8002",
		"n8.beta.proofofparticipation.network:8002",
		"n9.beta.proofofparticipation.network:8002",
		"n10.beta.proofofparticipation.network:8002",
		"n11.beta.proofofparticipation.network:8002",
		"n12.beta.proofofparticipation.network:8002",
		"n13.beta.proofofparticipation.network:8002",
		"n14.beta.proofofparticipation.network:8002",
		"n15.beta.proofofparticipation.network:8002",
		"n16.beta.proofofparticipation.network:8002",
		"n17.beta.proofofparticipation.network:8002",
		"n18.beta.proofofparticipation.network:8002",
		"n19.beta.proofofparticipation.network:8002",
	}
	alpha = []string{
		"n0.alpha.proofofparticipation.network:8001",
		"n1.alpha.proofofparticipation.network:8001",
		"n2.alpha.proofofparticipation.network:8001",
		"n3.alpha.proofofparticipation.network:8001",
		"n4.alpha.proofofparticipation.network:8001",
		"n5.alpha.proofofparticipation.network:8001",
	}
	dev = []string{
		"172.104.34.10:8001",
		"45.79.39.58:8001",
		"85.90.246.90:8001",
	}

	// maxAttemptPromptFailed the maximum allowed to try re-input prompt
	maxAttemptPromptFailed = 3
)
