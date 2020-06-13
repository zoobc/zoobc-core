package service

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
)

type (
	SnapshotMainBlockService struct {
		SnapshotPath                   string
		chainType                      chaintype.ChainType
		TransactionUtil                transaction.UtilInterface
		TypeActionSwitcher             transaction.TypeActionSwitcher
		Logger                         *log.Logger
		SnapshotBasicChunkStrategy     SnapshotChunkStrategyInterface
		QueryExecutor                  query.ExecutorInterface
		AccountBalanceQuery            query.AccountBalanceQueryInterface
		NodeRegistrationQuery          query.NodeRegistrationQueryInterface
		ParticipationScoreQuery        query.ParticipationScoreQueryInterface
		AccountDatasetQuery            query.AccountDatasetQueryInterface
		EscrowTransactionQuery         query.EscrowTransactionQueryInterface
		PublishedReceiptQuery          query.PublishedReceiptQueryInterface
		PendingTransactionQuery        query.PendingTransactionQueryInterface
		PendingSignatureQuery          query.PendingSignatureQueryInterface
		MultisignatureInfoQuery        query.MultisignatureInfoQueryInterface
		MultiSignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		SkippedBlocksmithQuery         query.SkippedBlocksmithQueryInterface
		BlockQuery                     query.BlockQueryInterface
		SnapshotQueries                map[string]query.SnapshotQuery
		BlocksmithSafeQuery            map[string]bool
		DerivedQueries                 []query.DerivedQuery
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
	accountDatasetQuery query.AccountDatasetQueryInterface,
	escrowTransactionQuery query.EscrowTransactionQueryInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	pendingTransactionQuery query.PendingTransactionQueryInterface,
	pendingSignatureQuery query.PendingSignatureQueryInterface,
	multisignatureInfoQuery query.MultisignatureInfoQueryInterface,
	skippedBlocksmithQuery query.SkippedBlocksmithQueryInterface,
	blockQuery query.BlockQueryInterface,
	snapshotQueries map[string]query.SnapshotQuery,
	blocksmithSafeQueries map[string]bool,
	derivedQueries []query.DerivedQuery,
	transactionUtil transaction.UtilInterface,
	typeSwitcher transaction.TypeActionSwitcher,
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
		BlocksmithSafeQuery:        blocksmithSafeQueries,
		DerivedQueries:             derivedQueries,
		TransactionUtil:            transactionUtil,
		TypeActionSwitcher:         typeSwitcher,
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
			// if current query repo is blocksmith safe,
			// include more blocks to make sure we don't break smithing process due to missing data such as blocks,
			// published receipts and node registrations
			if ss.BlocksmithSafeQuery[qryRepoName] && snapshotPayloadHeight > constant.MinRollbackBlocks {
				fromHeight = snapshotPayloadHeight - constant.MinRollbackBlocks
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
	var (
		snapshotPayload *model.SnapshotPayload
		currentBlock    *model.Block
		err             error
	)

	snapshotPayload, err = ss.SnapshotBasicChunkStrategy.BuildSnapshotFromChunks(
		snapshotFileInfo.GetSnapshotFileHash(),
		snapshotFileInfo.GetFileChunksHashes(),
		ss.SnapshotPath,
	)
	if err != nil {
		return err
	}
	err = ss.InsertSnapshotPayloadToDB(snapshotPayload, snapshotFileInfo.Height)
	if err != nil {
		return err
	}

	ss.Logger.Infof("Need Re-ApplyUnconfirmed in %d pending transactions", len(snapshotPayload.GetPendingTransactions()))
	/*
		Need to manually ApplyUnconfirmed the pending transaction
		after finished insert snapshot payload into DB
	*/
	currentBlock, err = commonUtil.GetLastBlock(ss.QueryExecutor, ss.BlockQuery)
	if err != nil {
		return err
	}
	for _, pendingTX := range snapshotPayload.GetPendingTransactions() {
		var (
			innerTX *model.Transaction
			txType  transaction.TypeAction
		)
		if pendingTX.GetStatus() == model.PendingTransactionStatus_PendingTransactionPending {

			innerTX, err = ss.TransactionUtil.ParseTransactionBytes(pendingTX.GetTransactionBytes(), false)
			if err != nil {
				return err
			}

			innerTX.Height = currentBlock.GetHeight()
			txType, err = ss.TypeActionSwitcher.GetTransactionType(innerTX)
			if err != nil {
				return err
			}
			err = txType.ApplyUnconfirmed()
			if err != nil {
				return err
			}
		}
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
		queries      [][]interface{}
		highestBlock *model.Block
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
				if highestBlock == nil || highestBlock.Height < rec.Height {
					highestBlock = rec
				}
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
				qry, args := ss.AccountDatasetQuery.InsertAccountDataset(rec)
				queries = append(queries,
					append(
						[]interface{}{qry}, args...),
				)
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

				var participants = make([]*model.MultiSignatureParticipant, len(rec.GetAddresses()))
				for k, address := range rec.GetAddresses() {
					participants = append(participants, &model.MultiSignatureParticipant{
						MultiSignatureAddress: rec.GetMultisigAddress(),
						AccountAddress:        address,
						AccountAddressIndex:   uint32(k),
						BlockHeight:           rec.GetBlockHeight(),
						Latest:                rec.GetLatest(),
					})
				}
				participantQ := ss.MultiSignatureParticipantQuery.InsertMultisignatureParticipants(participants)
				queries = append(queries, participantQ...)
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
		queries = dQuery.Rollback(height)
		err = ss.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			ss.Logger.Errorf("Failed execute rollback queries in %d: %s", key, err.Error())
			rollbackErr := ss.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				ss.Logger.Warnf("Failed to run RollbackTX DB: %v", rollbackErr)
			}
			return err
		}
	}

	err = ss.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	monitoring.SetLastBlock(ss.chainType, highestBlock)
	return nil
}

// DeleteFileByChunkHashes delete the files included in the file chunk hashes.
func (ss *SnapshotMainBlockService) DeleteFileByChunkHashes(fileChunkHashes []byte) error {
	return ss.SnapshotBasicChunkStrategy.DeleteFileByChunkHashes(fileChunkHashes, ss.SnapshotPath)
}
