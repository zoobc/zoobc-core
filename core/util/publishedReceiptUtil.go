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
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// PublishedReceiptUtilInterface act as interface for data getter on published_receipt entity
	PublishedReceiptUtilInterface interface {
		GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		GetPublishedReceiptsByBlockHeightRange(fromBlockHeight, toBlockHeight uint32) ([]*model.PublishedReceipt, error)
		GetPublishedReceiptByLinkedRMR(root []byte) (*model.PublishedReceipt, error)
		InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt, tx bool) error
	}
	PublishedReceiptUtil struct {
		PublishedReceiptQuery query.PublishedReceiptQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
)

func NewPublishedReceiptUtil(
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	queryExecutor query.ExecutorInterface,
) *PublishedReceiptUtil {
	return &PublishedReceiptUtil{
		PublishedReceiptQuery: publishedReceiptQuery,
		QueryExecutor:         queryExecutor,
	}
}

// GetPublishedReceiptByBlockHeight get data from published_receipt table by the block height they were published / broadcasted
func (psu *PublishedReceiptUtil) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	var publishedReceipts []*model.PublishedReceipt

	// get published receipts of the block
	publishedReceiptQ, publishedReceiptArg := psu.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(blockHeight)
	rows, err := psu.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArg...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	publishedReceipts, err = psu.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return nil, err
	}
	return publishedReceipts, nil
}

// GetPublishedReceiptByBlockHeightRange get data from published_receipt table by the block height they were published / broadcasted
func (psu *PublishedReceiptUtil) GetPublishedReceiptsByBlockHeightRange(fromBlockHeight, toBlockHeight uint32) ([]*model.PublishedReceipt, error) {
	var publishedReceipts []*model.PublishedReceipt

	// get published receipts of the block
	publishedReceiptQ, publishedReceiptArg := psu.PublishedReceiptQuery.GetPublishedReceiptByBlockHeightRange(
		fromBlockHeight, toBlockHeight,
	)
	rows, err := psu.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArg...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	publishedReceipts, err = psu.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return nil, err
	}
	return publishedReceipts, nil
}

func (psu *PublishedReceiptUtil) GetPublishedReceiptByLinkedRMR(root []byte) (*model.PublishedReceipt, error) {
	var (
		publishedReceipt = &model.PublishedReceipt{
			Receipt:            &model.Receipt{},
			IntermediateHashes: nil,
			BlockHeight:        0,
			PublishedIndex:     0,
		}
		err error
	)
	// look up root in published_receipt table
	rcQ, rcArgs := psu.PublishedReceiptQuery.GetPublishedReceiptByLinkedRMR(root)
	row, _ := psu.QueryExecutor.ExecuteSelectRow(rcQ, false, rcArgs...)
	err = psu.PublishedReceiptQuery.Scan(publishedReceipt, row)
	if err != nil {
		return nil, err
	}
	return publishedReceipt, nil
}

func (psu *PublishedReceiptUtil) InsertPublishedReceipt(publishedReceipt *model.PublishedReceipt, tx bool) error {
	var err error
	insertPublishedReceiptQ, insertPublishedReceiptArgs := psu.PublishedReceiptQuery.InsertPublishedReceipt(
		publishedReceipt,
	)
	if tx {
		err = psu.QueryExecutor.ExecuteTransaction(insertPublishedReceiptQ, insertPublishedReceiptArgs...)
	} else {
		_, err = psu.QueryExecutor.ExecuteStatement(insertPublishedReceiptQ, insertPublishedReceiptArgs...)
	}
	if err != nil {
		return err
	}
	return nil
}
