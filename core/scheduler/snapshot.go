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
		CheckChunksIntegrity(chainType chaintype.ChainType) error
	}
)

func (ss *SnapshotScheduler) CheckChunksIntegrity(ct chaintype.ChainType) error {
	var (
		spineBlockManifest *model.SpineBlockManifest
		err, blockerErr    error
		chunks             [][]byte
	)

	spineBlockManifest, err = ss.SpineBlockManifestService.GetLastSpineBlockManifest(ct, model.SpineBlockManifestType_Snapshot)
	if err != nil {
		blockerErr = blocker.NewBlocker(blocker.SchedulerError, err.Error())
		ss.Logger.Warn(blockerErr.Error())
		return err
	}
	if spineBlockManifest == nil {
		blockerErr = blocker.NewBlocker(blocker.SchedulerError, "SpineBlockManifest is nil")
		ss.Logger.Warn(blockerErr.Error())
		return blockerErr
	}

	chunks, err = ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), sha256.Size)
	if err != nil {
		blockerErr = blocker.NewBlocker(blocker.SchedulerError, err.Error())
		ss.Logger.Warn(blockerErr.Error())
		return blockerErr
	}
	if len(chunks) == 0 {
		blockerErr = blocker.NewBlocker(blocker.SchedulerError, "Failed parsing File Chunk Hashes from Spine Block Manifest")
		ss.Logger.Warn(blockerErr.Error())
		return blockerErr
	}

	return nil
}
