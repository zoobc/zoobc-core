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
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// EscrowTransactionServiceInterface interface that contain methods of escrow transaction
	EscrowTransactionServiceInterface interface {
		GetEscrowTransactions(request *model.GetEscrowTransactionsRequest) (*model.GetEscrowTransactionsResponse, error)
		GetEscrowTransaction(request *model.GetEscrowTransactionRequest) (*model.Escrow, error)
	}
	// EscrowTransactionService struct that contain fields that needed
	escrowTransactionService struct {
		Query query.ExecutorInterface
	}
)

// NewEscrowTransactionService will create EscrowTransactionServiceInterface instance
func NewEscrowTransactionService(
	query query.ExecutorInterface,
) EscrowTransactionServiceInterface {
	return &escrowTransactionService{
		Query: query,
	}
}

// GetEscrowTransactions to get escrow transactions list
func (es *escrowTransactionService) GetEscrowTransactions(
	params *model.GetEscrowTransactionsRequest,
) (*model.GetEscrowTransactionsResponse, error) {
	var (
		escrowQuery = query.NewEscrowTransactionQuery()
		countQuery  string
		escrows     []*model.Escrow
		rows        *sql.Rows
		count       int64
		row         *sql.Row
		err         error
	)

	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}

	caseQuery.Select(escrowQuery.TableName, escrowQuery.Fields...)
	if params.GetApproverAddress() != nil {
		caseQuery.Where(caseQuery.Equal("approver_address", params.GetApproverAddress()))
	}

	if params.GetSenderAddress() != nil {
		caseQuery.Where(caseQuery.Equal("sender_address", params.GetSenderAddress()))
	}
	if params.GetRecipientAddress() != nil {
		caseQuery.Or(caseQuery.Equal("recipient_address", params.GetRecipientAddress()))
	}

	if len(params.GetStatuses()) > 0 {
		var statuses []interface{}
		for _, v := range params.GetStatuses() {
			statuses = append(statuses, int32(v))
		}
		caseQuery.And(caseQuery.In("status", statuses...))
	}
	if params.GetID() != 0 {
		caseQuery.And(caseQuery.Equal("id", params.GetID()))
	}
	caseQuery.And(caseQuery.Equal("latest", params.GetLatest()))

	blockHeightStart := params.GetBlockHeightStart()
	blockHeightEnd := params.GetBlockHeightEnd()
	if blockHeightStart > 0 {
		caseQuery.Where(caseQuery.Between("block_height", blockHeightStart, blockHeightEnd))
	}

	// count first
	selectQuery, args := caseQuery.Build()
	countQuery = query.GetTotalRecordOfSelect(selectQuery)

	row, err = es.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = row.Scan(&count)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// select records
	page := params.GetPagination()
	if page.GetOrderField() == "" {
		caseQuery.OrderBy("id", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())

	escrowQ, escrowArgs := caseQuery.Build()
	rows, err = es.Query.ExecuteSelect(escrowQ, false, escrowArgs...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	escrows, err = escrowQuery.BuildModels(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetEscrowTransactionsResponse{
		Total:   uint64(count),
		Escrows: escrows,
	}, nil
}

// GetEscrowTransaction to get escrow by id and status
func (es *escrowTransactionService) GetEscrowTransaction(
	request *model.GetEscrowTransactionRequest,
) (*model.Escrow, error) {
	var (
		escrowQuery = query.NewEscrowTransactionQuery()
		escrow      model.Escrow
		row         *sql.Row
		err         error
	)

	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}

	caseQuery.Select(escrowQuery.TableName, escrowQuery.Fields...)
	caseQuery.Where(caseQuery.Equal("id", request.GetID()))
	caseQuery.Where(caseQuery.Equal("latest", 1))

	qStr, qArgs := caseQuery.Build()

	row, _ = es.Query.ExecuteSelectRow(qStr, false, qArgs...)
	err = escrowQuery.Scan(&escrow, row)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &escrow, nil
}
