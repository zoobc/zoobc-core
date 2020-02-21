package service

import (
	"bytes"
	"database/sql"
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
		SnapshotPath            string
		chainType               chaintype.ChainType
		Logger                  *log.Logger
		FileService             FileServiceInterface
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		AccountDatasetQuery     query.AccountDatasetsQueryInterface
		EscrowTransactionQuery  query.EscrowTransactionQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		SnapshotQueries         map[string]query.SnapshotQuery
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
	logger *log.Logger,
	fileService FileServiceInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	accountDatasetQuery query.AccountDatasetsQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	snapshotQueries map[string]query.SnapshotQuery,
) *SnapshotMainBlockService {
	return &SnapshotMainBlockService{
		SnapshotPath:            snapshotPath,
		chainType:               &chaintype.MainChain{},
		Logger:                  logger,
		FileService:             fileService,
		QueryExecutor:           queryExecutor,
		AccountBalanceQuery:     accountBalanceQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		AccountDatasetQuery:     accountDatasetQuery,
		ParticipationScoreQuery: participationScoreQuery,
		EscrowTransactionQuery:  escrowTransactionQuery,
		PublishedReceiptQuery:   publishedReceiptQuery,
		SnapshotQueries:         snapshotQueries,
	}
}

// NewSnapshotFile creates a new snapshot file (or multiple file chunks) and return the snapshotFileInfo
func (ss *SnapshotMainBlockService) NewSnapshotFile(block *model.Block, chunkSizeBytes int64) (snapshotFileInfo *model.SnapshotFileInfo,
	err error) {
	var (
		fileChunkHashes             = make([][]byte, 0)
		snapshotPayload             = new(SnapshotPayload)
		snapshotExpirationTimestamp = block.Timestamp + int64(ss.chainType.GetSnapshotGenerationTimeout().Seconds())
		// (safe) height to get snapshot's data from
		snapshotPayloadHeight int = int(block.Height) - int(constant.MinRollbackBlocks)
	)

	if snapshotPayloadHeight <= 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			fmt.Sprintf("invalid snapshot height: %d", snapshotPayloadHeight))
	}

	for qryRepoName, snapshotQuery := range ss.SnapshotQueries {
		func() {
			var (
				fromHeight uint32
				rows       *sql.Rows
			)
			if qryRepoName == "publishedReceipt" {
				if uint32(snapshotPayloadHeight) > constant.LinkedReceiptBlocksLimit {
					fromHeight = uint32(snapshotPayloadHeight) - constant.LinkedReceiptBlocksLimit
				}
			}
			qry := snapshotQuery.SelectDataForSnapshot(fromHeight, uint32(snapshotPayloadHeight))
			rows, err = ss.QueryExecutor.ExecuteSelect(qry, false)
			if err != nil {
				return
			}
			defer rows.Close()
			switch qryRepoName {
			case "accountBalance":
				snapshotPayload.AccountBalances, err = ss.AccountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
			case "nodeRegistration":
				snapshotPayload.NodeRegistrations, err = ss.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{},
					rows)
			case "accountDataset":
				snapshotPayload.AccountDatasets, err = ss.AccountDatasetQuery.BuildModel([]*model.AccountDataset{}, rows)
			case "participationScore":
				snapshotPayload.ParticipationScores, err = ss.ParticipationScoreQuery.BuildModel([]*model.
					ParticipationScore{}, rows)
			case "publishedReceipt":
				snapshotPayload.PublishedReceipts, err = ss.PublishedReceiptQuery.BuildModel([]*model.
					PublishedReceipt{}, rows)
			case "escrowTransaction":
				snapshotPayload.EscrowTransactions, err = ss.EscrowTransactionQuery.BuildModels(rows)
			default:
				err = blocker.NewBlocker(blocker.ParserErr, fmt.Sprintf("Invalid Snapshot Query Repository: %s", qryRepoName))
			}
		}()
		if err != nil {
			return nil, err
		}
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

	err = ss.InsertSnapshotPayloadToDb(snapshotPayload)
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

// InsertSnapshotPayloadToDb insert snapshot data to db
func (ss *SnapshotMainBlockService) InsertSnapshotPayloadToDb(payload SnapshotPayload) error {
	var (
		queries [][]interface{}
	)

	err := ss.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}

	for _, rec := range payload.AccountBalances {
		qry, args := ss.AccountBalanceQuery.InsertAccountBalance(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)

	}

	for _, rec := range payload.NodeRegistrations {
		qry, args := ss.NodeRegistrationQuery.InsertNodeRegistration(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.PublishedReceipts {
		qry, args := ss.PublishedReceiptQuery.InsertPublishedReceipt(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.ParticipationScores {
		qry, args := ss.ParticipationScoreQuery.InsertParticipationScore(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.EscrowTransactions {
		qryArgs := ss.EscrowTransactionQuery.InsertEscrowTransaction(rec)
		queries = append(queries, qryArgs...)
	}

	for _, rec := range payload.AccountDatasets {
		qryArgs := ss.AccountDatasetQuery.AddDataset(rec)
		queries = append(queries, qryArgs...)
	}

	err = ss.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := ss.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			ss.Logger.Error(rollbackErr.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("fail to insert snapshot into db: %v", err))
	}
	err = ss.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}
