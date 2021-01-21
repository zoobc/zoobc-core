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

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	MempoolTransactionServiceInterface interface {
		GetMempoolTransaction(
			chainType chaintype.ChainType,
			params *model.GetMempoolTransactionRequest,
		) (*model.GetMempoolTransactionResponse, error)
		GetMempoolTransactions(
			chainType chaintype.ChainType,
			params *model.GetMempoolTransactionsRequest,
		) (*model.GetMempoolTransactionsResponse, error)
	}
	MempoolTransactionService struct {
		Query query.ExecutorInterface
	}
)

func NewMempoolTransactionsService(
	queryExecutor query.ExecutorInterface,
) *MempoolTransactionService {
	return &MempoolTransactionService{
		Query: queryExecutor,
	}
}

func (ut *MempoolTransactionService) GetMempoolTransaction(
	chainType chaintype.ChainType,
	params *model.GetMempoolTransactionRequest,
) (*model.GetMempoolTransactionResponse, error) {
	var (
		err error
		row *sql.Row
		tx  model.MempoolTransaction
	)

	txQuery := query.NewMempoolQuery(chainType)
	row, _ = ut.Query.ExecuteSelectRow(txQuery.GetMempoolTransaction(), false, params.GetID())
	if row == nil {
		return nil, status.Error(codes.NotFound, "transaction not found in mempool")
	}

	err = txQuery.Scan(&tx, row)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(tx.GetTransactionBytes()) == 0 {
		return nil, status.Error(codes.NotFound, "tx byte is empty")
	}
	return &model.GetMempoolTransactionResponse{
		Transaction: &tx,
	}, nil
}

func (ut *MempoolTransactionService) GetMempoolTransactions(
	chainType chaintype.ChainType,
	params *model.GetMempoolTransactionsRequest,
) (*model.GetMempoolTransactionsResponse, error) {
	var (
		err                     error
		count                   uint64
		selectQuery, countQuery string
		rowCount                *sql.Row
		rows2                   *sql.Rows
		txs                     []*model.MempoolTransaction
		response                *model.GetMempoolTransactionsResponse
		args                    []interface{}
	)

	txQuery := query.NewMempoolQuery(chainType)
	caseQuery := query.CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}

	caseQuery.Select(txQuery.TableName, txQuery.Fields...)

	timestampStart := params.GetTimestampStart()
	timestampEnd := params.GetTimestampEnd()
	if timestampStart > 0 {
		caseQuery.Where(caseQuery.Between("arrival_timestamp", timestampStart, timestampEnd))
	}

	address := params.GetAddress()
	if address != nil {
		caseQuery.And(caseQuery.Equal("sender_account_address", address)).
			Or(caseQuery.Equal("recipient_account_address", address))

	}

	// count first
	selectQuery, args = caseQuery.Build()
	countQuery = query.GetTotalRecordOfSelect(selectQuery)

	rowCount, err = ut.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = rowCount.Scan(&count)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// select records
	page := params.GetPagination()
	if page.GetOrderField() == "" {
		caseQuery.OrderBy("arrival_timestamp", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())

	selectQuery, args = caseQuery.Build()
	rows2, err = ut.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows2.Close()

	txs, err = txQuery.BuildModel(txs, rows2)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response = &model.GetMempoolTransactionsResponse{
		MempoolTransactions: txs,
		Total:               count,
	}
	return response, nil
}
