package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	// NodeRegistrationServiceInterface represents interface for NodeRegistrationService
	NodeRegistrationServiceInterface interface {
		SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error)
		SelectNodesToBeExpelled() ([]*model.NodeRegistration, error)
		GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error)
		GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error)
		AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error
		NodeRegistryListener() observer.Listener
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		// mockable variables
		NodeAdmittanceCycle uint32
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:           queryExecutor,
		AccountBalanceQuery:     accountBalanceQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		ParticipationScoreQuery: participationScoreQuery,
		NodeAdmittanceCycle:     constant.NodeAdmittanceCycle,
	}
}

// SelectNodesToBeAdmitted Select n (=limit) queued nodes with the highest locked balance
func (nrs *NodeRegistrationService) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(limit, true)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistrations := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	return nodeRegistrations, nil
}

func (nrs *NodeRegistrationService) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsWithZeroScore(false)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistrations := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	return nodeRegistrations, nil
}

func (nrs *NodeRegistrationService) GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(height)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
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

func (nrs *NodeRegistrationService) GetNodeRegistrationByNodePublicKey(nodePublicKey []byte) (*model.NodeRegistration, error) {
	rows, err := nrs.QueryExecutor.ExecuteSelect(nrs.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey(), false, nodePublicKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistrations := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if len(nodeRegistrations) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
	}

	return nodeRegistrations[0], nil
}

// AdmitNodes update given node registrations' queued field to false and set default participation score to it
func (nrs *NodeRegistrationService) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	err := nrs.QueryExecutor.BeginTx()
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// prepare all node registrations to be updated (set queued to false and new height) and default participation scores to be added
	for _, nodeRegistration := range nodeRegistrations {
		nodeRegistration.Queued = false
		nodeRegistration.Height = height
		// update the node registry (set queued to zero)
		queries := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
		ps := &model.ParticipationScore{
			NodeID: nodeRegistration.NodeID,
			Score:  constant.MaxParticipationScore / 10,
			Latest: true,
			Height: height,
		}
		// add default participation score to the node
		insertParticipationScoreQ, insertParticipationScoreArg := nrs.ParticipationScoreQuery.InsertParticipationScore(ps)
		queries = append(queries,
			append([]interface{}{insertParticipationScoreQ}, insertParticipationScoreArg...),
		)
		err = nrs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			// no need to rollback here, since we already do it in ExecuteTransactions
			return err
		}
	}

	if err := nrs.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
	}

	return nil
}

// ExpelNode (similar to delete node registration) Increase node's owner account balance by node registration's locked balance, then
// update the node registration by setting queued field to true and locked balance to zero
func (nrs *NodeRegistrationService) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	err := nrs.QueryExecutor.BeginTx()
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	for _, nodeRegistration := range nodeRegistrations {
		// update the node registry (set queued to 1 and lockedbalance to 0)
		nodeRegistration.Queued = true
		nodeRegistration.LockedBalance = 0
		nodeQueries := nrs.NodeRegistrationQuery.UpdateNodeRegistration(nodeRegistration)
		// return lockedbalance to the node's owner account
		updateAccountBalanceQ := nrs.AccountBalanceQuery.AddAccountBalance(
			nodeRegistration.LockedBalance,
			map[string]interface{}{
				"account_address": nodeRegistration.AccountAddress,
				"block_height":    height,
			},
		)

		queries := append(updateAccountBalanceQ, nodeQueries...)
		err := nrs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			return err
		}
	}
	if err := nrs.QueryExecutor.CommitTx(); err != nil {
		return blocker.NewBlocker(blocker.DBErr, "TxNotCommitted")
	}

	return nil
}

// NodeRegistryListener handle node admission/expulsion after a block is pushed, at regular interval
func (nrs *NodeRegistrationService) NodeRegistryListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			pushedBlock := block.(*model.Block)
			if pushedBlock.Height > 0 && pushedBlock.Height%nrs.NodeAdmittanceCycle != 0 {
				return
			}
			nodeRegistrations, err := nrs.SelectNodesToBeAdmitted(constant.MaxNodeAdmittancePerCycle)
			if err != nil {
				log.Errorf("Can't get list of nodes from node registry: %s", err)
				return
			}
			if len(nodeRegistrations) > 0 {
				err = nrs.AdmitNodes(nodeRegistrations, pushedBlock.Height)
				if err != nil {
					log.Errorf("Can't admit nodes to registry: %s", err)
					return
				}
			}
			// expel nodes with zero score from node registry
			nodeRegistrations, err = nrs.SelectNodesToBeExpelled()
			if err != nil {
				log.Errorf("Can't get list of nodes from node registry: %s", err)
				return
			}
			if len(nodeRegistrations) > 0 {
				err = nrs.ExpelNodes(nodeRegistrations, pushedBlock.Height)
				if err != nil {
					log.Errorf("Can't expel nodes from registry: %s", err)
					return
				}
			}
		},
	}
}
