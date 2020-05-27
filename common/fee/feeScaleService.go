package fee

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	FeeScaleServiceInterface interface {
		InsertFeeScale(feeScale *model.FeeScale, dbTx bool) error
		GetLatestFeeScale(feeScale *model.FeeScale) error
	}

	FeeScaleService struct {
		lastFeeScale  model.FeeScale
		feeScaleQuery query.FeeScaleQueryInterface
		executor      query.ExecutorInterface
	}
)

func NewFeeScaleService(
	feeScaleQuery query.FeeScaleQueryInterface,
	executor query.ExecutorInterface,
) *FeeScaleService {
	return &FeeScaleService{
		feeScaleQuery: feeScaleQuery,
		executor:      executor,
	}
}

func (fss *FeeScaleService) InsertFeeScale(feeScale *model.FeeScale, dbTx bool) error {
	insertQry, args := fss.feeScaleQuery.InsertFeeScale(feeScale)
	if dbTx {
		return fss.executor.ExecuteTransaction(insertQry, args...)
	}
	_, err := fss.executor.ExecuteStatement(insertQry, args...)
	return err
}

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
