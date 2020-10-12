package service

import (
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	BlocksmithServiceInterface interface {
		GetBlocksmithAccountAddress(block *model.Block) ([]byte, error)
		RewardBlocksmithAccountAddresses(
			blocksmithAccountAddresses [][]byte,
			totalReward, blockTimestamp int64,
			height uint32,
		) error
	}
	BlocksmithService struct {
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		AccountLedgerQuery    query.AccountLedgerQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
		Chaintype             chaintype.ChainType
	}
)

func NewBlocksmithService(
	accountBalanceQuery query.AccountBalanceQueryInterface,
	accountLedgerQuery query.AccountLedgerQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	queryExecutor query.ExecutorInterface,
	chaintype chaintype.ChainType,
) *BlocksmithService {
	return &BlocksmithService{
		AccountBalanceQuery:   accountBalanceQuery,
		AccountLedgerQuery:    accountLedgerQuery,
		NodeRegistrationQuery: nodeRegistrationQuery,
		QueryExecutor:         queryExecutor,
		Chaintype:             chaintype,
	}
}

// GetBlocksmithAccountAddress get the address of blocksmith by its public key at the block's height
func (bs *BlocksmithService) GetBlocksmithAccountAddress(block *model.Block) ([]byte, error) {
	var (
		nr []*model.NodeRegistration
	)
	// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
	qry, args := bs.NodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(block.BlocksmithPublicKey, block.Height)
	rows, err := bs.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	nr, err = bs.NodeRegistrationQuery.BuildModel(nr, rows)
	if (err != nil) || len(nr) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "VersionedNodeRegistrationNotFound")
	}
	return nr[0].AccountAddress, nil
}

// RewardBlocksmithAccountAddresses accrue the block total fees + total coinbase to selected list of accounts
func (bs *BlocksmithService) RewardBlocksmithAccountAddresses(
	blocksmithAccountAddresses [][]byte,
	totalReward, blockTimestamp int64,
	height uint32,
) error {
	queries := make([][]interface{}, 0)
	if len(blocksmithAccountAddresses) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NoAccountToBeRewarded")
	}
	blocksmithReward := totalReward / int64(len(blocksmithAccountAddresses))
	for _, blocksmithAccountAddress := range blocksmithAccountAddresses {
		accountBalanceRecipientQ := bs.AccountBalanceQuery.AddAccountBalance(
			blocksmithReward,
			map[string]interface{}{
				"account_address": blocksmithAccountAddress,
				"block_height":    height,
			},
		)
		queries = append(queries, accountBalanceRecipientQ...)

		accountLedgerQ, accountLedgerArgs := bs.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: blocksmithAccountAddress,
			BalanceChange:  blocksmithReward,
			BlockHeight:    height,
			EventType:      model.EventType_EventReward,
			Timestamp:      uint64(blockTimestamp),
		})

		accountLedgerArgs = append([]interface{}{accountLedgerQ}, accountLedgerArgs...)
		queries = append(queries, accountLedgerArgs)
	}
	if err := bs.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}
