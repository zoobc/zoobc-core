package scheduler

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
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
		BlockStateStorage          storage.CacheStorageInterface
		BlockCoreService           service.BlockServiceInterface
		BlockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface
		NodeConfigurationService   service.NodeConfigurationServiceInterface
	}
)

func NewSnapshotScheduler(
	spineBlockManifestService service.SpineBlockManifestServiceInterface,
	fileService service.FileServiceInterface,
	snapshotChunkUtil util.ChunkUtilInterface,
	nodeShardStorage, blockStateStorage storage.CacheStorageInterface,
	blockCoreService service.BlockServiceInterface,
	blockSpinePublicKeyService service.BlockSpinePublicKeyServiceInterface,
	nodeConfigurationService service.NodeConfigurationServiceInterface,
) *SnapshotScheduler {
	return &SnapshotScheduler{
		SpineBlockManifestService:  spineBlockManifestService,
		FileService:                fileService,
		SnapshotChunkUtil:          snapshotChunkUtil,
		NodeShardStorage:           nodeShardStorage,
		BlockStateStorage:          blockStateStorage,
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

	var (
		manifests []*model.SpineBlockManifest
		interval  int
		block     model.Block
	)

	err = ss.BlockStateStorage.GetItem(0, &block)
	if err != nil {
		return nil
	}
	interval = int(block.GetHeight()) - int(constant.SnapshotSchedulerUnmaintainedChunksAtHeight)

	if interval <= 0 {
		// No need to continuing the process
		return
	}

	manifests, err = ss.SpineBlockManifestService.GetSpineBlockManifestsByManifestReferenceHeightRange(
		block.GetHeight(),
		uint32(interval),
	)
	if err != nil {
		return err
	}

	for _, manifest := range manifests {
		var (
			spinePublicKeys []*model.SpinePublicKey
			snapshotDir     = base64.URLEncoding.EncodeToString(manifest.GetFileChunkHashes())
			nodeIDs         []int64
			shards          storage.ShardMap
		)
		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(manifest.GetManifestReferenceHeight())
		if err != nil {
			return err
		}
		for _, spinePublicKey := range spinePublicKeys {
			nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
		}

		shards, err = ss.SnapshotChunkUtil.GetShardAssigment(manifest.GetFileChunkHashes(), sha256.Size, nodeIDs, false)
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
	return nil
}
