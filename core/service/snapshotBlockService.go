package service

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SnapshotBlockServiceInterface interface {
		NewSnapshotFile(block *model.Block) (*model.SnapshotFileInfo, error)
		ImportSnapshotFile(snapshotFileInfo *model.SnapshotFileInfo) error
		IsSnapshotHeight(height uint32) bool
	}
)
