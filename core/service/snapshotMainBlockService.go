package service

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	SnapshotMainBlockService struct {
		SnapshotPath               string
		chainType                  chaintype.ChainType
		Logger                     *log.Logger
		SnapshotBasicChunkStrategy SnapshotChunkStrategyInterface
		QueryExecutor              query.ExecutorInterface
		AccountBalanceQuery        query.AccountBalanceQueryInterface
		NodeRegistrationQuery      query.NodeRegistrationQueryInterface
		ParticipationScoreQuery    query.ParticipationScoreQueryInterface
		AccountDatasetQuery        query.AccountDatasetsQueryInterface
		EscrowTransactionQuery     query.EscrowTransactionQueryInterface
		PublishedReceiptQuery      query.PublishedReceiptQueryInterface
		SnapshotQueries            map[string]query.SnapshotQuery
	}
)

func NewSnapshotMainBlockService(
	snapshotPath string,
	queryExecutor query.ExecutorInterface,
	logger *log.Logger,
	snapshotChunkStrategy SnapshotChunkStrategyInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	accountDatasetQuery query.AccountDatasetsQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	snapshotQueries map[string]query.SnapshotQuery,
) *SnapshotMainBlockService {
	return &SnapshotMainBlockService{
		SnapshotPath:               snapshotPath,
		chainType:                  &chaintype.MainChain{},
		Logger:                     logger,
		SnapshotBasicChunkStrategy: snapshotChunkStrategy,
		QueryExecutor:              queryExecutor,
		AccountBalanceQuery:        accountBalanceQuery,
		NodeRegistrationQuery:      nodeRegistrationQuery,
		AccountDatasetQuery:        accountDatasetQuery,
		ParticipationScoreQuery:    participationScoreQuery,
		EscrowTransactionQuery:     escrowTransactionQuery,
		PublishedReceiptQuery:      publishedReceiptQuery,
		SnapshotQueries:            snapshotQueries,
	}
}

// NewSnapshotFile creates a new snapshot file (or multiple file chunks) and return the snapshotFileInfo
func (ss *SnapshotMainBlockService) NewSnapshotFile(block *model.Block) (snapshotFileInfo *model.SnapshotFileInfo,
	err error) {
	var (
		snapshotFileHash            []byte
		fileChunkHashes             [][]byte
		snapshotPayload             = new(model.SnapshotPayload)
		snapshotExpirationTimestamp = block.Timestamp + int64(ss.chainType.GetSnapshotGenerationTimeout().Seconds())
	)

	if block.Height <= constant.MinRollbackBlocks {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			fmt.Sprintf("invalid snapshot height: %d", block.Height))
	}
	// (safe) height to get snapshot's data from
	snapshotPayloadHeight := block.Height - constant.MinRollbackBlocks

	for qryRepoName, snapshotQuery := range ss.SnapshotQueries {
		func() {
			var (
				fromHeight uint32
				rows       *sql.Rows
			)
			if qryRepoName == "publishedReceipt" {
				if snapshotPayloadHeight > constant.LinkedReceiptBlocksLimit {
					fromHeight = snapshotPayloadHeight - constant.LinkedReceiptBlocksLimit
				}
			}
			qry := snapshotQuery.SelectDataForSnapshot(fromHeight, snapshotPayloadHeight)
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

	// encode and save snapshot payload to file/s
	snapshotFileHash, fileChunkHashes, err = ss.SnapshotBasicChunkStrategy.GenerateSnapshotChunks(snapshotPayload, ss.SnapshotPath)
	if err != nil {
		return nil, err
	}
	return &model.SnapshotFileInfo{
		SnapshotFileHash:           snapshotFileHash,
		FileChunksHashes:           fileChunkHashes,
		ChainType:                  ss.chainType.GetTypeInt(),
		Height:                     snapshotPayloadHeight,
		ProcessExpirationTimestamp: snapshotExpirationTimestamp,
		SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
	}, nil
}

// ImportSnapshotFile parses a downloaded snapshot file into db
func (ss *SnapshotMainBlockService) ImportSnapshotFile(snapshotFileInfo *model.SnapshotFileInfo) error {
	snapshotPayload, err := ss.SnapshotBasicChunkStrategy.BuildSnapshotFromChunks(snapshotFileInfo.GetSnapshotFileHash(),
		snapshotFileInfo.GetFileChunksHashes(), ss.SnapshotPath)
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
func (ss *SnapshotMainBlockService) InsertSnapshotPayloadToDb(payload *model.SnapshotPayload) error {
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
