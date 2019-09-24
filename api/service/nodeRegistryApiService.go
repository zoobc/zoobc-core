package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	NodeRegistryServiceInterface interface {
		GetNodeRegistrations(*model.GetNodeRegistrationsRequest) (*model.GetNodeRegistrationsResponse, error)
		GetNodeRegistration(*model.GetNodeRegistrationRequest) (*model.GetNodeRegistrationResponse, error)
	}

	NodeRegistryService struct {
		Query query.ExecutorInterface
	}
)

func NewNodeRegistryService(queryExecutor query.ExecutorInterface) *NodeRegistryService {
	return &NodeRegistryService{
		Query: queryExecutor,
	}
}

func (ns NodeRegistryService) GetNodeRegistrations(params *model.GetNodeRegistrationsRequest,
) (*model.GetNodeRegistrationsResponse, error) {
	var (
		err               error
		rows              *sql.Rows
		selectQuery       string
		args              []interface{}
		totalRecords      uint64
		nodeRegistrations []*model.NodeRegistration
	)

	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	caseQuery := query.NewCaseQuery()
	maxHeight := params.GetMaxRegistrationHeight()
	page := params.GetPagination()

	caseQuery.Select(nodeRegistrationQuery.TableName, nodeRegistrationQuery.Fields[1:]...)
	caseQuery.Where(caseQuery.Equal("latest", 1))
	caseQuery.And(caseQuery.Equal("queued", params.GetQueued()))
	caseQuery.And(caseQuery.GreaterEqual("registration_height", params.GetMinRegistrationHeight()))
	if maxHeight > 0 {
		caseQuery.And(caseQuery.LessEqual("registration_height", maxHeight))
	}

	// count first
	selectQuery, args = caseQuery.Build()
	countQuery := query.GetTotalRecordOfSelect(selectQuery)
	rows, err = ns.Query.ExecuteSelect(countQuery, false, args...)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.DBErr,
			err.Error(),
		)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&totalRecords,
		)
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.DBErr,
				err.Error(),
			)
		}
	}

	if page.GetOrderField() == "" {
		caseQuery.OrderBy("registration_height", page.GetOrderBy())
	} else {
		caseQuery.OrderBy(page.GetOrderField(), page.GetOrderBy())
	}
	caseQuery.Paginate(page.GetLimit(), page.GetPage())
	selectQuery, args = caseQuery.Build()

	// Get list of node registry
	rows, err = ns.Query.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.DBErr,
			err.Error(),
		)
	}
	for rows.Next() {
		var nr model.NodeRegistration
		_ = rows.Scan(
			&nr.NodePublicKey,
			&nr.AccountAddress,
			&nr.RegistrationHeight,
			&nr.NodeAddress,
			&nr.LockedBalance,
			&nr.Queued,
			&nr.Latest,
			&nr.Height)
		nodeRegistrations = append(nodeRegistrations, &nr)
	}

	nodeRegistrations = nodeRegistrationQuery.BuildModel(nodeRegistrations, rows)
	return &model.GetNodeRegistrationsResponse{
		Total:             totalRecords,
		NodeRegistrations: nodeRegistrations,
	}, nil
}

func (ns NodeRegistryService) GetNodeRegistration(params *model.GetNodeRegistrationRequest,
) (*model.GetNodeRegistrationResponse, error) {
	var (
		row              *sql.Row
		err              error
		nodeRegistration model.NodeRegistration
	)
	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	caseQuery := query.NewCaseQuery()

	caseQuery.Select(nodeRegistrationQuery.TableName, nodeRegistrationQuery.Fields[1:]...)
	if len(params.GetNodePublicKey()) != 0 {
		caseQuery.And(caseQuery.Equal("node_public_key", params.GetNodePublicKey()))
	}
	if params.GetAccountAddress() != "" {
		caseQuery.And(caseQuery.Equal("account_address", params.GetAccountAddress()))
	}
	if params.GetRegistrationHeight() != 0 {
		caseQuery.And(caseQuery.Equal("registration_height", params.GetRegistrationHeight()))
	}
	if params.GetNodeAddress() != "" {
		caseQuery.And(caseQuery.Equal("node_address", params.GetNodeAddress()))
	}
	caseQuery.And(caseQuery.Equal("latest", 1))
	caseQuery.OrderBy("registration_height", model.OrderBy_ASC)
	caseQuery.Limit(1)
	selectQuery, args := caseQuery.Build()

	row = ns.Query.ExecuteSelectRow(selectQuery, args...)
	err = row.Scan(
		&nodeRegistration.NodePublicKey,
		&nodeRegistration.AccountAddress,
		&nodeRegistration.RegistrationHeight,
		&nodeRegistration.NodeAddress,
		&nodeRegistration.LockedBalance,
		&nodeRegistration.Queued,
		&nodeRegistration.Latest,
		&nodeRegistration.Height,
	)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.DBErr,
			err.Error(),
		)
	}

	return &model.GetNodeRegistrationResponse{
		NodeRegistration: &nodeRegistration,
	}, nil
}
