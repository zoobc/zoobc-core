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
	NodeRegistryServiceInterface interface {
		GetNodeRegistrations(*model.GetNodeRegistrationsRequest) (*model.GetNodeRegistrationsResponse, error)
		GetNodeRegistration(*model.GetNodeRegistrationRequest) (*model.GetNodeRegistrationResponse, error)
		GetNodeRegistrationsByNodePublicKeys(*model.GetNodeRegistrationsByNodePublicKeysRequest,
		) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error)
		GetPendingNodeRegistrations(*model.GetPendingNodeRegistrationsRequest) (*model.GetPendingNodeRegistrationsResponse, error)
	}

	NodeRegistryService struct {
		Query                 query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
)

func NewNodeRegistryService(queryExecutor query.ExecutorInterface) *NodeRegistryService {
	return &NodeRegistryService{
		Query: queryExecutor,
	}
}

func (ns NodeRegistryService) GetNodeRegistrations(params *model.GetNodeRegistrationsRequest) (
	*model.GetNodeRegistrationsResponse,
	error,
) {

	var (
		err               error
		rowCount          *sql.Row
		rows2             *sql.Rows
		selectQuery       string
		args              []interface{}
		totalRecords      uint64
		nodeRegistrations []*model.NodeRegistration
	)

	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	caseQuery := query.NewCaseQuery()
	maxHeight := params.GetMaxRegistrationHeight()
	page := params.GetPagination()

	caseQuery.Select(nodeRegistrationQuery.TableName, nodeRegistrationQuery.Fields...)
	caseQuery.Where(caseQuery.Equal("latest", 1))

	var statuses []interface{}
	for _, s := range params.GetRegistrationStatuses() {
		statuses = append(statuses, s)
	}
	if len(statuses) > 0 {
		caseQuery.Where(caseQuery.In("registration_status", statuses...))
	}
	caseQuery.And(caseQuery.GreaterEqual("registration_height", params.GetMinRegistrationHeight()))
	if maxHeight > 0 {
		caseQuery.And(caseQuery.LessEqual("registration_height", maxHeight))
	}

	// count first
	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	rowCount, err = ns.Query.ExecuteSelectRow(countQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = rowCount.Scan(
		&totalRecords,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if page.GetOrderField() == "" {
		caseQuery.OrderBy("registration_height", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())
	selectQuery, args = caseQuery.Build()

	// Get list of node registry
	rows2, err = ns.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows2.Close()

	nodeRegistrations, err = nodeRegistrationQuery.BuildModel(nodeRegistrations, rows2)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetNodeRegistrationsResponse{
		Total:             totalRecords,
		NodeRegistrations: nodeRegistrations,
	}, nil
}

func (ns NodeRegistryService) GetNodeRegistrationsByNodePublicKeys(params *model.GetNodeRegistrationsByNodePublicKeysRequest,
) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error) {

	var (
		err                 error
		rows                *sql.Rows
		selectQuery         string
		args                []interface{}
		nodeRegistrations   []*model.NodeRegistration
		publicKeyInterfaces []interface{}
	)

	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	caseQuery := query.NewCaseQuery()

	caseQuery.Select(nodeRegistrationQuery.TableName, nodeRegistrationQuery.Fields...)
	for _, npk := range params.NodePublicKeys {
		publicKeyInterfaces = append(publicKeyInterfaces, npk)
	}
	if len(publicKeyInterfaces) > 0 {
		caseQuery.Where(caseQuery.In("node_public_key", publicKeyInterfaces...))
	}
	caseQuery.And(caseQuery.Equal("latest", 1))
	caseQuery.OrderBy("height", model.OrderBy_DESC)

	selectQuery, args = caseQuery.Build()

	// Get list of node registry
	rows, err = ns.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	nodeRegistrations, err = nodeRegistrationQuery.BuildModel(nodeRegistrations, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetNodeRegistrationsByNodePublicKeysResponse{
		NodeRegistrations: nodeRegistrations,
	}, nil
}

func (ns NodeRegistryService) GetNodeRegistration(
	params *model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {

	var (
		row              *sql.Row
		err              error
		nodeRegistration model.NodeRegistration
	)

	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	caseQuery := query.NewCaseQuery()

	caseQuery.Select(nodeRegistrationQuery.TableName, nodeRegistrationQuery.Fields...)
	if len(params.GetNodePublicKey()) != 0 {
		caseQuery.And(caseQuery.Equal("node_public_key", params.GetNodePublicKey()))
	}
	if params.GetAccountAddress() != nil {
		caseQuery.And(caseQuery.Equal("account_address", params.GetAccountAddress()))
	}
	if params.GetRegistrationHeight() != 0 {
		caseQuery.And(caseQuery.Equal("registration_height", params.GetRegistrationHeight()))
	}
	caseQuery.And(caseQuery.Equal("latest", 1))
	caseQuery.OrderBy("height", model.OrderBy_DESC)
	caseQuery.Limit(1)
	selectQuery, args := caseQuery.Build()

	row, _ = ns.Query.ExecuteSelectRow(selectQuery, false, args...)
	err = nodeRegistrationQuery.Scan(&nodeRegistration, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetNodeRegistrationResponse{
		NodeRegistration: &nodeRegistration,
	}, nil
}

func (ns NodeRegistryService) GetPendingNodeRegistrations(
	req *model.GetPendingNodeRegistrationsRequest) (*model.GetPendingNodeRegistrationsResponse, error) {
	var (
		err               error
		rows              *sql.Rows
		args              []interface{}
		nodeRegistrations []*model.NodeRegistration
		limit             = req.Limit
	)
	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	selectQuery := nodeRegistrationQuery.GetPendingNodeRegistrations(limit)

	// Get list of node registry
	rows, err = ns.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	nodeRegistrations, err = nodeRegistrationQuery.BuildModel(nodeRegistrations, rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetPendingNodeRegistrationsResponse{
		NodeRegistrations: nodeRegistrations,
	}, nil
}
