package service

import (
	"database/sql"
	"fmt"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	TransactionServiceInterface interface {
		DeleteBlocksFrom(id int64) (prevBlock *model.Block, err error)
		CommitTransaction() error
		RollbackTransaction() error
		EndTransaction() error
		RollbackDerivedTables(height uint32) error
		ProcessLater(txs []*model.Transaction) error
		BeginTransaction() error
		IsInTransaction() bool
		GetBlockAtHeight(height uint32) (*model.Block, error)
		ClearCache() error
		ScanFinish() error
		ScheduleScan(height uint32, validate bool) (model.Block, error)
	}

	TransactionService struct {
		Db               *sql.DB
		QueryExecutor    query.ExecutorInterface
		Chaintype        chaintype.ChainType
		TransactionQuery query.TransactionQueryInterface
		BlockService     BlockServiceInterface
		dbTx             *sql.Tx
		isInTransaction  bool
		txToProcessLater []model.Transaction
	}
)

func (dts *TransactionService) DeleteBlocksFrom(id int64) (prevBlock *model.Block, err error) {
	tx, _ := dts.Db.Begin()

	block, err := dts.BlockService.GetBlockByID(id)

	query := fmt.Sprintf(dts.TransactionQuery.DeleteTransactions(id))
	fmt.Printf("transactional delete from: %v", query)
	deleteQuery, err := tx.Prepare(query)
	deleteQuery.Exec()
	tx.Commit()

	prevBlockid := coreUtil.GetBlockIDFromHash(block.PreviousBlockHash)
	prevBlock, err = dts.BlockService.GetBlockByID(prevBlockid)
	return prevBlock, err
}

//to be implemented later
// func (dts *TransactionService) RollbackDerivedTables(height uint32) error {
// 	var err error
// 	// derivedRepositories := models.derivedRepositories.Repositories
// 	for _, repo := range derivedRepositories {
// 		_, err = dts.dbTx.Prepare(repo.RollbackQuery(height))
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (dts *TransactionService) ClearCache() error {
	// TODO:
	// Implement clear cache
	return nil
}

func (dts *TransactionService) CommitTransaction() error {
	return dts.dbTx.Commit()
}

func (dts *TransactionService) RollbackTransaction() error {
	defer dts.EndTransaction()
	dts.dbTx.Rollback()
	return nil
}

func (dts *TransactionService) EndTransaction() {
	dts.isInTransaction = false
}

func (dts *TransactionService) IsInTransaction() bool {
	return dts.isInTransaction
}

func (dts *TransactionService) BeginTransaction() error {
	dts.isInTransaction = true
	tx, _ := dts.Db.Begin()
	dts.dbTx = tx
	return nil
}

func (dts *TransactionService) ScheduleScan(height uint32, validate bool) (model.Block, error) {
	// TODO:
	// implement ScheduleScan using transaction
	tx := dts.dbTx
	query := fmt.Sprintf("UPDATE scan SET rescan = TRUE, height = %v, validate = %v", height, validate)
	tx.Prepare(query)
	return model.Block{}, nil
}

func (dts *TransactionService) ScanFinish() error {
	// TODO:
	// implement ScheduleSScanFinish using trasaction
	tx := dts.dbTx
	query := fmt.Sprintf("UPDATE scan SET rescan = FALSE, height = 0, validate = FALSE")
	tx.Prepare(query)
	return nil
}

func (ts TransactionService) ProcessLater(txs []*model.Transaction) error {
	// TODO: Implement txProcessLater

	for _, tx := range txs {
		ts.txToProcessLater = append(ts.txToProcessLater, *tx)
	}
	return nil
}

func (ts TransactionService) GetBlockAtHeight(height uint32) (*model.Block, error) {
	block, err := ts.BlockService.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	return block, nil
}
