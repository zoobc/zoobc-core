package service

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SnapshotFileServiceInterface interface {
		NewSnapshotFile(block *model.Block) (*model.SnapshotFileInfo, error)
	}
)
