package testfixtures

import (
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

func GetFixtureForSnapshotBasicChunks(
	chunkSize int,
	fileService coreService.FileServiceInterface,
	payload *model.SnapshotPayload,
) (fullHash []byte, fileChunkHashes [][]byte, err error) {
	return coreService.NewSnapshotBasicChunkStrategy(chunkSize, fileService).GenerateSnapshotChunks(payload)
}
