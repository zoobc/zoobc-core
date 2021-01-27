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

package fee

import (
	"errors"
	"math"
	"time"

	"github.com/zoobc/zoobc-core/common/model"
)

func CalculateTxMinimumFee(tx *model.Transaction, feeScale int64) (int64, error) {
	feePerCharacterMultiplier := float64(0.1)
	minFeeMultiplier := 1
	escrowInstructionFeeMultiplier := float64(0)
	escrowLifetimeMultiplier := 24
	escrowLifeDays := float64(1)
	if tx.Escrow != nil {
		// escrowInstructionFeeMultiplier: fee for instruction in the instruction (1 ZBC for every 1000 character)
		escrowInstructionFeeMultiplier = float64(len(tx.Escrow.Instruction)) * feePerCharacterMultiplier

		escrowTimeout := time.Unix(tx.Escrow.Timeout, 0)
		txTimestamp := time.Unix(tx.Timestamp, 0)
		hoursDifference := txTimestamp.Sub(escrowTimeout).Hours()
		if hoursDifference < 0 {
			return 0, errors.New("InvalidTime:Passed")
		}

		// increment fee by feePerBlockPeriod every 24 hours
		escrowLifeDays = math.Ceil(hoursDifference / float64(escrowLifetimeMultiplier))
	}

	// txMessageFeeMultiplier: fee multiplier for message in the instruction (1 ZBC for every 1000 character)
	txMessageFeeMultiplier := float64(len(tx.Message)) * feePerCharacterMultiplier
	fee := int64(math.Ceil((txMessageFeeMultiplier + escrowInstructionFeeMultiplier + float64(minFeeMultiplier)) * escrowLifeDays * float64(feeScale)))
	return fee, nil
}
