package service

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SnapshotChunkStrategyInterface interface {
		GenerateSnapshotChunks(snapshotPayload *model.SnapshotPayload, filePath string) (fullHash []byte,
			fileChunkHashes [][]byte, err error)
		BuildSnapshotFromChunks(fullHash []byte, fileChunkHashes [][]byte, filePath string) (*model.SnapshotPayload, error)
	}
)
