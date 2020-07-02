package scheduler

import (
	"crypto/sha256"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	SnapshotScheduler struct {
		Logger                     *logrus.Logger
		SpineBlockManifestService  service.SpineBlockManifestServiceInterface
		FileService                service.FileServiceInterface
		SnapshotChunkUtil          util.ChunkUtil
		NodeShardStorage           storage.NodeShardCacheStorage
		NodeRegQuery               query.NodeRegistrationQueryInterface
		QueryExecutor              query.ExecutorInterface
		BlockCoreService           service.BlockServiceInterface
		BlockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface
	}
	SnapshotSchedulerService interface {
		CheckChunksIntegrity(chainType chaintype.ChainType, filePath string) error
		DeleteUnmaintainedChunks(filePath string) error
	}
)

func (ss *SnapshotScheduler) CheckChunksIntegrity(chainType chaintype.ChainType, filePath string) error {
	panic("implement me")
}

// DeleteUnmaintainedChunks deleting chunks in previous manifest that might be not unmaintained since new one already there
func (ss *SnapshotScheduler) DeleteUnmaintainedChunks(filePath string) error {
	var (
		err error
	)

	if (ss.NodeShardStorage.GetSize()) >= 0 {
		var (
			spineBlockManifest []*model.SpineBlockManifest
			spinePublicKeys    []*model.SpinePublicKey
			shardMap           storage.ShardMap
			nodeIDs            []int64
			block              *model.Block
		)

		block, err = ss.BlockCoreService.GetLastBlock()
		if err != nil {
			ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
			return err
		}

		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(block.GetHeight() - 1)
		if err != nil {
			ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
			return err
		}

		for _, spinePublicKey := range spinePublicKeys {
			nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
		}

		spineBlockManifest, err = ss.SpineBlockManifestService.GetSpineBlockManifestBySpineBlockHeight(block.GetHeight() - 1)
		if err != nil {
			ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
			return err
		}
		if spineBlockManifest != nil {

			shardMap, err = ss.SnapshotChunkUtil.GetShardAssigment(spineBlockManifest[0].GetFileChunkHashes(), sha256.Size, nodeIDs)
			if err != nil {
				ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
				return err
			}

			for _, shardChunk := range shardMap.ShardChunks {
				err = ss.FileService.DeleteFilesByHash(filePath, shardChunk)
				if err != nil {
					ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
					return err
				}
			}
		}

	}
	// No need to deleting

	return nil
}
