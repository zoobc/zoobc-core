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
	"sync"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	BlockIncompleteQueueServiceInterface interface {
		GetBlockQueue(blockID int64) *model.Block
		AddBlockQueue(block *model.Block)
		SetTransactionsRequired(blockIDs int64, requiredTxIDs TransactionIDsMap)
		AddTransaction(transaction *model.Transaction) []*model.Block
		RequestBlockTransactions(txIds []int64, blockID int64, peer *model.Peer)
		PruneTimeoutBlockQueue()
	}

	// BlockIncompleteQueueService reperesent a list of blocks while waiting their transaction
	BlockIncompleteQueueService struct {
		// map of block ID with the blocks that have been received but waiting transactions to be completed
		BlocksQueue map[int64]*BlockWithMetaData
		// map of blockID with an array of transactionIds it requires
		BlockRequiringTransactionsMap map[int64]TransactionIDsMap
		// map of transactionIds with blockIds that requires them
		TransactionsRequiredMap map[int64]BlockIDsMap
		Chaintype               chaintype.ChainType
		BlockQueueLock          sync.Mutex
		Observer                *observer.Observer
	}

	// TransactionIDs reperesent key value of transaction ID with it's index potition in the block
	TransactionIDsMap map[int64]int
	BlockIDsMap       map[int64]bool
	// BlockWithMetaData is incoming block with some information while waiting transaction
	BlockWithMetaData struct {
		Block     *model.Block
		Timestamp int64
	}
)

func NewBlockIncompleteQueueService(
	ct chaintype.ChainType,
	obsr *observer.Observer,
) BlockIncompleteQueueServiceInterface {
	return &BlockIncompleteQueueService{
		BlocksQueue:                   make(map[int64]*BlockWithMetaData),
		BlockRequiringTransactionsMap: make(map[int64]TransactionIDsMap),
		TransactionsRequiredMap:       make(map[int64]BlockIDsMap),
		Chaintype:                     ct,
		Observer:                      obsr,
	}
}

// GetBlockQueue return a block based on block ID
func (buqs *BlockIncompleteQueueService) GetBlockQueue(blockID int64) *model.Block {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	if buqs.BlocksQueue[blockID] == nil {
		return nil
	}
	var block = buqs.BlocksQueue[blockID].Block
	return block
}

// AddBlockQueue add new block into block queue list
func (buqs *BlockIncompleteQueueService) AddBlockQueue(block *model.Block) {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	buqs.BlocksQueue[block.ID] = &BlockWithMetaData{
		Block:     block,
		Timestamp: time.Now().Unix(),
	}
}

// SetTransactionsRequired setup map of  block with required transactions and map of transaction required by block
func (buqs *BlockIncompleteQueueService) SetTransactionsRequired(blockIDs int64, requiredTxIDs TransactionIDsMap) {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	buqs.BlockRequiringTransactionsMap[blockIDs] = requiredTxIDs
	for txID := range requiredTxIDs {
		if buqs.TransactionsRequiredMap[txID] == nil {
			buqs.TransactionsRequiredMap[txID] = make(BlockIDsMap)
		}
		buqs.TransactionsRequiredMap[txID][blockIDs] = true
	}
}

// RequestBlockTransactions request transactons to the peers
func (buqs *BlockIncompleteQueueService) RequestBlockTransactions(txIds []int64, blockID int64, peer *model.Peer) {
	// TODO: chunks requested transaction
	buqs.Observer.Notify(observer.BlockRequestTransactions, txIds, blockID, buqs.Chaintype, peer)
}

// AddTransaction will add validated transaction for queue block and return completed block
func (buqs *BlockIncompleteQueueService) AddTransaction(transaction *model.Transaction) []*model.Block {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()

	var completedBlocks []*model.Block
	for blockID := range buqs.TransactionsRequiredMap[transaction.GetID()] {
		// check if waiting block is exist
		if buqs.BlocksQueue[blockID] == nil || buqs.BlockRequiringTransactionsMap[blockID] == nil {
			continue
		}
		var (
			txs     = buqs.BlocksQueue[blockID].Block.GetTransactions()
			txIndex = buqs.BlockRequiringTransactionsMap[blockID][transaction.GetID()]
		)
		// joining new transaction into list of transactions
		if len(txs) < txIndex {
			continue
		}
		txs[txIndex] = transaction
		buqs.BlocksQueue[blockID].Block.Transactions = txs
		delete(buqs.BlockRequiringTransactionsMap[blockID], transaction.GetID())
		// process block when all transactions are completed
		if len(buqs.BlockRequiringTransactionsMap[blockID]) == 0 {
			completedBlocks = append(completedBlocks, buqs.BlocksQueue[blockID].Block)
			// remove waited block and list of transaction ID map when block already completed their transaction
			delete(buqs.BlocksQueue, blockID)
			delete(buqs.BlockRequiringTransactionsMap, blockID)
		}
	}
	// removing required transaction ID when it's not needed by any block
	delete(buqs.TransactionsRequiredMap, transaction.GetID())
	return completedBlocks
}

// PruneTimeoutBlockQueue used as scheduler remove block when already expired
func (buqs *BlockIncompleteQueueService) PruneTimeoutBlockQueue() {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	for blockID, blockWithMetaData := range buqs.BlocksQueue {
		// check waiting time block
		if blockWithMetaData.Timestamp <= time.Now().Unix()-constant.TimeOutBlockWaitingTransactions {
			for _, transactionID := range blockWithMetaData.Block.GetTransactionIDs() {
				delete(buqs.TransactionsRequiredMap[transactionID], blockID)
			}
			delete(buqs.BlocksQueue, blockID)
			delete(buqs.BlockRequiringTransactionsMap, blockID)
		}
	}
}
