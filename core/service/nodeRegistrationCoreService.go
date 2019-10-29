package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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
		GetNodeAdmittanceCycle() uint32
	}

	// NodeRegistrationService mockable service methods
	NodeRegistrationService struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeAdmittanceCycle     uint32
		Logger                  *log.Logger
	}
)

func NewNodeRegistrationService(
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	logger *log.Logger,
) *NodeRegistrationService {
	return &NodeRegistrationService{
		QueryExecutor:           queryExecutor,
		AccountBalanceQuery:     accountBalanceQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		ParticipationScoreQuery: participationScoreQuery,
		NodeAdmittanceCycle:     constant.NodeAdmittanceCycle,
		Logger:                  logger,
	}
}

// SelectNodesToBeAdmitted Select n (=limit) queued nodes with the highest locked balance
func (nrs *NodeRegistrationService) SelectNodesToBeAdmitted(limit uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsByHighestLockedBalance(limit, model.NodeRegistrationState_NodeQueued)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}
	return nodeRegistrations, nil
}

// SelectNodesToBeExpelled Select n (=limit) registered nodes with participation score = 0
func (nrs *NodeRegistrationService) SelectNodesToBeExpelled() ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistrationsWithZeroScore(model.NodeRegistrationState_NodeRegistered)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}
	return nodeRegistrations, nil
}

func (nrs *NodeRegistrationService) GetNodeRegistryAtHeight(height uint32) ([]*model.NodeRegistration, error) {
	qry := nrs.NodeRegistrationQuery.GetNodeRegistryAtHeight(height)
	rows, err := nrs.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if (err != nil) || len(nodeRegistrations) == 0 {
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

	nodeRegistrations, err := nrs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if (err != nil) || len(nodeRegistrations) == 0 {
		return nil, blocker.NewBlocker(blocker.AppErr, "NoRegisteredNodesFound")
	}

	return nodeRegistrations[0], nil
}

// AdmitNodes update given node registrations' registrationStatus field to NodeRegistrationState_NodeRegistered (=0)
// and set default participation score to it
func (nrs *NodeRegistrationService) AdmitNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	// prepare all node registrations to be updated (set registrationStatus to NodeRegistrationState_NodeRegistered and new height)
	// and default participation scores to be added
	for _, nodeRegistration := range nodeRegistrations {
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeRegistered)
		nodeRegistration.Height = height
		// update the node registry (set registrationStatus to zero)
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
		if err := nrs.QueryExecutor.ExecuteTransactions(queries); err != nil {
			return err
		}
	}
	return nil
}

// ExpelNode (similar to delete node registration) Increase node's owner account balance by node registration's locked balance, then
// update the node registration by setting registrationStatus field to 3 (deleted) and locked balance to zero
func (nrs *NodeRegistrationService) ExpelNodes(nodeRegistrations []*model.NodeRegistration, height uint32) error {
	for _, nodeRegistration := range nodeRegistrations {
		// update the node registry (set registrationStatus to 1 and lockedbalance to 0)
		nodeRegistration.RegistrationStatus = uint32(model.NodeRegistrationState_NodeDeleted)
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
		if err := nrs.QueryExecutor.ExecuteTransactions(queries); err != nil {
			return err
		}
	}
	return nil
}

// GetNodeAdmittanceCycle get the offset, in number of blocks, when we accept and expel nodes from registry
func (nrs *NodeRegistrationService) GetNodeAdmittanceCycle() uint32 {
	if nrs.NodeAdmittanceCycle == 0 {
		return constant.NodeAdmittanceCycle
	}
	return nrs.NodeAdmittanceCycle
}
