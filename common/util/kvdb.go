package util

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
)

// GetKvDbMempoolDBKey returns the mempool key for a given chaintype
func GetKvDbMempoolDBKey(ct chaintype.ChainType) string {
	return fmt.Sprintf("%s_%s", constant.KVDBMempoolsBackup, ct.GetName())
}
