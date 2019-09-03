package service

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// NodeRegistrationServiceInterface represents interface for NodeRegistrationService
	NodeRegistrationServiceInterface interface {
		SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error)
		AdmitNode(nodeRegistration *model.NodeRegistration, height uint32) error
		KickOutNode(nodeRegistration *model.NodeRegistration, height uint32) error
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor         query.ExecutorInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:         queryExecutor,
		AccountBalanceQuery:   accountBalanceQuery,
		NodeRegistrationQuery: nodeRegistrationQuery,
	}
}

// SelectNodesToBeAdmitted Select n (=limit) nodes with the highest locked balance
// TODO: add check to filter out (either here or in the query) nodes with reputation score = 0
func (nrs *NodeRegistrationService) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(limit)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistrations := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if len(nodeRegistrations) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
	}

	return nodeRegistrations, nil
}

// AdmitNodes update given node registrations' queued field to false
func (nrs *NodeRegistrationService) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	queries := make([][]interface{}, 0)
	if len(nodeRegistrations) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NoNodesToBeAdmitted")
	}
	// prepare all node registrations to be updated (set queued to false and new height)
	for _, nodeRegistration := range nodeRegistrations {
		nodeRegistration.Queued = false
		nodeRegistration.Height = height
		updateNodeQ, updateNodeArg := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
		newQ := []interface{}{
			updateNodeQ, updateNodeArg,
		}
		queries = append(queries, newQ)
	}
	err := nrs.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	if err := nrs.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
	}

	return nil
}

// KickOutNode (similar to delete node registration) Increase node's owner account balance by node registration's locked balance, then
// update the node registration by setting queued field to true and locked balance to zero
func (nrs *NodeRegistrationService) KickOutNode(nodeRegistration *model.NodeRegistration, height uint32) error {
	updateAccountBalanceQ := nrs.AccountBalanceQuery.AddAccountBalance(
		nodeRegistration.LockedBalance,
		map[string]interface{}{
			"account_address": nodeRegistration.AccountAddress,
			"block_height":    height,
		},
	)

	nodeRegistration.Queued = true
	nodeRegistration.LockedBalance = 0
	updateNodeQ, updateNodeArg := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)

	queries := append(append([][]interface{}{}, updateAccountBalanceQ...),
		append([]interface{}{updateNodeQ}, updateNodeArg...),
	)
	err := nrs.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}

	if err := nrs.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
	}

	return nil
}
