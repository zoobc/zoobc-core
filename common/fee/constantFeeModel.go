package fee

import "github.com/zoobc/zoobc-core/common/model"

type (
	// ConstantFeeModel is fee that'll be set as a constant number
	ConstantFeeModel struct {
		constantFee int64
	}
)

func NewConstantFeeModel(constantFee int64) *ConstantFeeModel {
	return &ConstantFeeModel{
		constantFee: constantFee,
	}
}

func (cfm *ConstantFeeModel) CalculateTxMinimumFee(
	model.TransactionBodyInterface, *model.Escrow,
) (int64, error) {
	return cfm.constantFee, nil
}
