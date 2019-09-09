package blockchainsync

import "github.com/zoobc/zoobc-core/common/constant"

func getMinRollbackHeight(currentHeight uint32) (uint32, error) {
	if currentHeight < constant.MinRollbackBlocks {
		return 0, nil
	}
	return currentHeight - constant.MinRollbackBlocks, nil
}
