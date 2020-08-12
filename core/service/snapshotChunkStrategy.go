package service

import (
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	SnapshotChunkStrategyInterface interface {
		GenerateSnapshotChunks(snapshotPayload *model.SnapshotPayload) (fullHash []byte, fileChunkHashes [][]byte, err error)
		BuildSnapshotFromChunks(snapshotHash []byte, fileChunkHashes [][]byte) (*model.SnapshotPayload, error)
		DeleteFileByChunkHashes(concatenatedFileChunks []byte) error
	}
)
