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
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// SkippedBlockSmithServiceInterface represents interface for BlockService
	SkippedBlockSmithServiceInterface interface {
		GetSkippedBlockSmiths(*model.GetSkippedBlocksmithsRequest) (*model.GetSkippedBlocksmithsResponse, error)
	}

	// SkippedBlockSmithService represents struct of SkippedBlockSmithService
	SkippedBlockSmithService struct {
		QueryExecutor          query.ExecutorInterface
		SkippedBlocksmithQuery *query.SkippedBlocksmithQuery
	}
)

func NewSkippedBlockSmithService(
	skippedBlocksmithQuery *query.SkippedBlocksmithQuery,
	queryExecutor query.ExecutorInterface,
) SkippedBlockSmithServiceInterface {
	return &SkippedBlockSmithService{
		SkippedBlocksmithQuery: skippedBlocksmithQuery,
		QueryExecutor:          queryExecutor,
	}
}

func (sbs *SkippedBlockSmithService) GetSkippedBlockSmiths(
	req *model.GetSkippedBlocksmithsRequest,
) (*model.GetSkippedBlocksmithsResponse, error) {
	var (
		err                error
		rowCount           *sql.Row
		count              uint64
		caseQ              = query.NewCaseQuery()
		skippedBlockSmiths []*model.SkippedBlocksmith
	)
	caseQ.Select(sbs.SkippedBlocksmithQuery.TableName, sbs.SkippedBlocksmithQuery.Fields...)
	caseQ.Where(caseQ.Between("block_height", req.BlockHeightStart, req.BlockHeightEnd))
	caseQ.OrderBy("block_height", model.OrderBy_ASC)

	selectQ, args := caseQ.Build()
	rowCount, _ = sbs.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(selectQ), false, args...)
	if err = rowCount.Scan(&count); err != nil {
		if err != sql.ErrNoRows {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.NotFound, "Record not found")
	}

	rows, err := sbs.QueryExecutor.ExecuteSelect(selectQ, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()
	skippedBlockSmiths, err = sbs.SkippedBlocksmithQuery.BuildModel([]*model.SkippedBlocksmith{}, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &model.GetSkippedBlocksmithsResponse{
		Total:              count,
		SkippedBlocksmiths: skippedBlockSmiths,
	}, nil
}
