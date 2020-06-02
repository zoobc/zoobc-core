package fee

import (
	"time"

	"github.com/zoobc/zoobc-core/common/util"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	FeeScaleServiceInterface interface {
		InsertFeeScale(feeScale *model.FeeScale, dbTx bool) error
		GetLatestFeeScale(feeScale *model.FeeScale) error
		GetCurrentPhase(
			blockTimestamp int64,
			isPostTransaction bool,
		) (model.FeeVotePhase, bool, error)
	}

	FeeScaleService struct {
		lastBlockTimestamp  int64
		lastFeeScale        model.FeeScale
		feeScaleQuery       query.FeeScaleQueryInterface
		mainchainBlockQuery query.BlockQueryInterface
		executor            query.ExecutorInterface
	}
)

func NewFeeScaleService(
	feeScaleQuery query.FeeScaleQueryInterface,
	mainchainBlockQuery query.BlockQueryInterface,
	executor query.ExecutorInterface,
) *FeeScaleService {
	return &FeeScaleService{
		feeScaleQuery:       feeScaleQuery,
		mainchainBlockQuery: mainchainBlockQuery,
		executor:            executor,
	}
}

// InsertFeeScale insert newly agreed feeScale value
func (fss *FeeScaleService) InsertFeeScale(feeScale *model.FeeScale, dbTx bool) error {
	insertQry, args := fss.feeScaleQuery.InsertFeeScale(feeScale)
	if dbTx {
		return fss.executor.ExecuteTransaction(insertQry, args...)
	}
	_, err := fss.executor.ExecuteStatement(insertQry, args...)
	if err != nil {
		return err
	}
	fss.lastFeeScale = *feeScale
	return nil
}

// GetLatestFeeScale return the latest agreed fee-scale value and cached
func (fss *FeeScaleService) GetLatestFeeScale(feeScale *model.FeeScale) error {
	if fss.lastFeeScale.FeeScale != 0 {
		*feeScale = fss.lastFeeScale
		return nil
	}
	getLatestQry := fss.feeScaleQuery.GetLatestFeeScale()
	row, err := fss.executor.ExecuteSelectRow(getLatestQry, false)
	if err != nil {
		return err
	}
	err = fss.feeScaleQuery.Scan(feeScale, row)
	if err != nil {
		return err
	}
	fss.lastFeeScale = *feeScale
	return nil
}

// GetCurrentPhase require 2 parameters the blockTimestamp (when pushBlock) or currentTimestamp (when postTransaction)
// and isPostTransaction parameter when set true will not update the cache, and blockTimestamp need to be filled with
// node's current timestamp instead
func (fss *FeeScaleService) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (model.FeeVotePhase, bool, error) {
	// check if lastBlockstimestamp is 0
	if fss.lastBlockTimestamp == 0 {
		lastBlock, err := util.GetLastBlock(fss.executor, fss.mainchainBlockQuery)
		if err != nil {
			return model.FeeVotePhase_FeeVotePhaseCommmit, false, err
		}
		fss.lastBlockTimestamp = lastBlock.Timestamp
	}
	currentLastBlockTime := time.Unix(blockTimestamp, 0)
	lastRecordedLastBlockTime := time.Unix(fss.lastBlockTimestamp, 0)
	// curr and last
	currYear, currMonth, currDay := currentLastBlockTime.UTC().Date()
	lastYear, lastMonth, _ := lastRecordedLastBlockTime.UTC().Date()
	// cache if not post-transaction checks
	if !isPostTransaction {
		fss.lastBlockTimestamp = blockTimestamp
	}
	// check if can adjust fee -> changes of month or year since last block time
	if (currMonth != lastMonth) || (currYear != lastYear) {
		return model.FeeVotePhase_FeeVotePhaseCommmit, true, nil
	}
	// same month, year and under the commit phase day
	if currDay <= constant.CommitPhaseEndDay {
		return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
	}
	// same month, year, over the commit phase
	return model.FeeVotePhase_FeeVotePhaseReveal, false, nil
}
