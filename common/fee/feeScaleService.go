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
package fee

import (
	"math"
	"time"

	"github.com/montanaflynn/stats"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	FeeScaleServiceInterface interface {
		InsertFeeScale(feeScale *model.FeeScale) error
		GetLatestFeeScale(feeScale *model.FeeScale) error
		GetCurrentPhase(
			blockTimestamp int64,
			isPostTransaction bool,
		) (phase model.FeeVotePhase, canAdjust bool, err error)
		SelectVote(votes []*model.FeeVoteInfo, originalSendZBCFee int64) int64
	}

	FeeScaleService struct {
		lastBlockTimestamp    int64
		lastFeeScale          model.FeeScale
		FeeScaleQuery         query.FeeScaleQueryInterface
		MainBlockStateStorage storage.CacheStorageInterface
		Executor              query.ExecutorInterface
	}
)

func NewFeeScaleService(
	feeScaleQuery query.FeeScaleQueryInterface,
	mainBlockStateStorage storage.CacheStorageInterface,
	executor query.ExecutorInterface,
) *FeeScaleService {
	return &FeeScaleService{
		FeeScaleQuery:         feeScaleQuery,
		MainBlockStateStorage: mainBlockStateStorage,
		Executor:              executor,
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
// todo: @andy-shi88 discard `isPostTransaction` parameter as there is no way to flag that in validate function with current state
// of the code
func (fss *FeeScaleService) GetCurrentPhase(
	blockTimestamp int64,
	isPostTransaction bool,
) (phase model.FeeVotePhase, canAdjust bool, err error) {
	// check if lastBlockstimestamp is 0
	if fss.lastBlockTimestamp == 0 || blockTimestamp < fss.lastBlockTimestamp {
		if blockTimestamp == constant.MainchainGenesisBlockTimestamp { // genesis exception
			return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
		}
		var (
			lastBlock model.Block
			err       = fss.MainBlockStateStorage.GetItem(nil, &lastBlock)
		)
		if err != nil {
			return model.FeeVotePhase_FeeVotePhaseCommmit, false, err
		}
		if lastBlock.Timestamp == constant.MainchainGenesisBlockTimestamp { // genesis exception
			fss.lastBlockTimestamp = blockTimestamp
			return model.FeeVotePhase_FeeVotePhaseCommmit, false, nil
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

// SelectVote return the scaled vote relative to original / unscaled send-money fee
func (fss *FeeScaleService) SelectVote(votes []*model.FeeVoteInfo, originalSendZBCFee int64) int64 {
	var (
		floats stats.Float64Data
		err    error
	)
	// sort votes and get median value
	for _, vote := range votes {
		floats = append(floats, float64(vote.FeeVote))
	}
	median, err := stats.Median(floats)
	if err != nil { // stats.Median can only return stats.EmptyInputErr
		return fss.lastFeeScale.FeeScale
	}
	// constraints 0.5 to 2.0 from previous scale
	scale := math.Floor(median / float64(originalSendZBCFee) * float64(constant.OneZBC))
	compareToPreviousScale := scale / float64(fss.lastFeeScale.FeeScale)
	if compareToPreviousScale < FeeScaleLowerConstraints {
		scale = math.Floor(FeeScaleLowerConstraints * float64(fss.lastFeeScale.FeeScale))
	} else if compareToPreviousScale > 2.0 {
		scale = math.Floor(FeeScaleUpperConstraints * float64(fss.lastFeeScale.FeeScale))
	}
	// scale median value to currentSendZBCFee
	return int64(scale)
}
