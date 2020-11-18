package util

import (
	"fmt"
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
)

// CalculateParticipationScore to calculate score change of node
func CalculateParticipationScore(linkedReceipt, unlinkedReceipt, maxReceipt uint32) (int64, error) {
	if maxReceipt == 0 {
		return constant.MaxScoreChange, nil
	}
	if (linkedReceipt + unlinkedReceipt) > maxReceipt {
		return 0, blocker.NewBlocker(
			blocker.ValidationErr,
			fmt.Sprintf("CalculateParticipationScore, the number of receipt exceeds get %d max allowed %d",
				linkedReceipt+unlinkedReceipt, maxReceipt,
			),
		)
	}

	// Maximum score will get when create a block
	maxBlockScore := int64(float32(maxReceipt) * constant.LinkedReceiptScore * constant.ScalarReceiptScore)
	halfMaxBlockScore := maxBlockScore / 2

	linkedBlockScore := float32(linkedReceipt) * constant.LinkedReceiptScore * constant.ScalarReceiptScore
	unlinkedBlockScore := float32(unlinkedReceipt) * constant.UnlinkedReceiptScore * constant.ScalarReceiptScore
	blockScore := int64(linkedBlockScore + unlinkedBlockScore)

	scoreDiffBig := new(big.Int).SetInt64(blockScore - halfMaxBlockScore)
	scoreDiffBigMul := new(big.Int).Mul(scoreDiffBig, new(big.Int).SetInt64(constant.MaxScoreChange))
	scoreChangeOfANode := new(big.Int).Div(scoreDiffBigMul, new(big.Int).SetInt64(halfMaxBlockScore))
	return scoreChangeOfANode.Int64(), nil
}

func GetReceiptValue(linkedReceipt, unlinkedReceipt uint32) int64 {
	linkedBlockScore := float32(linkedReceipt) * constant.LinkedReceiptScore * constant.ScalarReceiptScore
	unlinkedBlockScore := float32(unlinkedReceipt) * constant.UnlinkedReceiptScore * constant.ScalarReceiptScore
	blockScore := int64(linkedBlockScore + unlinkedBlockScore)
	return blockScore
}
