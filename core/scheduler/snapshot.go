// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
	"github.com/zoobc/zoobc-core/p2p"
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
		FileDownloaderService      p2p.FileDownloaderInterface
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
	fileDownloaderService p2p.FileDownloaderInterface,
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
		FileDownloaderService:      fileDownloaderService,
	}
}

// CheckChunksIntegrity checking availability of snapshot read files from last manifest
func (ss *SnapshotScheduler) CheckChunksIntegrity() error {
	var (
		err       error
		manifests []*model.SpineBlockManifest
		interval  int
		block     model.Block
	)

	err = ss.BlockStateStorage.GetItem(0, &block)
	if err != nil {
		return err
	}
	interval = int(block.GetHeight()) - int(constant.SnapshotSchedulerUnmaintainedChunksAtHeight)
	if interval <= 0 {
		return nil
	}

	manifests, err = ss.SpineBlockManifestService.GetSpineBlockManifestsByManifestReferenceHeightRange(uint32(interval), block.GetHeight())
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		var (
			spinePublicKeys  []*model.SpinePublicKey
			nodeIDs          []int64
			shards           storage.ShardMap
			snapshotDir      = base64.URLEncoding.EncodeToString(manifest.GetFileChunkHashes())
			snapshotFileInfo *model.SnapshotFileInfo
		)

		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(manifest.GetManifestSpineBlockHeight())
		if err != nil {
			return err
		}
		for _, spinePublicKey := range spinePublicKeys {
			nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
		}
		shards, err = ss.SnapshotChunkUtil.GetShardAssignment(manifest.GetFileChunkHashes(), sha256.Size, nodeIDs, false)
		if err != nil {
			return err
		}

		if shardNumbers, ok := shards.NodeShards[ss.NodeConfigurationService.GetHost().GetInfo().GetID()]; ok {
			var needToDownload bool
			for _, shardNumber := range shardNumbers {
				for _, chunkByte := range shards.ShardChunks[shardNumber] {
					_, err = ss.FileService.ReadFileFromDir(snapshotDir, ss.FileService.GetFileNameFromBytes(chunkByte))
					if err != nil {
						needToDownload = true
						break
					}
				}
			}

			if needToDownload {
				snapshotFileInfo, err = ss.FileDownloaderService.DownloadSnapshot(&chaintype.MainChain{}, manifest)
				if err != nil {
					return err
				}
				_, err = ss.FileService.SaveSnapshotChunks(
					base64.URLEncoding.EncodeToString(snapshotFileInfo.GetSnapshotFileHash()),
					snapshotFileInfo.GetFileChunksHashes(),
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
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
		spinePublicKeys, err = ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(manifest.GetManifestSpineBlockHeight())
		if err != nil {
			return err
		}
		for _, spinePublicKey := range spinePublicKeys {
			nodeIDs = append(nodeIDs, spinePublicKey.GetNodeID())
		}

		shards, err = ss.SnapshotChunkUtil.GetShardAssignment(manifest.GetFileChunkHashes(), sha256.Size, nodeIDs, false)
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
