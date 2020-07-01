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
		DeleteChunks(chainType chaintype.ChainType) error
	}
)

func (ss *SnapshotScheduler) CheckChunksIntegrity(chainType chaintype.ChainType, filePath string) error {
	panic("implement me")
}

func (ss *SnapshotScheduler) DeleteChunks(chainType chaintype.ChainType) error {
	var (
		err error
	)

	if (ss.NodeShardStorage.GetSize()) <= 0 { // Need to sharding
		var (
			block              *model.Block
			spineBlockManifest *model.SpineBlockManifest
			spinePublicKeys    []*model.SpinePublicKey
			nodeIDs            []int64
			// chunksHashed       [][]byte
		)
		block, err = ss.BlockCoreService.GetLastBlock()
		if err != nil {
			ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
			return err
		}

		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(block.GetHeight())
		if err != nil {
			ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
			return err
		}

		for _, spinePublicKey := range spinePublicKeys {
			nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
		}

		spineBlockManifest, err = ss.SpineBlockManifestService.GetLastSpineBlockManifest(chainType, model.SpineBlockManifestType_Snapshot)
		if err != nil {
			ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
			return err
		}
		if spineBlockManifest != nil {
			_, err = ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), sha256.Size)
			if err != nil {
				ss.Logger.Warn(blocker.NewBlocker(blocker.SchedulerError, err.Error()))
				return err
			}

		}

	}
	return nil
}
