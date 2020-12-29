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
package service

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"
)

type (
	// PublishedReceiptServiceInterface act as interface for processing the published receipt data
	PublishedReceiptServiceInterface interface {
		ProcessPublishedReceipts(block *model.Block) (int, error)
	}

	PublishedReceiptService struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		ReceiptUtil           util.ReceiptUtilInterface
		PublishedReceiptUtil  util.PublishedReceiptUtilInterface
		ReceiptService        ReceiptServiceInterface
		QueryExecutor         query.ExecutorInterface
	}
)

func NewPublishedReceiptService(
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil util.ReceiptUtilInterface,
	publishedReceiptUtil util.PublishedReceiptUtilInterface,
	receiptService ReceiptServiceInterface,
	queryExecutor query.ExecutorInterface,
) *PublishedReceiptService {
	return &PublishedReceiptService{
		PublishedReceiptQuery: publishedReceiptQuery,
		ReceiptUtil:           receiptUtil,
		PublishedReceiptUtil:  publishedReceiptUtil,
		ReceiptService:        receiptService,
		QueryExecutor:         queryExecutor,
	}
}

// ProcessPublishedReceipts takes published receipts in a block and validate them, this function will run in a db transaction
// so ensure queryExecutor.Begin() is called before calling this function.
func (ps *PublishedReceiptService) ProcessPublishedReceipts(block *model.Block) (int, error) {
	var (
		linkedCount int
		err         error
	)
	for index, rc := range block.GetPublishedReceipts() {
		// validate sender and recipient of receipt
		err = ps.ReceiptService.ValidateReceipt(rc.GetReceipt())
		if err != nil {
			return 0, err
		}
		// check if linked
		if rc.IntermediateHashes != nil && len(rc.IntermediateHashes) > 0 {
			merkle := &commonUtils.MerkleRoot{}
			rcByte := ps.ReceiptUtil.GetSignedReceiptBytes(rc.GetReceipt())
			rcHash := sha3.Sum256(rcByte)
			root, err := merkle.GetMerkleRootFromIntermediateHashes(
				rcHash[:],
				rc.ReceiptIndex,
				merkle.RestoreIntermediateHashes(rc.IntermediateHashes),
			)
			if err != nil {
				return 0, err
			}
			// look up root in published_receipt table
			_, err = ps.PublishedReceiptUtil.GetPublishedReceiptByLinkedRMR(root)
			if err != nil {
				return 0, err
			}
			// add to linked receipt count for calculation later
			linkedCount++
		}
		// store in database
		// assign index and height, index is the order of the receipt in the block,
		// it's different with receiptIndex which is used to validate merkle root.
		rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)

		err := ps.PublishedReceiptUtil.InsertPublishedReceipt(rc, true)
		if err != nil {
			return 0, err
		}
	}
	return linkedCount, nil
}
