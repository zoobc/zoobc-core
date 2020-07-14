package scheduler

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

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

// CheckChunksIntegrity checking availability of snapshot read files from last manifest
func (ss *SnapshotScheduler) CheckChunksIntegrity(chainType chaintype.ChainType, filePath string) error {
	var (
		spineBlockManifest *model.SpineBlockManifest
		chunksHashed       [][]byte
		err                error
	)

	spineBlockManifest, err = ss.SpineBlockManifestService.GetLastSpineBlockManifest(chainType, model.SpineBlockManifestType_Snapshot)
	if err != nil {
		return blocker.NewBlocker(blocker.SchedulerError, err.Error())
	}
	// NOTE: Need to check this meanwhile err checked
	if spineBlockManifest == nil {
		return blocker.NewBlocker(blocker.SchedulerError, "SpineBlockManifest is nil")
	}

	chunksHashed, err = ss.FileService.ParseFileChunkHashes(spineBlockManifest.GetFileChunkHashes(), sha256.Size)
	if err != nil {
		return blocker.NewBlocker(blocker.SchedulerError, err.Error())
	}
	if len(chunksHashed) != 0 {
		for _, chunkHashed := range chunksHashed {
			_, err = ss.FileService.ReadFileFromDir(
				base64.URLEncoding.EncodeToString(spineBlockManifest.GetFileChunkHashes()),
				ss.FileService.GetFileNameFromHash(chunkHashed),
			)
			if err != nil {
				// Could be requesting a missing chunk p2p
				fmt.Println(err) // TODO: Will update when p2p finish
			}
		}
	}
	return blocker.NewBlocker(blocker.SchedulerError, "Failed parsing File Chunk Hashes from Spine Block Manifest")
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
			return err
		}

		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(block.GetHeight() - 1)
		if err != nil {
			return err
		}

		for _, spinePublicKey := range spinePublicKeys {
			nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
		}

		spineBlockManifest, err = ss.SpineBlockManifestService.GetSpineBlockManifestBySpineBlockHeight(block.GetHeight() - 1)
		if err != nil {
			return err
		}
		if spineBlockManifest != nil {

			shardMap, err = ss.SnapshotChunkUtil.GetShardAssigment(spineBlockManifest[0].GetFileChunkHashes(), sha256.Size, nodeIDs, false)
			if err != nil {
				return err
			}

			for _, shardChunk := range shardMap.ShardChunks {
				err = ss.FileService.DeleteFilesByHash(filePath, shardChunk)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}
