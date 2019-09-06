package util

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
)

// CalculateParticipationScore to calculate score change of node
func CalculateParticipationScore(linkedReceipt, unlinkedReceipt uint32) (int64, error) {
	if (linkedReceipt + unlinkedReceipt) > constant.MaxReceipt {
		return 0, blocker.NewBlocker(
			blocker.ValidationErr,
			"CalculateParticipationScore, the number of receipt exceeds",
		)
	}

	// Maximum score will get when create a block
	maxBlockScore := int64(float32(constant.MaxReceipt) * constant.LinkedReceiptScore * constant.SkalarReceiptScore)
	halfMaxBlockScore := maxBlockScore / 2

	linkedBlockScore := (float32(linkedReceipt) * constant.LinkedReceiptScore * constant.SkalarReceiptScore)
	unlinkedBlockScore := (float32(unlinkedReceipt) * constant.UnlinkedReceiptScore * constant.SkalarReceiptScore)
	blockScore := int64(linkedBlockScore + unlinkedBlockScore)

	scoreChangeOfANode := ((blockScore - halfMaxBlockScore) * constant.MaxScoreChange) / halfMaxBlockScore
	return scoreChangeOfANode, nil
}
