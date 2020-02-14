package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"path/filepath"
)

type (
	SnapshotMainBlockService struct {
		SnapshotPath string
		chainType    chaintype.ChainType
		Logger       *log.Logger
		QueryService SnapshotMainBlockQueryServiceInterface
		FileService  FileServiceInterface
	}

	SnapshotMainBlockQueryServiceInterface interface {
		GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error)
	}

	SnapshotMainBlockQueryService struct {
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
		SnapshotPath: snapshotPath,
		chainType:    &chaintype.MainChain{},
		Logger:       logger,
		QueryService: &SnapshotMainBlockQueryService{
			QueryExecutor:             queryExecutor,
			SpineBlockManifestService: spineBlockManifestService,
			MainBlockQuery:            mainBlockQuery,
			AccountBalanceQuery:       accountBalanceQuery,
			NodeRegistrationQuery:     nodeRegistrationQuery,
			AccountDatasetQuery:       accountDatasetQuery,
			EscrowTransactionQuery:    escrowTransactionQuery,
			ParticipationScoreQuery:   participationScoreQuery,
		},
		FileService: &FileService{
			Logger: logger,
		},
	}
}

// GetAccountBalances get account balances for snapshot (wrapper function around account balance query)
func (smbq *SnapshotMainBlockQueryService) GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error) {
	qry := smbq.AccountBalanceQuery.GetAccountBalancesForSnapshot(fromHeight, toHeight)
	balanceRows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer balanceRows.Close()
	accountBalances, err := smbq.AccountBalanceQuery.BuildModel([]*model.AccountBalance{}, balanceRows)
	if err != nil {
		return nil, err
	}
	return accountBalances, nil

}

// NewSnapshotFile creates a new snapshot file (or multiple file chunks) and return the snapshotFileInfo
func (ss *SnapshotMainBlockService) NewSnapshotFile(block *model.Block, chunkSizeBytes int64) (*model.SnapshotFileInfo, error) {
	var (
		snapshotFullHash            []byte
		fileChunkHashes             = make([][]byte, 0)
		snapshotExpirationTimestamp int64
		enc                         *codec.Encoder
		h                           codec.Handle = new(codec.CborHandle)
		b                           []byte
		fileName                    string
	)
	enc = codec.NewEncoderBytes(&b, h)

	snapshotExpirationTimestamp = block.Timestamp + ss.chainType.GetSnapshotGenerationTimeout()

	// AccountBalance processing
	accountBalances, err := ss.QueryService.GetAccountBalances(0, block.Height)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(accountBalances)
	if err != nil {
		return nil, err
	}
	// TODO: STEF test only
	fmt.Printf("%v\n", b)

	//  the snapshot chunks' hashes
	//  the snapshot full hash
	// FIXME: below logic is only for live testing without real snapshots
	digest := sha3.New256()
	_, err = digest.Write(util.ConvertUint64ToBytes(uint64(snapshotExpirationTimestamp)))
	if err != nil {
		return nil, err
	}
	hash1 := ss.FileService.HashPayload(b)
	fileChunkHashes = append(fileChunkHashes, hash1)

	digest.Reset()
	_, err = digest.Write(util.ConvertUint64ToBytes(uint64(snapshotExpirationTimestamp + 1)))
	if err != nil {
		return nil, err
	}

	snapshotFullHash = ss.FileService.HashPayload(b)
	fileName, err = ss.FileService.GetFileNameFromHash(snapshotFullHash)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(ss.SnapshotPath, fileName)
	_, err = ss.FileService.SaveBytesToFile(filePath, b)
	if err != nil {
		return nil, err
	}

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
	snapshotInterval := ss.chainType.GetSnapshotInterval()
	//FIXME: uncomment this when we are sure that snapshot downloads work
	// if snapshotInterval < constant.MinRollbackBlocks {
	// 	if height < constant.MinRollbackBlocks {
	// 		return false
	// 	} else if height == constant.MinRollbackBlocks {
	// 		return true
	// 	}
	// 	return (constant.MinRollbackBlocks+height)%snapshotInterval == 0
	// }
	return height%snapshotInterval == 0

}
