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
		PendingTransactionQuery    query.PendingTransactionQueryInterface
		PendingSignatureQuery      query.PendingSignatureQueryInterface
		MultisignatureInfoQuery    query.MultisignatureInfoQueryInterface
		SkippedBlocksmithQuery     query.SkippedBlocksmithQueryInterface
		BlockQuery                 query.BlockQueryInterface
		SnapshotQueries            map[string]query.SnapshotQuery
		DerivedQueries             []query.DerivedQuery
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
	pendingTransactionQuery query.PendingTransactionQueryInterface,
	pendingSignatureQuery query.PendingSignatureQueryInterface,
	multisignatureInfoQuery query.MultisignatureInfoQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	blockQuery query.BlockQueryInterface,
	snapshotQueries map[string]query.SnapshotQuery,
	derivedQueries []query.DerivedQuery,
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
		PendingTransactionQuery:    pendingTransactionQuery,
		PendingSignatureQuery:      pendingSignatureQuery,
		MultisignatureInfoQuery:    multisignatureInfoQuery,
		SkippedBlocksmithQuery:     skippedBlocksmithQuery,
		BlockQuery:                 blockQuery,
		SnapshotQueries:            snapshotQueries,
		DerivedQueries:             derivedQueries,
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

	// (safe) height to get snapshot's data from
	snapshotPayloadHeight := block.Height - constant.MinRollbackBlocks
	if block.Height <= constant.MinRollbackBlocks {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			fmt.Sprintf("invalid snapshot height: %d", int32(snapshotPayloadHeight)))
	}

	for qryRepoName, snapshotQuery := range ss.SnapshotQueries {
		func() {
			var (
				fromHeight uint32
				rows       *sql.Rows
			)
			if qryRepoName == "block" {
				if snapshotPayloadHeight > constant.MinRollbackBlocks {
					fromHeight = snapshotPayloadHeight - constant.MinRollbackBlocks
				}
			}
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
			case "block":
				snapshotPayload.Blocks, err = ss.BlockQuery.BuildModel([]*model.Block{}, rows)
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
			case "pendingTransaction":
				snapshotPayload.PendingTransactions, err = ss.PendingTransactionQuery.BuildModel([]*model.PendingTransaction{}, rows)
			case "pendingSignature":
				snapshotPayload.PendingSignatures, err = ss.PendingSignatureQuery.BuildModel([]*model.PendingSignature{}, rows)
			case "multisignatureInfo":
				snapshotPayload.MultiSignatureInfos, err = ss.MultisignatureInfoQuery.BuildModel([]*model.MultiSignatureInfo{}, rows)
			case "skippedBlocksmith":
				snapshotPayload.SkippedBlocksmiths, err = ss.SkippedBlocksmithQuery.BuildModel([]*model.SkippedBlocksmith{}, rows)
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
	err = ss.InsertSnapshotPayloadToDB(snapshotPayload, snapshotFileInfo.Height)
	if err != nil {
		return err
	}

	return nil
}

// IsSnapshotHeight returns true if chain height passed is a snapshot height
func (ss *SnapshotMainBlockService) IsSnapshotHeight(height uint32) bool {
	if height <= constant.MinRollbackBlocks {
		return false
	}
	snapshotInterval := ss.chainType.GetSnapshotInterval()
	return height%snapshotInterval == 0

}

// InsertSnapshotPayloadToDB insert snapshot data to db
func (ss *SnapshotMainBlockService) InsertSnapshotPayloadToDB(payload *model.SnapshotPayload, height uint32) error {
	var (
		queries [][]interface{}
	)

	err := ss.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}

	dummyArgs := make([]interface{}, 0)
	for qryRepoName, snapshotQuery := range ss.SnapshotQueries {
		qry := snapshotQuery.TrimDataBeforeSnapshot(0, height)
		queries = append(queries,
			append(
				[]interface{}{qry}, dummyArgs...),
		)

		switch qryRepoName {
		case "block":
			for _, rec := range payload.Blocks {
				qry, args := ss.BlockQuery.InsertBlock(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
			}
		case "accountBalance":
			for _, rec := range payload.AccountBalances {
				qry, args := ss.AccountBalanceQuery.InsertAccountBalance(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
			}
		case "nodeRegistration":
			for _, rec := range payload.NodeRegistrations {
				qry, args := ss.NodeRegistrationQuery.InsertNodeRegistration(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
			}
		case "accountDataset":
			for _, rec := range payload.AccountDatasets {
				qryArgs := ss.AccountDatasetQuery.AddDataset(rec)
				queries = append(queries, qryArgs...)
			}
		case "participationScore":
			for _, rec := range payload.ParticipationScores {
				qry, args := ss.ParticipationScoreQuery.InsertParticipationScore(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
			}
		case "publishedReceipt":
			for _, rec := range payload.PublishedReceipts {
				qry, args := ss.PublishedReceiptQuery.InsertPublishedReceipt(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
			}
		case "escrowTransaction":
			for _, rec := range payload.EscrowTransactions {
				qryArgs := ss.EscrowTransactionQuery.InsertEscrowTransaction(rec)
				queries = append(queries, qryArgs...)
			}
		case "pendingTransaction":
			for _, rec := range payload.PendingTransactions {
				qryArgs := ss.PendingTransactionQuery.InsertPendingTransaction(rec)
				queries = append(queries, qryArgs...)
			}
		case "pendingSignature":
			for _, rec := range payload.PendingSignatures {
				qryArgs := ss.PendingSignatureQuery.InsertPendingSignature(rec)
				queries = append(queries, qryArgs...)
			}
		case "multisignatureInfo":
			for _, rec := range payload.MultiSignatureInfos {
				qryArgs := ss.MultisignatureInfoQuery.InsertMultisignatureInfo(rec)
				queries = append(queries, qryArgs...)
			}
		case "skippedBlocksmith":
			for _, rec := range payload.SkippedBlocksmiths {
				qry, args := ss.SkippedBlocksmithQuery.InsertSkippedBlocksmith(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
			}
		default:
			return blocker.NewBlocker(blocker.ParserErr, fmt.Sprintf("Invalid Snapshot Query Repository: %s", qryRepoName))
		}
	}

	err = ss.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := ss.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			ss.Logger.Error(rollbackErr.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, fmt.Sprintf("fail to insert snapshot into db: %v", err))
	}

	for key, dQuery := range ss.DerivedQueries {
		queries := dQuery.Rollback(height)
		err = ss.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			fmt.Println(key)
			fmt.Println("Failed execute rollback queries, ", err.Error())
			err = ss.QueryExecutor.RollbackTx()
			if err != nil {
				fmt.Println("Failed to run RollbackTX DB")
			}
			break
		}
	}

	err = ss.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}
