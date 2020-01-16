package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	BlockSpineSnapshotServiceInterface interface {
		GetMegablockFromSpineHeight(spineHeight uint32) (*model.Megablock, error)
		GetLastMegablock() (*model.Megablock, error)
		CreateMegablock(snapshotHash []byte, mainHeight uint32) (*model.Megablock, error)
		GetNextSnapshotHeight(mainHeight uint32) uint32
	}

	BlockSpineSnapshotService struct {
		QueryExecutor   query.ExecutorInterface
		MegablockQuery  query.MegablockQueryInterface
		SpineBlockQuery query.BlockQueryInterface
		MainBlockQuery  query.BlockQueryInterface
		Logger          *log.Logger
		// this is mostly for mocking
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
)

func NewSnapshotService(
	queryExecutor query.ExecutorInterface,
	mainBlockQuery query.BlockQueryInterface,
	spineBlockQuery query.BlockQueryInterface,
	megablockQuery query.MegablockQueryInterface,
	logger *log.Logger,
) *BlockSpineSnapshotService {
	return &BlockSpineSnapshotService{
		QueryExecutor:             queryExecutor,
		MegablockQuery:            megablockQuery,
		SpineBlockQuery:           spineBlockQuery,
		MainBlockQuery:            mainBlockQuery,
		Spinechain:                &chaintype.SpineChain{},
		Mainchain:                 &chaintype.MainChain{},
		SnapshotInterval:          constant.SnapshotInterval,
		SnapshotGenerationTimeout: constant.SnapshotGenerationTimeout,
		Logger:                    logger,
	}
}

// GetMegablockFromSpineHeight retrieve a megablock for a given spine height
func (ss *BlockSpineSnapshotService) GetMegablockFromSpineHeight(spineHeight uint32) (*model.Megablock, error) {
	var (
		megablock model.Megablock
	)
	qry := ss.MegablockQuery.GetMegablocksByBlockHeight(spineHeight, ss.Spinechain)
	row, err := ss.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = ss.MegablockQuery.Scan(&megablock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return nil, nil
}

// GetMegablockFromSpineHeight retrieve a megablock for a given spine height
func (ss *BlockSpineSnapshotService) GetLastMegablock() (*model.Megablock, error) {
	var (
		megablock model.Megablock
	)
	qry := ss.MegablockQuery.GetLastMegablock()
	row, err := ss.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = ss.MegablockQuery.Scan(&megablock, row)
	if err != nil {
		if blockErr, ok := err.(blocker.Blocker); ok && blockErr.Type != blocker.DBRowNotFound {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		// return nil if no megablocks are found
		return nil, nil
	}
	return &megablock, nil
}

// CreateMegablock persist a new megablock
// snapshotHash: hash of the full snapshot content
// mainHeight: mainchain height at which the snapshot has started (note: this is not the captured snapshot's height,
// which should be = mainHeight - minRollbackHeight)
func (ss *BlockSpineSnapshotService) CreateMegablock(snapshotHash []byte, mainHeight uint32) (*model.Megablock,
	error) {
	var (
		lastMainBlock         model.Block
		firstValidSpineHeight uint32
	)
	row, err := ss.QueryExecutor.ExecuteSelectRow(ss.MainBlockQuery.GetLastBlock(), false)
	if err != nil {
		return nil, err
	}
	err = ss.MainBlockQuery.Scan(&lastMainBlock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// spine blocks have discrete timing,
	// so we can calculate accurately next spine timestamp and give enough time to all nodes to complete their snapshot
	spinechainInterval := ss.Spinechain.GetSmithingPeriod() + ss.Spinechain.GetChainSmithingDelayTime()
	// lastMainBlock.Timestamp is the timestamp at which the snapshot started to be computed
	nextMinimumSpineBlockTimestamp := lastMainBlock.Timestamp + ss.SnapshotGenerationTimeout
	firstValidTime := util.GetNextStep(nextMinimumSpineBlockTimestamp, spinechainInterval)
	firstValidSpineHeight = uint32((firstValidTime - constant.SpinechainGenesisBlockTimestamp) / spinechainInterval)
	megablock := &model.Megablock{
		FullSnapshotHash: snapshotHash,
		MainBlockHeight:  mainHeight,
		SpineBlockHeight: firstValidSpineHeight,
	}
	qry, args := ss.MegablockQuery.InsertMegablock(megablock)
	err = ss.QueryExecutor.ExecuteTransaction(qry, args)
	if err != nil {
		return nil, err
	}
	return megablock, nil
}

// GetNextSnapshotHeight calculate next snapshot (main block) height given an arbitrary main block height
func (ss *BlockSpineSnapshotService) GetNextSnapshotHeight(mainHeight uint32) uint32 {
	// first snapshot cannot be taken before minRollBack height
	if mainHeight < constant.MinRollbackBlocks {
		mainHeight = constant.MinRollbackBlocks
	}
	avgBlockTime := ss.Mainchain.GetSmithingPeriod() + ss.Mainchain.GetChainSmithingDelayTime()
	avgBlockInterval := ss.SnapshotInterval / avgBlockTime
	return uint32(util.GetNextStep(int64(mainHeight), avgBlockInterval))
}
