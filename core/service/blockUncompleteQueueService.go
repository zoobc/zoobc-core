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
	BlockUncompleteQueueServiceInterface interface {
		GetBlockQueue(blockID int64) *model.Block
		AddBlockQueue(block *model.Block)
		SetTransactionsRequired(blockIDs int64, requiredTxIDs TransactionIDsMap)
		AddTransaction(transaction *model.Transaction) []*model.Block
		RequestBlockTransactions(txIds TransactionIDsMap)
		PruneTimeoutBlockQueue()
	}

	// BlockUncompleteQueueService reperesent a list of blocks while waiting their transaction
	BlockUncompleteQueueService struct {
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

func NewBlockUncompleteQueueService(
	ct chaintype.ChainType,
	obsr *observer.Observer,
) BlockUncompleteQueueServiceInterface {
	return &BlockUncompleteQueueService{
		BlocksQueue:                   make(map[int64]*BlockWithMetaData),
		BlockRequiringTransactionsMap: make(map[int64]TransactionIDsMap),
		TransactionsRequiredMap:       make(map[int64]BlockIDsMap),
		Chaintype:                     ct,
		Observer:                      obsr,
	}
}

// GetBlockQueue return a block based on block ID
func (buqs *BlockUncompleteQueueService) GetBlockQueue(blockID int64) *model.Block {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	if buqs.BlocksQueue[blockID] == nil {
		return nil
	}
	var block = buqs.BlocksQueue[blockID].Block
	return block
}

// AddBlockQueue add new block into block queue list
func (buqs *BlockUncompleteQueueService) AddBlockQueue(block *model.Block) {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	buqs.BlocksQueue[block.ID] = &BlockWithMetaData{
		Block:     block,
		Timestamp: time.Now().Unix(),
	}
}

// SetTransactionsRequired setup map of  block with required transactions and map of transaction required by block
func (buqs *BlockUncompleteQueueService) SetTransactionsRequired(blockIDs int64, requiredTxIDs TransactionIDsMap) {
	buqs.BlockQueueLock.Lock()
	defer buqs.BlockQueueLock.Unlock()
	buqs.BlockRequiringTransactionsMap[blockIDs] = requiredTxIDs
	for txID := range requiredTxIDs {
		// save transaction ID when transaction not found
		if buqs.TransactionsRequiredMap[txID] == nil {
			buqs.TransactionsRequiredMap[txID] = make(BlockIDsMap)
		}
		buqs.TransactionsRequiredMap[txID][blockIDs] = true
	}
}

// RequestBlockTransactions request transactons to the peers
func (buqs *BlockUncompleteQueueService) RequestBlockTransactions(txIds TransactionIDsMap) {
	// TODO: chunks requested transaction
	buqs.Observer.Notify(observer.BlockRequestTransactions, txIds, buqs.Chaintype)
}

// AddTransaction will add validated transaction for queue block and return completed block
func (buqs *BlockUncompleteQueueService) AddTransaction(transaction *model.Transaction) []*model.Block {
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
func (buqs *BlockUncompleteQueueService) PruneTimeoutBlockQueue() {
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
