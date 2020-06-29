package scheduler

import (
	"crypto/sha256"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	SnapshotScheduler struct {
		Logger                    *logrus.Logger
		SpineBlockManifestService service.SpineBlockManifestServiceInterface
		FileService               service.FileServiceInterface
	}
	SnapshotSchedulerService interface {
		CheckChunksIntegrity(chainType chaintype.ChainType, filePath string) error
	}
)

// CheckChunksIntegrity checking availability of snapshot read files from last manifest
func (ss *SnapshotScheduler) CheckChunksIntegrity(chainType chaintype.ChainType, filePath string) error {
	var (
		spineBlockManifest *model.SpineBlockManifest
		chunksHashed       [][]byte
		err                error
	)

	spineBlockManifest, err = ss.SpineBlockManifestService.GetLastSpineBlockManifest(chainType, model.SpineBlockManifestType_Snapshot)
	if err != nil {
		ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
		return blocker.NewBlocker(blocker.SchedulerError, err.Error())
	}
	// NOTE: Need to check this meanwhile err checked
	if spineBlockManifest == nil {
		ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, "SpineBlockManifest is nil"))
		return blocker.NewBlocker(blocker.SchedulerError, "SpineBlockManifest is nil")
	}

	chunksHashed, err = ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), sha256.Size)
	if err != nil {
		ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
		return blocker.NewBlocker(blocker.SchedulerError, err.Error())
	}
	if len(chunksHashed) != 0 {
		for _, chunkHashed := range chunksHashed {
			_, err = ss.FileService.ReadFileByHash(filePath, chunkHashed)
			if err != nil {
				ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
				// Could be requesting a missing chunk p2p
			}
		}
	}
	ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, "Failed parsing File Chunk Hashes from Spine Block Manifest"))
	return blocker.NewBlocker(blocker.SchedulerError, "Failed parsing File Chunk Hashes from Spine Block Manifest")

}
