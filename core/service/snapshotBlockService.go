package service

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// Snapshot logic specific of Mainchain blocks
	SnapshotBlockServiceInterface interface {
		NewSnapshotFile(block *model.Block, chunkSizeBytes int64) (*model.SnapshotFileInfo, error)
		IsSnapshotHeight(height uint32) bool
	}
)
