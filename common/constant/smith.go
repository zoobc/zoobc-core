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

import (
	"time"
)

var (
	CoinbaseTotalDistribution        int64   = 3000000 * OneZBC // 3 million * 10^8 in production
	CoinbaseTime                     int64   = 5 * OneYear      // 5 years in production
	CoinbaseSigmoidStart             float64 = 3
	CoinbaseSigmoidEnd               float64 = 6
	CoinbaseNumberRewardsPerSecond   int64   = 1 // probably this will always be 1
	CoinbaseMaxNumberRewardsPerBlock int64   = 600

	GenerateBlockTimeoutSec     = int64(15)
	CumulativeDifficultyDivisor = int64(1000000)
	// BlockPoolScanPeriod define the periodic time to scan the whole block pool for legal block to persist to the chain
	BlockPoolScanPeriod = 5 * time.Second
	// TimeOutBlockWaitingTransactions is the timeout of block while waiting transactions
	TimeOutBlockWaitingTransactions = int64(2 * 60) // 2 minute
	// CheckTimedOutBlock to use in scheduler to check timedout block while waiting transaction
	CheckTimedOutBlock        = 30 * time.Second
	SpineChainSmithIdlePeriod = 500 * time.Millisecond
	// SpineChainSmithingPeriod intervals between spine blocks in seconds
	SpineChainSmithingPeriod = int64(86400)
	MainChainSmithIdlePeriod = 500 * time.Millisecond
	// MainChainSmithingPeriod one main block every 15 seconds + block pool delay (max +30 seconds)
	MainChainSmithingPeriod = int64(15)
	// EmptyBlockSkippedBlocksmithLimit state the number of allowed skipped blocksmith until only empty block can be generated
	// 0 will set node to always create empty block
	EmptyBlockSkippedBlocksmithLimit = int64(10) // 10 in production
	/*
		Mainchain smithing
	*/

	MainSmithingBlockCreationTime = int64(30)
	MainSmithingNetworkTolerance  = int64(15)
	MainSmithingBlocksmithTimeGap = int64(10)

	/*
		Spinechain smithing
	*/

	SpineSmithingBlockCreationTime  = int64(30)
	SpineSmithingNetworkTolerance   = int64(15)
	SpineSmithingBlocksmithTimeGap  = int64(10)
	SpineReferenceBlockHeightOffset = uint32(5)
)
