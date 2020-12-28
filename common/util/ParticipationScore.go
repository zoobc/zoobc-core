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
package util

import (
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
)

// CalculateParticipationScore to calculate score change of node
func CalculateParticipationScore(linkedReceipt, unlinkedReceipt, maxReceipt uint32) (int64, error) {
	if maxReceipt == 0 {
		return constant.MaxScoreChange, nil
	}
	if (linkedReceipt + unlinkedReceipt) > maxReceipt {
		return 0, blocker.NewBlocker(
			blocker.ValidationErr,
			"CalculateParticipationScore, the number of receipt exceeds",
		)
	}

	// Maximum score will get when create a block
	maxBlockScore := int64(float32(maxReceipt) * constant.LinkedReceiptScore * constant.ScalarReceiptScore)
	halfMaxBlockScore := maxBlockScore / 2

	linkedBlockScore := float32(linkedReceipt) * constant.LinkedReceiptScore * constant.ScalarReceiptScore
	unlinkedBlockScore := float32(unlinkedReceipt) * constant.UnlinkedReceiptScore * constant.ScalarReceiptScore
	blockScore := int64(linkedBlockScore + unlinkedBlockScore)

	scoreDiffBig := new(big.Int).SetInt64(blockScore - halfMaxBlockScore)
	scoreDiffBigMul := new(big.Int).Mul(scoreDiffBig, new(big.Int).SetInt64(constant.MaxScoreChange))
	scoreChangeOfANode := new(big.Int).Div(scoreDiffBigMul, new(big.Int).SetInt64(halfMaxBlockScore))
	return scoreChangeOfANode.Int64(), nil
}

func GetReceiptValue(linkedReceipt, unlinkedReceipt uint32) int64 {
	linkedBlockScore := float32(linkedReceipt) * constant.LinkedReceiptScore * constant.ScalarReceiptScore
	unlinkedBlockScore := float32(unlinkedReceipt) * constant.UnlinkedReceiptScore * constant.ScalarReceiptScore
	blockScore := int64(linkedBlockScore + unlinkedBlockScore)
	return blockScore
}
