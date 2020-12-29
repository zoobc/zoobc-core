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
