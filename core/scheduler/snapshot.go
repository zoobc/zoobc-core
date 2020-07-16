package scheduler

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	// SnapshotScheduler struct containing fields that needed
	SnapshotScheduler struct {
		SpineBlockManifestService  service.SpineBlockManifestServiceInterface
		FileService                service.FileServiceInterface
		SnapshotChunkUtil          util.ChunkUtilInterface
		NodeShardStorage           storage.CacheStorageInterface
		BlockCoreService           service.BlockServiceInterface
		BlockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface
		NodeConfigurationService   service.NodeConfigurationServiceInterface
	}
	// SnapshotSchedulerService bounce of methods of snapshot scheduler service
	// SnapshotSchedulerService interface {
	// 	CheckChunksIntegrity(chainType chaintype.ChainType) error
	// 	DeleteUnmaintainedChunks() (err error)
	// }
)

func NewSnapshotScheduler(
	spineBlockManifestService service.SpineBlockManifestServiceInterface,
	fileService service.FileServiceInterface,
	snapshotChunkUtil util.ChunkUtilInterface,
	nodeShardStorage storage.CacheStorageInterface,
	blockCoreService service.BlockServiceInterface,
	blockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface,
	nodeConfigurationService service.NodeConfigurationServiceInterface,
) *SnapshotScheduler {
	return &SnapshotScheduler{
		SpineBlockManifestService:  spineBlockManifestService,
		FileService:                fileService,
		SnapshotChunkUtil:          snapshotChunkUtil,
		NodeShardStorage:           nodeShardStorage,
		BlockCoreService:           blockCoreService,
		BlockSpinePublicKeyService: blockSpinePublicKeyService,
		NodeConfigurationService:   nodeConfigurationService,
	}
}

// CheckChunksIntegrity checking availability of snapshot read files from last manifest
func (ss *SnapshotScheduler) CheckChunksIntegrity(chainType chaintype.ChainType) error {
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

// DeleteUnmaintainedChunks deleting chunks in previous manifest that might be not maintained since new one already there
func (ss *SnapshotScheduler) DeleteUnmaintainedChunks() (err error) {

	if (ss.NodeShardStorage.GetSize()) >= 0 {
		var (
			spineBlockManifest []*model.SpineBlockManifest
			spinePublicKeys    []*model.SpinePublicKey
			block              *model.Block
			prevBlockHeight    uint32
		)

		block, err = ss.BlockCoreService.GetLastBlock()
		if err != nil {
			return err
		}
		prevBlockHeight = block.GetHeight() - 1
		if prevBlockHeight < 1 {
			/*
				Which mean there isn't previous
				and no need to continuing the process
			*/
			return nil
		}

		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(prevBlockHeight)
		if err != nil {
			return err
		}

		spineBlockManifest, err = ss.SpineBlockManifestService.GetSpineBlockManifestBySpineBlockHeight(prevBlockHeight)
		if err != nil {
			return err
		}

		if spineBlockManifest != nil {
			var (
				snapshotDir = base64.URLEncoding.EncodeToString(spineBlockManifest[0].GetFileChunkHashes())
				shards      storage.ShardMap
				nodeIDs     []int64
			)
			for _, spinePublicKey := range spinePublicKeys {
				nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
			}

			shards, err = ss.SnapshotChunkUtil.GetShardAssigment(spineBlockManifest[0].GetFileChunkHashes(), sha256.Size, nodeIDs, false)
			if err != nil {
				return err
			}

			if shardNumbers, ok := shards.NodeShards[ss.NodeConfigurationService.GetHost().GetInfo().GetID()]; ok {
				for _, shardNumber := range shardNumbers {
					delete(shards.ShardChunks, shardNumber)
				}

				for _, shardChunk := range shards.ShardChunks {
					for _, chunkByte := range shardChunk {
						err = ss.FileService.DeleteSnapshotChunkFromDir(
							snapshotDir,
							ss.FileService.GetFileNameFromBytes(chunkByte),
						)
						if err != nil {
							return blocker.NewBlocker(
								blocker.SchedulerError,
								fmt.Sprintf(
									"failed deleting %s from %s: %s",
									ss.FileService.GetFileNameFromBytes(chunkByte),
									snapshotDir,
									err.Error(),
								),
							)
						}
					}
				}
			}
		}
	}
	return nil
}
