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
	"bytes"
	"github.com/zoobc/zoobc-core/common/crypto"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	NodeAuthValidationInterface interface {
		ValidateProofOfOwnership(
			poown *model.ProofOfOwnership,
			nodePublicKey []byte,
			queryExecutor query.ExecutorInterface,
			blockQuery query.BlockQueryInterface,
		) error
		ValidateProofOfOrigin(
			poorig *model.ProofOfOrigin,
			nodePublicKey,
			challengeResponse []byte,
		) error
	}

	// Signature object handle signing and verifying different signature
	NodeAuthValidation struct {
		Signature crypto.SignatureInterface
	}
)

func NewNodeAuthValidation(
	signature crypto.SignatureInterface,
) *NodeAuthValidation {
	return &NodeAuthValidation{
		Signature: signature,
	}
}

// ValidateProofOfOwnership validates a proof of ownership message
func (nav *NodeAuthValidation) ValidateProofOfOwnership(
	poown *model.ProofOfOwnership,
	nodePublicKey []byte,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) error {

	if !nav.Signature.VerifyNodeSignature(poown.MessageBytes, poown.Signature, nodePublicKey) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}

	message, err := util.ParseProofOfOwnershipMessageBytes(poown.MessageBytes)
	if err != nil {
		return err
	}

	lastBlock, err := util.GetLastBlock(queryExecutor, blockQuery)
	if err != nil {
		return err
	}
	// Expiration, in number of blocks, of a proof of ownership message
	if lastBlock.Height-message.BlockHeight > constant.ProofOfOwnershipExpiration {
		return blocker.NewBlocker(blocker.ValidationErr, "ProofOfOwnershipExpired")
	}

	poownBlockRef, err := util.GetBlockByHeight(message.BlockHeight, queryExecutor, blockQuery)
	if err != nil {
		return err
	}
	if !bytes.Equal(poownBlockRef.BlockHash, message.BlockHash) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidBlockHash")
	}
	return nil
}

// ValidateProofOfOrigin validates a proof of origin message
func (nav *NodeAuthValidation) ValidateProofOfOrigin(
	poorig *model.ProofOfOrigin,
	nodePublicKey,
	challengeResponse []byte,
) error {
	if poorig == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "ProofOfOriginNotProvided")
	}
	if poorig.Timestamp < time.Now().Unix() {
		return blocker.NewBlocker(blocker.ValidationErr, "ProofOfOriginExpired")
	}

	if !bytes.Equal(challengeResponse, poorig.MessageBytes) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidChallengeResponse")
	}

	if !nav.Signature.VerifyNodeSignature(
		util.GetProofOfOriginUnsignedBytes(poorig),
		poorig.Signature,
		nodePublicKey,
	) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidSignature")
	}

	return nil
}
