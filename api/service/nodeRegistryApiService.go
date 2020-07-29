package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	NodeRegistryServiceInterface interface {
		GetNodeRegistrations(*model.GetNodeRegistrationsRequest) (*model.GetNodeRegistrationsResponse, error)
		GetNodeRegistration(*model.GetNodeRegistrationRequest) (*model.GetNodeRegistrationResponse, error)
		GetNodeRegistrationsByNodePublicKeys(*model.GetNodeRegistrationsByNodePublicKeysRequest) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error)
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
	caseQuery.And(caseQuery.Equal("registration_status", params.GetRegistrationStatus()))
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

func (ns NodeRegistryService) GetNodeRegistrationsByNodePublicKeys(params *model.GetNodeRegistrationsByNodePublicKeysRequest) (*model.GetNodeRegistrationsByNodePublicKeysResponse, error) {
	rows, err := ns.Query.ExecuteSelect(ns.NodeRegistrationQuery.GetNodeRegistrationsByNodePublicKeys(), false, params.NodePublicKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := ns.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if (err != nil) || len(nodeRegistrations) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
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
	if params.GetAccountAddress() != "" {
		caseQuery.And(caseQuery.Equal("account_address", params.GetAccountAddress()))
	}
	if params.GetRegistrationHeight() != 0 {
		caseQuery.And(caseQuery.Equal("registration_height", params.GetRegistrationHeight()))
	}
	if params.GetNodeAddress() != nil {
		caseQuery.And(caseQuery.Equal("node_address", nodeRegistrationQuery.ExtractNodeAddress(params.GetNodeAddress())))
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
