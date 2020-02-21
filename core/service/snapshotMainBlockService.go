package service

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"os"
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

	SnapshotPayload struct {
		AccountBalances     []*model.AccountBalance
		NodeRegistrations   []*model.NodeRegistration
		AccountDatasets     []*model.AccountDataset
		ParticipationScores []*model.ParticipationScore
		PublishedReceipts   []*model.PublishedReceipt
		EscrowTransactions  []*model.Escrow
	}
)

func NewSnapshotMainBlockService(
	snapshotPath string,
	queryExecutor query.ExecutorInterface,
	spineBlockManifestService SpineBlockManifestServiceInterface,
	logger *log.Logger,
	fileService FileServiceInterface,
	mainBlockQuery query.BlockQueryInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	accountDatasetQuery query.AccountDatasetsQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	snapshotQueries map[string]query.SnapshotQuery,
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
			ParticipationScoreQuery:   participationScoreQuery,
			EscrowTransactionQuery:    escrowTransactionQuery,
			PublishedReceiptQuery:     publishedReceiptQuery,
			SnapshotQueries:           snapshotQueries,
		},
		FileService: fileService,
	}
}

// NewSnapshotFile creates a new snapshot file (or multiple file chunks) and return the snapshotFileInfo
func (ss *SnapshotMainBlockService) NewSnapshotFile(block *model.Block, chunkSizeBytes int64) (*model.SnapshotFileInfo, error) {
	var (
		fileChunkHashes             = make([][]byte, 0)
		snapshotPayload             = new(SnapshotPayload)
		err                         error
		snapshotExpirationTimestamp = block.Timestamp + int64(ss.chainType.GetSnapshotGenerationTimeout().Seconds())
		// (safe) height to get snapshot's data from
		snapshotPayloadHeight int = int(block.Height) - int(constant.MinRollbackBlocks)
	)

	if snapshotPayloadHeight <= 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			fmt.Sprintf("invalid snapshot height: %d", snapshotPayloadHeight))
	}

	snapshotPayload.AccountBalances, err = ss.QueryService.GetAccountBalances(0, uint32(snapshotPayloadHeight))
	if err != nil {
		return nil, err
	}
	snapshotPayload.NodeRegistrations, err = ss.QueryService.GetNodeRegistrations(0, uint32(snapshotPayloadHeight))
	if err != nil {
		return nil, err
	}
	snapshotPayload.AccountDatasets, err = ss.QueryService.GetAccountDatasets(0, uint32(snapshotPayloadHeight))
	if err != nil {
		return nil, err
	}
	snapshotPayload.ParticipationScores, err = ss.QueryService.GetParticipationScores(0, uint32(snapshotPayloadHeight))
	if err != nil {
		return nil, err
	}
	snapshotPayload.PublishedReceipts, err = ss.QueryService.GetPublishedReceipts(0, uint32(snapshotPayloadHeight),
		constant.LinkedReceiptBlocksLimit)
	if err != nil {
		return nil, err
	}
	snapshotPayload.EscrowTransactions, err = ss.QueryService.GetEscrowTransactions(0, uint32(snapshotPayloadHeight))
	if err != nil {
		return nil, err
	}

	// encode the snapshot payload
	b, err := ss.FileService.EncodePayload(snapshotPayload)
	if err != nil {
		return nil, err
	}

	//  the snapshot full hash
	digest := sha3.New256()
	_, err = digest.Write(util.ConvertUint64ToBytes(uint64(snapshotExpirationTimestamp)))
	if err != nil {
		return nil, err
	}
	digest.Reset()

	snapshotFullHash := ss.FileService.HashPayload(b)
	fileName, err := ss.FileService.GetFileNameFromHash(snapshotFullHash)
	if err != nil {
		return nil, err
	}
	err = ss.FileService.SaveBytesToFile(ss.SnapshotPath, fileName, b)
	if err != nil {
		return nil, err
	}
	// make extra sure that the file created is not corrupted
	filePath := filepath.Join(ss.SnapshotPath, fileName)
	match, err := ss.FileService.VerifyFileHash(filePath, snapshotFullHash)
	if err != nil || !match {
		// try remove saved file if file validation fails
		_ = os.Remove(filePath)
		return nil, err
	}
	// TODO: for now only whole snapshot is one file chunk
	fileChunkHashes = append(fileChunkHashes, snapshotFullHash)

	return &model.SnapshotFileInfo{
		SnapshotFileHash:           snapshotFullHash,
		FileChunksHashes:           fileChunkHashes,
		ChainType:                  ss.chainType.GetTypeInt(),
		Height:                     block.Height,
		ProcessExpirationTimestamp: snapshotExpirationTimestamp,
		SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
	}, nil
}

// ImportSnapshotFile parses a downloaded snapshot file into db
func (ss *SnapshotMainBlockService) ImportSnapshotFile(snapshotFileInfo *model.SnapshotFileInfo) error {
	var (
		snapshotPayload SnapshotPayload
		b               []byte
	)

	fileName, err := ss.FileService.GetFileNameFromHash(snapshotFileInfo.SnapshotFileHash)
	if err != nil {
		return err
	}
	filePath := filepath.Join(ss.SnapshotPath, fileName)
	b, err = ioutil.ReadFile(filePath)
	if err != nil {
		return blocker.NewBlocker(blocker.AppErr,
			fmt.Sprintf("Cannot read snapshot file from disk: %v", err))
	}

	payloadHash := sha3.Sum256(b)
	if !bytes.Equal(payloadHash[:], snapshotFileInfo.SnapshotFileHash) {
		return blocker.NewBlocker(blocker.ValidationErr,
			"Snapshot File Hash doesn't match with the one in database")
	}
	// decode the snapshot payload
	err = ss.FileService.DecodePayload(b, &snapshotPayload)
	if err != nil {
		return err
	}

	err = ss.QueryService.InsertSnapshotPayloadToDb(snapshotPayload)
	if err != nil {
		return err
	}

	return nil
}

// IsSnapshotHeight returns true if chain height passed is a snapshot height
func (ss *SnapshotMainBlockService) IsSnapshotHeight(height uint32) bool {
	snapshotInterval := ss.chainType.GetSnapshotInterval()
	if snapshotInterval < constant.MinRollbackBlocks {
		if height < constant.MinRollbackBlocks {
			return false
		}
		return (constant.MinRollbackBlocks+height)%snapshotInterval == 0
	}
	return height%snapshotInterval == 0

}
