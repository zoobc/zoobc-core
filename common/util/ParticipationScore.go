package util

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
)

// TODO: For now calculation score in float, next the func should optimize to avoid miss calculation in each node
func CalculateParticipationScore(linkedReceipt, unlinkedReceipt uint32) (float32, error) {
	if (linkedReceipt + unlinkedReceipt) > constant.MaxReceipt {
		return 0, blocker.NewBlocker(
			blocker.ValidationErr,
			"CalculateScoreParticipation, the number of receipt exceeds",
		)
	}

	maxBlockScore := float32(constant.MaxReceipt) * constant.LinkedReceiptScore
	halfMaxScore := maxBlockScore / 2

	blockScore := (float32(linkedReceipt) * constant.LinkedReceiptScore) + (float32(unlinkedReceipt) * constant.UnlinkedReceiptScore)
	comparionScore := (blockScore - halfMaxScore) / halfMaxScore

	scoreChangeOfANode := comparionScore * constant.MaxScoreChange
	return scoreChangeOfANode, nil
}
