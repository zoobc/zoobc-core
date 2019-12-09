package blockchainsync

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
)

type BlockPopper struct {
	BlockService            service.BlockServiceInterface
	MempoolService          service.MempoolServiceInterface
	NodeRegistrationService service.NodeRegistrationServiceInterface
	QueryExecutor           query.ExecutorInterface
	ChainType               chaintype.ChainType
	ActionTypeSwitcher      transaction.TypeActionSwitcher
	KVDB                    kvdb.KVExecutorInterface
	Logger                  *log.Logger
	ReceiptService          service.ReceiptServiceInterface
}

// PopOffToBlock will remove the block in current Chain until commonBlock is reached
func (bp *BlockPopper) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {

	var (
		mempoolsBackupBytes *bytes.Buffer
		mempoolsBackup      []*model.MempoolTransaction
		err                 error
	)
	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	lastBlock, err := bp.BlockService.GetLastBlock()
	if err != nil {
		return []*model.Block{}, err
	}
	minRollbackHeight := getMinRollbackHeight(lastBlock.Height)

	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork process
		bp.Logger.Warn("the node blockchain detects hardfork, please manually delete the database to recover")
		return []*model.Block{}, nil
	}

	_, err = bp.BlockService.GetBlockByID(commonBlock.ID)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	var poppedBlocks []*model.Block
	block := lastBlock

	txs, _ := bp.BlockService.GetTransactionsByBlockID(block.ID)
	block.Transactions = txs

	publishedReceipts, err := bp.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.PublishedReceipts = publishedReceipts

	genesisBlockID := bp.ChainType.GetGenesisBlockID()
	for block.ID != commonBlock.ID && block.ID != genesisBlockID {
		poppedBlocks = append(poppedBlocks, block)

		block, err = bp.BlockService.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
		txs, _ := bp.BlockService.GetTransactionsByBlockID(block.ID)
		block.Transactions = txs

		publishedReceipts, err := bp.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		block.PublishedReceipts = publishedReceipts

	}

	// Backup existing transactions from mempool before rollback
	mempoolsBackup, err = bp.MempoolService.GetMempoolTransactionsWantToBackup(commonBlock.Height)
	if err != nil {
		return nil, err
	}
	bp.Logger.Warnf("mempool tx backup %d in total with block_height %d", len(mempoolsBackup), commonBlock.GetHeight())
	derivedQueries := query.GetDerivedQuery(bp.ChainType)
	err = bp.QueryExecutor.BeginTx()
	if err != nil {
		return []*model.Block{}, err
	}

	for _, dQuery := range derivedQueries {
		queries := dQuery.Rollback(commonBlock.Height)
		err = bp.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			_ = bp.QueryExecutor.RollbackTx()
			return []*model.Block{}, err
		}
	}
	err = bp.QueryExecutor.CommitTx()
	if err != nil {
		return []*model.Block{}, err
	}

	mempoolsBackupBytes = bytes.NewBuffer([]byte{})
	err = bp.QueryExecutor.BeginTx()
	if err != nil {
		return []*model.Block{}, err
	}

	for _, mempool := range mempoolsBackup {
		var (
			tx     *model.Transaction
			txType transaction.TypeAction
		)
		tx, err := transaction.ParseTransactionBytes(mempool.GetTransactionBytes(), true)
		if err != nil {
			return nil, err
		}
		txType, err = bp.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}

		err = txType.UndoApplyUnconfirmed()
		if err != nil {
			return nil, err
		}

		/*
			mempoolsBackupBytes format is
			[...{4}byteSize,{bytesSize}transactionBytes]
		*/
		sizeMempool := uint32(len(mempool.GetTransactionBytes()))
		mempoolsBackupBytes.Write(util.ConvertUint32ToBytes(sizeMempool))
		mempoolsBackupBytes.Write(mempool.GetTransactionBytes())
	}
	err = bp.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}

	if mempoolsBackupBytes.Len() > 0 {
		err = bp.KVDB.Insert(constant.KVDBMempoolsBackup, mempoolsBackupBytes.Bytes(), int(constant.KVDBMempoolsBackupExpiry))
		if err != nil {
			return nil, err
		}
	}
	// remove peer memoization
	bp.NodeRegistrationService.ResetScrambledNodes()
	return poppedBlocks, nil
}
