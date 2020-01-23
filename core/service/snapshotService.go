package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
)

type (
	SnapshotServiceInterface interface {
		GetNextSnapshotHeight(mainHeight uint32, ct chaintype.ChainType) uint32
		GenerateSnapshot(mainHeight uint32, ct chaintype.ChainType) (*model.SnapshotFileInfo, error)
		StartSnapshotListener() observer.Listener
	}

	SnapshotService struct {
		QueryExecutor   query.ExecutorInterface
		SpineBlockQuery query.BlockQueryInterface
		MainBlockQuery  query.BlockQueryInterface
		Logger          *log.Logger
		// below fields are for better code testability
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
		SpineBlockManifestService SpineBlockManifestServiceInterface
	}
)

func NewSnapshotService(
	queryExecutor query.ExecutorInterface,
	mainBlockQuery, spineBlockQuery query.BlockQueryInterface,
	spineBlockManifestService SpineBlockManifestServiceInterface,
	logger *log.Logger,
) *SnapshotService {
	return &SnapshotService{
		QueryExecutor:             queryExecutor,
		SpineBlockQuery:           spineBlockQuery,
		MainBlockQuery:            mainBlockQuery,
		Spinechain:                &chaintype.SpineChain{},
		Mainchain:                 &chaintype.MainChain{},
		SnapshotInterval:          constant.SnapshotInterval,
		SnapshotGenerationTimeout: constant.SnapshotGenerationTimeout,
		SpineBlockManifestService: spineBlockManifestService,
		Logger:                    logger,
	}
}

// GetNextSnapshotHeight calculate next snapshot (main block) height given an arbitrary main block height
// snapshotHeight is the height, on the chain type been snapshotted at which the snapshot is taken (start computing)
func (ss *SnapshotService) GetNextSnapshotHeight(snapshotHeight uint32, ct chaintype.ChainType) uint32 {
	var (
		avgBlockTime int64
	)
	// first snapshot cannot be taken before minRollBack height
	// FIXME: uncomment this. for testing only!
	// if snapshotHeight < constant.MinRollbackBlocks {
	// 	snapshotHeight = constant.MinRollbackBlocks
	// }
	switch ct.(type) {
	case *chaintype.MainChain:
		avgBlockTime = ss.Mainchain.GetSmithingPeriod() + ss.Mainchain.GetChainSmithingDelayTime()
	default:
		// for now, only mainchain is supported
		ss.Logger.Fatalf("block type not supported for snapshots!")
	}

	avgBlockInterval := ss.SnapshotInterval / avgBlockTime
	return uint32(util.GetNextStep(int64(snapshotHeight), avgBlockInterval))
}

