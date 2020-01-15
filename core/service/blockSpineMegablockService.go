package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	BlockSpineMegablockServiceInterface interface {
		GetMegablockFromHeight(height uint32) (*model.Megablock, error)
		CreateMegablock(megablock *model.Megablock) error
	}

	BlockSpineMegablockService struct {
		QueryExecutor  query.ExecutorInterface
		MegablockQuery query.MegablockQueryInterface
		Logger         *log.Logger
	}
)

func (mbl *BlockSpineMegablockService) GetMegablockFromHeight(height uint32) (*model.Megablock, error) {
	return nil, nil
}

func (mbl *BlockSpineMegablockService) CreateMegablock(megablock *model.Megablock) error {
	return nil
}
