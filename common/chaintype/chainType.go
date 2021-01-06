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
package chaintype

import "time"

// ChainType interface define the different behavior of each chain
type (
	ChainType interface {
		// GetTypeInt return the value of the chain type in int
		GetTypeInt() int32
		// GetTablePrefix return the value of current chain table prefix in the database
		GetTablePrefix() string
		// GetSmithingPeriod the time since last block that blocksmith can start to smith
		GetSmithingPeriod() int64
		// GetBlocksmithTimeGap return the time gap one to the next blocksmith
		GetBlocksmithTimeGap() int64
		// GetBlocksmithBlockCreationTime return the maximum allowed time to create block
		GetBlocksmithBlockCreationTime() int64
		// GetBlocksmithNetworkTolerance return the maximum allowed time to broadcast block
		GetBlocksmithNetworkTolerance() int64
		// GetName return the name of the chain : used in parsing chaintype across node
		GetName() string
		// GetGenesisBlockID return the block ID of genesis block in the chain
		GetGenesisBlockID() int64

		GetGenesisBlockSeed() []byte
		GetGenesisNodePublicKey() []byte
		GetGenesisBlockTimestamp() int64
		GetGenesisBlockSignature() []byte
		// HasTransactions true if this chain type implements transactions (thus has a mempool)
		HasTransactions() bool
		// HasSnapshots true if this chain type implements snapshots
		HasSnapshots() bool
		// If HasSnapshot is true, this must return the interval, in blocks, the snapshot has to be taken
		// If HasSnapshot is false, this will return zero
		GetSnapshotInterval() uint32
		// If HasSnapshot is true, this returns the seconds to pass, from the snapshot's process start (a block's timestamp),
		// before considering the snapshot's expired (= snapshot's process timeout)
		// If HasSnapshot is false, this will return zero
		GetSnapshotGenerationTimeout() time.Duration
	}
)
