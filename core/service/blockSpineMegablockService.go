package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	BlockSpineMegablockServiceInterface interface {
		GetMegablockFromSpineHeight(height uint32) (*model.Megablock, error)
		CreateMegablock(megablock *model.Megablock) error
	}

	BlockSpineMegablockService struct {
		QueryExecutor  query.ExecutorInterface
		MegablockQuery query.MegablockQueryInterface
		Logger         *log.Logger
	}
)

// GetMegablockFromSpineHeight retrieve a megablock from its
func (mbl *BlockSpineMegablockService) GetMegablockFromSpineHeight(height uint32) (*model.Megablock, error) {
	var (
		megablock model.Megablock
	)
	qry := mbl.MegablockQuery.GetMegablocksByBlockHeight(height, &chaintype.SpineChain{})
	row, err := mbl.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = mbl.MegablockQuery.Scan(&megablock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil, nil
}

// CreateMegablock persist a new megablock
func (mbl *BlockSpineMegablockService) CreateMegablock(megablock *model.Megablock) error {
	qry, args := mbl.MegablockQuery.InsertMegablock(megablock)
	err := mbl.QueryExecutor.ExecuteTransaction(qry, args)
	if err != nil {
		return err
	}
	return nil
}
