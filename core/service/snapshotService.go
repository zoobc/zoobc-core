package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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
		GenerateSnapshot(block *model.Block, ct chaintype.ChainType) (*model.SnapshotFileInfo, error)
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
		SnapshotInterval          uint32
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

// GenerateSnapshot compute and persist a snapshot to file
// Note: First iteration will save a single chunk, for simplicity, but in future we should be able to split the file into multiple parts
// TODO: in future generalise (maybe by injecting a method from another service/strategy that implements logic specific to a given
//  chaintype. At the moment is not needed because we only have mainchain as chain type that can be snapshotted
func (ss *SnapshotService) GenerateSnapshot(block *model.Block, ct chaintype.ChainType) (*model.SnapshotFileInfo, error) {
	var (
		snapshotFullHash            []byte
		fileChunkHashes             = make([][]byte, 0)
		snapshotExpirationTimestamp int64
	)

	switch ct.(type) {
	case *chaintype.MainChain:
		snapshotExpirationTimestamp = block.Timestamp + constant.SnapshotGenerationTimeout

		// FIXME: call here the function that compute the snapshot and returns:
		//  the snapshot chunks' hashes
		//  the snapshot full hash
		// FIXME: below logic is only for live testing without real snapshots
		digest := sha3.New256()
		_, err := digest.Write(util.ConvertUint64ToBytes(uint64(util.GetSecureRandom())))
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
		SnapshotFileHash:           snapshotFullHash,
		FileChunksHashes:           fileChunkHashes,
		ChainType:                  ct.GetTypeInt(),
		Height:                     block.Height,
		ProcessExpirationTimestamp: snapshotExpirationTimestamp,
		SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
	}, nil

}

// StartSnapshotListener setup listener for transaction to the list peer
func (ss *SnapshotService) StartSnapshotListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			b := block.(*model.Block)
			ct := args.(chaintype.ChainType)
			if ct.HasSnapshots() && b.Height%constant.SnapshotInterval == 0 {
				go func() {
					// TODO: implement some sort of process management,
					//  such as controlling if there is another snapshot running before starting to compute a new one (
					//  or compute the new one and kill the one already running...)
					snapshotInfo, err := ss.GenerateSnapshot(b, &chaintype.MainChain{})
					if err != nil {
						ss.Logger.Errorf("Snapshot at block "+
							"height %d terminated with errors %s", b.Height, err)
					}
					_, err = ss.SpineBlockManifestService.CreateSpineBlockManifest(
						snapshotInfo.SnapshotFileHash,
						snapshotInfo.Height,
						snapshotInfo.ProcessExpirationTimestamp,
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