// GenerateSnapshot compute and persist a snapshot to file
// Note: First iteration will save a single chunk, for simplicity, but in future we should be able to split the file into multiple parts
// TODO: in future generalise (maybe by injecting a method from another service/strategy that implements logic specific to a given
//  chaintype. At the moment is not needed because we only have mainchain as chain type that can be snapshotted
func (ss *SnapshotService) GenerateSnapshot(mainHeight uint32, ct chaintype.ChainType) (*model.SnapshotFileInfo, error) {
	var (
		lastMainBlock, lastSpineBlock model.Block
		firstValidSpineHeight         uint32
		snapshotFullHash              []byte
		fileChunkHashes               = make([][]byte, 0)
	)

	switch ct.(type) {
	case *chaintype.MainChain:
		// get the last main block
		row, err := ss.QueryExecutor.ExecuteSelectRow(ss.MainBlockQuery.GetLastBlock(), false)
		if err != nil {
			return nil, err
		}
		err = ss.MainBlockQuery.Scan(&lastMainBlock, row)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		// get the last spine block
		row, err = ss.QueryExecutor.ExecuteSelectRow(ss.SpineBlockQuery.GetLastBlock(), false)
		if err != nil {
			return nil, err
		}
		err = ss.MainBlockQuery.Scan(&lastSpineBlock, row)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}

		// calculate first valid spine block height for the snapshot (= spineBlockManifest) to be included in.
		// spine blocks have discrete timing,
		// so we can calculate accurately next spine timestamp and give enough time to all nodes to complete their snapshot
		spinechainInterval := ss.Spinechain.GetSmithingPeriod() + ss.Spinechain.GetChainSmithingDelayTime()
		// lastMainBlock.Timestamp is the timestamp at which the snapshot started to be computed
		nextMinimumSpineBlockTimestamp := lastMainBlock.Timestamp + ss.SnapshotGenerationTimeout
		firstValidTime := util.GetNextStep(nextMinimumSpineBlockTimestamp, spinechainInterval)
		// firstValidSpineHeight = previous spine block height + delta (in spine in blocks) based on previously computed first valid timestamp
		firstValidSpineHeight = lastSpineBlock.Height + uint32((firstValidTime-lastMainBlock.Timestamp)/spinechainInterval)
		// don't allow megablocks to reference past spine blocks
		if firstValidSpineHeight < lastSpineBlock.Height {
			firstValidSpineHeight = lastSpineBlock.Height + 1
		}

		// TODO: call here the function that compute the snapshot and returns:
		//  the snapshot chunks' hashes
		//  the snapshot full hash
		// TODO: below logic is only for live testing without real snapshots
		digest := sha3.New256()
		_, err = digest.Write(util.ConvertUint64ToBytes(uint64(util.GetSecureRandom())))
		if err != nil {
			return nil, err
		}
		hash1 := digest.Sum([]byte{})
		fileChunkHashes = append(fileChunkHashes, hash1)

		digest.Reset()
		_, err = digest.Write(util.ConvertUint64ToBytes(uint64(util.GetSecureRandom())))
		if err != nil {
			return nil, err
		}
		snapshotFullHash = digest.Sum([]byte{})
	default:
		// for now, only mainchain is supported
		return nil, fmt.Errorf("snapshot won't be generated for chain type %s", ct.GetName())
	}

	return &model.SnapshotFileInfo{
		ChainType:              ct.GetTypeInt(),
		SpineBlockManifestType: model.SpineBlockManifestType_Snapshot,
		FileChunksHashes:       fileChunkHashes,
		MainHeight:             mainHeight,
		SnapshotFileHash:       snapshotFullHash,
		SpineHeight:            firstValidSpineHeight,
	}, nil

}

// StartSnapshotListener setup listener for transaction to the list peer
func (ss *SnapshotService) StartSnapshotListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			b := block.(*model.Block)
			ct := args.(chaintype.ChainType)
			if ct.HasSnapshots() && b.Height == ss.GetNextSnapshotHeight(b.Height, ct) {
				go func() {
					// TODO: implement some process management,
					//  such as controlling if there is another snapshot running before starting to compute a new one (
					//  or compute the new one and kill the one already running...)
					snapshotInfo, err := ss.GenerateSnapshot(b.Height, &chaintype.MainChain{})
					if err != nil {
						ss.Logger.Errorf("Snapshot at block "+
							"height %d terminated with errors %s", b.Height, err)
					}
					snapshotExpirationTimestamp := b.Timestamp + constant.SnapshotGenerationTimeout
					_, err = ss.SpineBlockManifestService.CreateSpineBlockManifest(
						snapshotInfo.SnapshotFileHash,
						snapshotInfo.MainHeight,
						snapshotExpirationTimestamp,
						snapshotInfo.FileChunksHashes,
						ct,
						model.SpineBlockManifestType_Snapshot,
					)
					if err != nil {
						ss.Logger.Errorf("Cannot create spineBlockManifest at block "+
							"height %d. Error %s", b.Height, err)
					}
					ss.Logger.Infof("Snapshot at main block "+
						"height %d terminated successfully", b.Height)
				}()
			}
		},
	}
}
