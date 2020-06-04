package fee

import (
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	FeeScaleServiceInterface interface {
		InsertFeeScale(feeScale *model.FeeScale) error
		GetLatestFeeScale(feeScale *model.FeeScale) error
		GetCurrentPhase(
			blockTimestamp int64,
			isPostTransaction bool,
		) (phase model.FeeVotePhase, canAdjust bool, err error)
		IsInPhasePeriod(timestamp int64) error
	}

	FeeScaleService struct {
		lastBlockTimestamp  int64
		lastFeeScale        model.FeeScale
		FeeScaleQuery       query.FeeScaleQueryInterface
		MainchainBlockQuery query.BlockQueryInterface
		Executor            query.ExecutorInterface
	}
)

func NewFeeScaleService(
	feeScaleQuery query.FeeScaleQueryInterface,
	mainchainBlockQuery query.BlockQueryInterface,
	executor query.ExecutorInterface,
) *FeeScaleService {
	return &FeeScaleService{
		FeeScaleQuery:       feeScaleQuery,
		MainchainBlockQuery: mainchainBlockQuery,
		Executor:            executor,
	}
}

// InsertFeeScale insert newly agreed feeScale value must be called in database transaction
func (fss *FeeScaleService) InsertFeeScale(feeScale *model.FeeScale) error {
	insertQueries := fss.FeeScaleQuery.InsertFeeScale(feeScale)
	err := fss.Executor.ExecuteTransactions(insertQueries)
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
	getLatestQry := fss.FeeScaleQuery.GetLatestFeeScale()
	row, err := fss.Executor.ExecuteSelectRow(getLatestQry, false)
	if err != nil {
		return err
	}
	err = fss.FeeScaleQuery.Scan(feeScale, row)
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
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	// check if lastBlockstimestamp is 0
	if fss.lastBlockTimestamp == 0 {
		lastBlock, err := util.GetLastBlock(fss.Executor, fss.MainchainBlockQuery)

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

/*
IsInPhasePeriod calculate timestamp of recent block + constant.CommitPhaseEndDay,
Comparing with last block timestamp.
*/
func (fss *FeeScaleService) IsInPhasePeriod(timestamp int64) error {
	var (
		err       error
		lastBlock *model.Block
	)
	if timestamp != 0 {
		return fmt.Errorf("InvalidTimestamp")
	}

	lastBlock, err = util.GetLastBlock(fss.Executor, fss.MainchainBlockQuery)
	if err != nil {
		return err
	}

	recentBlockTime := time.Unix(timestamp, 0)
	recentBlockTime.AddDate(0, 0, constant.CommitPhaseEndDay)

	if time.Unix(lastBlock.GetTimestamp(), 0).Month() != recentBlockTime.Month() {
		return fmt.Errorf("TimeNotInPhasePeriodRange")
	}
	return nil
}
