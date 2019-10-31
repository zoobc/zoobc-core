package query

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
)

type (
	SkippedBlocksmithQueryInterface interface {
	}

	SkippedBlocksmithQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)
