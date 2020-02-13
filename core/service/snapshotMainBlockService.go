package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	SnapshotMainBlockService struct {
		chainType                 chaintype.ChainType
		SnapshotPath              string
		QueryExecutor             query.ExecutorInterface
		SpineBlockManifestService SpineBlockManifestServiceInterface
		Logger                    *log.Logger
		MainBlockQuery            query.BlockQueryInterface
		AccountBalanceQuery       query.AccountBalanceQueryInterface
		NodeRegistrationQuery     query.NodeRegistrationQueryInterface
		ParticipationScoreQuery   query.ParticipationScoreQueryInterface
		AccountDatasetQuery       query.AccountDatasetsQueryInterface
		EscrowTransactionQuery    query.EscrowTransactionQueryInterface
	}
)

func NewSnapshotMainBlockService(
	snapshotPath string,
	queryExecutor query.ExecutorInterface,
	spineBlockManifestService SpineBlockManifestServiceInterface,
	logger *log.Logger,
	mainBlockQuery query.BlockQueryInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	accountDatasetQuery query.AccountDatasetsQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
) *SnapshotMainBlockService {
	return &SnapshotMainBlockService{
		chainType:                 &chaintype.MainChain{},
		SnapshotPath:              snapshotPath,
		QueryExecutor:             queryExecutor,
		SpineBlockManifestService: spineBlockManifestService,
		Logger:                    logger,
		MainBlockQuery:            mainBlockQuery,
		AccountBalanceQuery:       accountBalanceQuery,
		NodeRegistrationQuery:     nodeRegistrationQuery,
		AccountDatasetQuery:       accountDatasetQuery,
		EscrowTransactionQuery:    escrowTransactionQuery,
		ParticipationScoreQuery:   participationScoreQuery,
	}
}

// NewSnapshotFile creates a new snapshot file (or multiple file chunks) and return the snapshotFileInfo
func (ss *SnapshotMainBlockService) NewSnapshotFile(block *model.Block, chunkSizeBytes int64) (*model.SnapshotFileInfo, error) {
	var (
		snapshotFullHash            []byte
		fileChunkHashes             = make([][]byte, 0)
		snapshotExpirationTimestamp int64
	)

	snapshotExpirationTimestamp = block.Timestamp + ss.chainType.GetSnapshotGenerationTimeout()

	// FIXME: call here the function that compute the snapshot and returns:
	//  the snapshot chunks' hashes
	//  the snapshot full hash
	// FIXME: below logic is only for live testing without real snapshots
	digest := sha3.New256()
	_, err := digest.Write(util.ConvertUint64ToBytes(uint64(snapshotExpirationTimestamp)))
	if err != nil {
		return nil, err
	}
	hash1 := digest.Sum([]byte{})
	fileChunkHashes = append(fileChunkHashes, hash1)

	digest.Reset()
	_, err = digest.Write(util.ConvertUint64ToBytes(uint64(snapshotExpirationTimestamp + 1)))
	if err != nil {
		return nil, err
	}
	snapshotFullHash = digest.Sum([]byte{})
	return &model.SnapshotFileInfo{
		SnapshotFileHash:           snapshotFullHash,
		FileChunksHashes:           fileChunkHashes,
		ChainType:                  ss.chainType.GetTypeInt(),
		Height:                     block.Height,
		ProcessExpirationTimestamp: snapshotExpirationTimestamp,
		SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
	}, nil
}

// IsSnapshotHeight returns true if chain height passed is a snapshot height
func (ss *SnapshotMainBlockService) IsSnapshotHeight(height uint32) bool {
	//FIXME: uncomment this when we are sure that snapshot downloads work
	// if snapshotInterval < constant.MinRollbackBlocks {
	// 	if height < constant.MinRollbackBlocks {
	// 		return false
	// 	} else if height == constant.MinRollbackBlocks {
	// 		return true
	// 	}
	// 	return (constant.MinRollbackBlocks+height)%snapshotInterval == 0
	// }
	return height%ss.chainType.GetSnapshotInterval() == 0

}
