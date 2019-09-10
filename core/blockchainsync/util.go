package blockchainsync

import "github.com/zoobc/zoobc-core/common/constant"

func getMinRollbackHeight(currentHeight uint32) uint32 {
	if currentHeight < constant.MinRollbackBlocks {
		return 0
	}
	return currentHeight - constant.MinRollbackBlocks
}
