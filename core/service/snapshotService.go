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
package service

import (
	"crypto/sha256"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	// SnapshotServiceInterface snapshot logic shared across block types
	SnapshotServiceInterface interface {
		GenerateSnapshot(block *model.Block, ct chaintype.ChainType, chunkSizeBytes int) (*model.SnapshotFileInfo, error)
		IsSnapshotProcessing(ct chaintype.ChainType) bool
		StopSnapshotGeneration(ct chaintype.ChainType) error
		StartSnapshotListener() observer.Listener
	}

	SnapshotService struct {
		SpineBlockManifestService  SpineBlockManifestServiceInterface
		BlockSpinePublicKeyService BlockSpinePublicKeyServiceInterface
		SnapshotChunkUtil          util.ChunkUtilInterface
		BlockchainStatusService    BlockchainStatusServiceInterface
		SnapshotBlockServices      map[int32]SnapshotBlockServiceInterface // map key = chaintype number (eg. mainchain = 0)
		Logger                     *log.Logger
	}
)

var (
	// this map holds boolean channels to all block types that support snapshots
	stopSnapshotGeneration = make(map[int32]chan bool)
	// this map holds boolean values to all block types that support snapshots
	generatingSnapshot = model.NewMapIntBool()
)

func NewSnapshotService(
	spineBlockManifestService SpineBlockManifestServiceInterface,
	blockSpinePublicKeyService BlockSpinePublicKeyServiceInterface,
	blockchainStatusService BlockchainStatusServiceInterface,
	snapshotBlockServices map[int32]SnapshotBlockServiceInterface,
	snapshotChunkUtil util.ChunkUtilInterface,
	logger *log.Logger,
) *SnapshotService {
	return &SnapshotService{
		SpineBlockManifestService:  spineBlockManifestService,
		BlockSpinePublicKeyService: blockSpinePublicKeyService,
		BlockchainStatusService:    blockchainStatusService,
		SnapshotBlockServices:      snapshotBlockServices,
		SnapshotChunkUtil:          snapshotChunkUtil,
		Logger:                     logger,
	}
}

// GenerateSnapshot compute and persist a snapshot to file
func (ss *SnapshotService) GenerateSnapshot(block *model.Block, ct chaintype.ChainType,
	snapshotChunkBytesLength int) (*model.SnapshotFileInfo, error) {
	stopSnapshotGeneration[ct.GetTypeInt()] = make(chan bool)
	for {
		select {
		case <-stopSnapshotGeneration[ct.GetTypeInt()]:
			ss.Logger.Infof("Snapshot generation for block type %s at height %d has been stopped",
				ct.GetName(), block.GetHeight())
			break
		default:
			snapshotBlockService, ok := ss.SnapshotBlockServices[ct.GetTypeInt()]
			if !ok {
				return nil, fmt.Errorf("snapshots for chaintype %s not implemented", ct.GetName())
			}
			generatingSnapshot.Store(ct.GetTypeInt(), true)
			snapshotInfo, err := snapshotBlockService.NewSnapshotFile(block)
			generatingSnapshot.Store(ct.GetTypeInt(), false)
			return snapshotInfo, err
		}
	}
}

// StopSnapshotGeneration stops current snapshot generation
func (ss *SnapshotService) StopSnapshotGeneration(ct chaintype.ChainType) error {
	if !ss.IsSnapshotProcessing(ct) {
		return blocker.NewBlocker(blocker.AppErr, "No snapshots running: nothing to stop")
	}
	stopSnapshotGeneration[ct.GetTypeInt()] <- true
	// TODO implement error handling for abrupt snapshot termination. for now we just wait a few seconds and return
	time.Sleep(2 * time.Second) // todo: move the constant
	return nil
}

func (*SnapshotService) IsSnapshotProcessing(ct chaintype.ChainType) bool {
	if res, ok := generatingSnapshot.Load(ct.GetTypeInt()); ok {
		return res
	}
	return false
}

// StartSnapshotListener setup listener for snapshots generation
// TODO: allow only active blocksmiths (registered nodes at this block height) to generate snapshots
// 	 one way to do this is to inject the actual node public key and nodeRegistration service into this service
func (ss *SnapshotService) StartSnapshotListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(blockI interface{}, args ...interface{}) {
			block := blockI.(*model.Block)
			ct, ok := args[0].(chaintype.ChainType)
			if !ok {
				ss.Logger.Fatalln("chaintype casting failures in StartSnapshotListener")
			}
			if ct.HasSnapshots() {
				snapshotBlockService, ok := ss.SnapshotBlockServices[ct.GetTypeInt()]
				if !ok {
					ss.Logger.Errorf("snapshots for chaintype %s not implemented", ct.GetName())
					return
				}
				if snapshotBlockService.IsSnapshotHeight(block.Height) {
					go func() {
						// if spine and main blocks are still downloading, after the node has started,
						// do not generate (or download from other peers) snapshots
						if !ss.BlockchainStatusService.IsFirstDownloadFinished(&chaintype.MainChain{}) {
							ss.Logger.Infof("Snapshot at block "+
								"height %d not generated because blockchain is still downloading",
								block.Height)
							return
						}
						// if there is another snapshot running before this, kill the one already running
						if ss.IsSnapshotProcessing(ct) {
							if err := ss.StopSnapshotGeneration(ct); err != nil {
								ss.Logger.Infoln(err)
							}
						}
						snapshotInfo, err := ss.GenerateSnapshot(block, ct, constant.SnapshotChunkSize)
						if err != nil {
							ss.Logger.Errorf("Snapshot at block "+
								"height %d terminated with errors %s", block.Height, err)
							return
						}
						manifestRes, err := ss.SpineBlockManifestService.CreateSpineBlockManifest(
							snapshotInfo.SnapshotFileHash,
							snapshotInfo.Height,
							snapshotInfo.ProcessExpirationTimestamp,
							snapshotInfo.FileChunksHashes,
							ct,
							model.SpineBlockManifestType_Snapshot,
						)
						if err != nil {
							ss.Logger.Errorf("Cannot create spineBlockManifest at block "+
								"height %d. Error %s", block.Height, err)
						}
						spinePublicKeys, err :=
							ss.BlockSpinePublicKeyService.GetSpinePublicKeysByBlockHeight(manifestRes.GetManifestSpineBlockHeight())
						if err != nil {
							ss.Logger.Errorf("Fail to get spinePublicKey at "+
								"spineBlock height %d. Error %s", manifestRes.GetManifestSpineBlockHeight(), err)
						}
						var nodeIDs = make([]int64, len(spinePublicKeys))
						for i, key := range spinePublicKeys {
							nodeIDs[i] = key.NodeID
						}

						_, err = ss.SnapshotChunkUtil.GetShardAssignment(manifestRes.GetFileChunkHashes(), sha256.Size, nodeIDs, true)
						if err != nil {
							ss.Logger.Errorf("Fail calculating snapshot shard assignment at "+
								"spineBlock height %d. Error %s", manifestRes.GetManifestSpineBlockHeight(), err)
						}

						ss.Logger.Infof("Generated Snapshot at main block "+
							"height %d - spineBlock - %d", block.Height, manifestRes.GetManifestSpineBlockHeight())
					}()
				}
			}
		},
	}
}
