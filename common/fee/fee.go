package fee

import "github.com/zoobc/zoobc-core/common/model"

type (
	FeeModelInterface interface {
		CalculateTxMinimumFee(txBody model.TransactionBodyInterface, escrow *model.Escrow) (int64, error)
	}
)
