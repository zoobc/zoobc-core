package fee

import (
	"math"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// BlockLifeTimeFeeModel will calculate the transaction fee based on expected lifetime on the chain. `timeout` field
	// must be present in the transaction body
	BlockLifeTimeFeeModel struct {
		blockPeriod       int64
		feePerBlockPeriod int64
	}
)

func NewBlockLifeTimeFeeModel(
	blockPeriod, feePerBlockPeriod int64,
) *BlockLifeTimeFeeModel {
	return &BlockLifeTimeFeeModel{
		blockPeriod:       blockPeriod,
		feePerBlockPeriod: feePerBlockPeriod,
	}
}

func (blt *BlockLifeTimeFeeModel) CalculateTxMinimumFee(
	txBody model.TransactionBodyInterface, escrow *model.Escrow,
) (int64, error) {
	// timeout / blockPeriod result is ceil
	fee := int64(math.Ceil(float64(escrow.Timeout)/float64(blt.blockPeriod))) * blt.feePerBlockPeriod
	return fee, nil
}
