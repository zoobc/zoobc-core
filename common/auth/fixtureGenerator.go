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
package auth

import (
	"strconv"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	// STEF TODO: insert real accountAddress here
	senderAddress1 = make([]byte, 36)
	// var senderSeed1 = "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"
	nodeSeed1   = "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
	nodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	block1 = &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: strconv.Itoa(100000000),
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  nodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
)

func GetFixturesProofOfOwnershipValidation(height uint32, hash []byte, block *model.Block) (poown *model.ProofOfOwnership) {
	if block == nil {
		block = block1
	}
	if block.Transactions == nil {
		block.Transactions = make([]*model.Transaction, 0)
	}
	if len(hash) == 0 {
		blockHash, _ := util.GetBlockHash(block, &chaintype.MainChain{})
		hash = blockHash
	}
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      hash,
		BlockHeight:    height,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	return
}
