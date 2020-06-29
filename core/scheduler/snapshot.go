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
		CheckChunksIntegrity(chainType chaintype.ChainType)
	}
)

func (ss *SnapshotScheduler) CheckChunksIntegrity(chainType chaintype.ChainType) {
	var (
		spineBlockManifest *model.SpineBlockManifest
		err, blockerErr    error
		chunks             [][]byte
	)

	spineBlockManifest, err = ss.SpineBlockManifestService.GetLastSpineBlockManifest(chainType, model.SpineBlockManifestType_Snapshot)
	if err != nil {
		ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
		return
	}
	if spineBlockManifest == nil {
		ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, "SpineBlockManifest is nil"))
		return
	}

	chunks, err = ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), sha256.Size)
	if err != nil {
		ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
		return
	}
	if len(chunks) == 0 {

		blockerErr = blocker.NewBlocker(blocker.SchedulerError, "Failed parsing File Chunk Hashes from Spine Block Manifest")
		ss.Logger.Warn(blockerErr.Error())
		return
	}
}
