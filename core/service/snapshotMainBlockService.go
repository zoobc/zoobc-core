package service

import (
	"database/sql"
	"fmt"
	"math"

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
		SnapshotPath                  string
		chainType                     chaintype.ChainType
		TransactionUtil               transaction.UtilInterface
		TypeActionSwitcher            transaction.TypeActionSwitcher
		Logger                        *log.Logger
		SnapshotBasicChunkStrategy    SnapshotChunkStrategyInterface
		QueryExecutor                 query.ExecutorInterface
		AccountBalanceQuery           query.AccountBalanceQueryInterface
		NodeRegistrationQuery         query.NodeRegistrationQueryInterface
		ParticipationScoreQuery       query.ParticipationScoreQueryInterface
		AccountDatasetQuery           query.AccountDatasetQueryInterface
		EscrowTransactionQuery        query.EscrowTransactionQueryInterface
		PublishedReceiptQuery         query.PublishedReceiptQueryInterface
		PendingTransactionQuery       query.PendingTransactionQueryInterface
		PendingSignatureQuery         query.PendingSignatureQueryInterface
		MultisignatureInfoQuery       query.MultisignatureInfoQueryInterface
		SkippedBlocksmithQuery        query.SkippedBlocksmithQueryInterface
		BlockQuery                    query.BlockQueryInterface
		FeeScaleQuery                 query.FeeScaleQueryInterface
		FeeVoteCommitmentVoteQuery    query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery        query.FeeVoteRevealVoteQueryInterface
		LiquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface
		NodeAdmissionTimestampQuery   query.NodeAdmissionTimestampQueryInterface
		SnapshotQueries               map[string]query.SnapshotQuery
		BlocksmithSafeQuery           map[string]bool
		DerivedQueries                []query.DerivedQuery
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
	feeScaleQuery query.FeeScaleQueryInterface,
	feeVoteCommitmentVoteQuery query.FeeVoteCommitmentVoteQueryInterface,
	feeVoteRevealVoteQuery query.FeeVoteRevealVoteQueryInterface,
	liquidPaymentTransactionQuery query.LiquidPaymentTransactionQueryInterface,
	nodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface,
	blockQuery query.BlockQueryInterface,
	snapshotQueries map[string]query.SnapshotQuery,
	blocksmithSafeQueries map[string]bool,
	derivedQueries []query.DerivedQuery,
	transactionUtil transaction.UtilInterface,
	typeSwitcher transaction.TypeActionSwitcher,
) *SnapshotMainBlockService {
	return &SnapshotMainBlockService{
		SnapshotPath:                  snapshotPath,
		chainType:                     &chaintype.MainChain{},
		Logger:                        logger,
		SnapshotBasicChunkStrategy:    snapshotChunkStrategy,
		QueryExecutor:                 queryExecutor,
		AccountBalanceQuery:           accountBalanceQuery,
		NodeRegistrationQuery:         nodeRegistrationQuery,
		AccountDatasetQuery:           accountDatasetQuery,
		ParticipationScoreQuery:       participationScoreQuery,
		EscrowTransactionQuery:        escrowTransactionQuery,
		PublishedReceiptQuery:         publishedReceiptQuery,
		PendingTransactionQuery:       pendingTransactionQuery,
		PendingSignatureQuery:         pendingSignatureQuery,
		MultisignatureInfoQuery:       multisignatureInfoQuery,
		SkippedBlocksmithQuery:        skippedBlocksmithQuery,
		FeeScaleQuery:                 feeScaleQuery,
		FeeVoteCommitmentVoteQuery:    feeVoteCommitmentVoteQuery,
		FeeVoteRevealVoteQuery:        feeVoteRevealVoteQuery,
		LiquidPaymentTransactionQuery: liquidPaymentTransactionQuery,
		NodeAdmissionTimestampQuery:   nodeAdmissionTimestampQuery,
		BlockQuery:                    blockQuery,
		SnapshotQueries:               snapshotQueries,
		BlocksmithSafeQuery:           blocksmithSafeQueries,
		DerivedQueries:                derivedQueries,
		TransactionUtil:               transactionUtil,
		TypeActionSwitcher:            typeSwitcher,
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
			case "feeScale":
				snapshotPayload.FeeScale, err = ss.FeeScaleQuery.BuildModel([]*model.FeeScale{}, rows)
			case "feeVoteCommit":
				snapshotPayload.FeeVoteCommitmentVote, err = ss.FeeVoteCommitmentVoteQuery.BuildModel([]*model.FeeVoteCommitmentVote{}, rows)
			case "feeVoteReveal":
				snapshotPayload.FeeVoteRevealVote, err = ss.FeeVoteRevealVoteQuery.BuildModel([]*model.FeeVoteRevealVote{}, rows)
			case "liquidPaymentTransaction":
				snapshotPayload.LiquidPayment, err = ss.LiquidPaymentTransactionQuery.BuildModels(rows)
			case "nodeAdmissionTimestamp":
				snapshotPayload.NodeAdmissionTimestamp, err = ss.NodeAdmissionTimestampQuery.BuildModel([]*model.NodeAdmissionTimestamp{}, rows)
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

// calculateBulkSize calculating max records might allowed in single sqlite transaction, since sqlite3 has maximum
// variables in single transactions called SQLITE_LIMIT_VARIABLE_NUMBER in sqlite3-binding.c which is 999
func (ss *SnapshotMainBlockService) calculateBulkSize(totalFields, totalRecords int) (recordsPerPeriod, rounds, remaining int) {

	perPeriod := math.Floor(999 / float64(totalFields))
	rounds = int(math.Floor(float64(totalRecords) / perPeriod))

	if perPeriod == 0 || rounds == 0 {
		return totalRecords, 1, 0
	}
	remaining = totalRecords % (rounds * int(perPeriod))
	return int(perPeriod), rounds, remaining
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

	for qryRepoName, snapshotQuery := range ss.SnapshotQueries {
		var (
			qry                                 = snapshotQuery.TrimDataBeforeSnapshot(0, height)
			args                                []interface{}
			recordsPerPeriod, rounds, remaining int
		)
		queries = append(queries, []interface{}{qry})

		switch qryRepoName {
		case "block":
			if len(payload.GetBlocks()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.BlockQuery.GetFields()), len(payload.GetBlocks()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.BlockQuery.InsertBlocks(payload.GetBlocks()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.BlockQuery.InsertBlocks(payload.GetBlocks()[len(payload.GetBlocks())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "accountBalance":
			if len(payload.GetAccountBalances()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.AccountBalanceQuery.GetFields()), len(payload.GetAccountBalances()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.AccountBalanceQuery.InsertAccountBalances(
						payload.GetAccountBalances()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.AccountBalanceQuery.InsertAccountBalances(payload.GetAccountBalances()[len(payload.GetAccountBalances())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "nodeRegistration":
			if len(payload.GetNodeRegistrations()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.NodeRegistrationQuery.GetFields()), len(payload.GetNodeRegistrations()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.NodeRegistrationQuery.InsertNodeRegistrations(
						payload.GetNodeRegistrations()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.NodeRegistrationQuery.InsertNodeRegistrations(payload.GetNodeRegistrations()[len(payload.GetNodeRegistrations())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "accountDataset":
			if len(payload.GetAccountDatasets()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.AccountDatasetQuery.GetFields()), len(payload.GetAccountDatasets()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.AccountDatasetQuery.InsertAccountDatasets(
						payload.GetAccountDatasets()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.AccountDatasetQuery.InsertAccountDatasets(payload.GetAccountDatasets()[len(payload.GetAccountDatasets())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "participationScore":
			if len(payload.GetParticipationScores()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.ParticipationScoreQuery.GetFields()), len(payload.GetParticipationScores()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.ParticipationScoreQuery.InsertParticipationScores(
						payload.GetParticipationScores()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.ParticipationScoreQuery.InsertParticipationScores(
						payload.GetParticipationScores()[len(payload.GetParticipationScores())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "publishedReceipt":
			if len(payload.GetPublishedReceipts()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.PublishedReceiptQuery.GetFields()), len(payload.GetPublishedReceipts()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.PublishedReceiptQuery.InsertPublishedReceipts(
						payload.GetPublishedReceipts()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.PublishedReceiptQuery.InsertPublishedReceipts(payload.GetPublishedReceipts()[len(payload.GetPublishedReceipts())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "escrowTransaction":
			if len(payload.GetEscrowTransactions()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.EscrowTransactionQuery.GetFields()), len(payload.GetEscrowTransactions()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.EscrowTransactionQuery.InsertEscrowTransactions(
						payload.GetEscrowTransactions()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.EscrowTransactionQuery.InsertEscrowTransactions(payload.GetEscrowTransactions()[len(payload.GetEscrowTransactions())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "pendingTransaction":
			if len(payload.GetPendingTransactions()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.PendingSignatureQuery.GetFields()), len(payload.GetPendingTransactions()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.PendingTransactionQuery.InsertPendingTransactions(
						payload.GetPendingTransactions()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.PendingTransactionQuery.InsertPendingTransactions(
						payload.GetPendingTransactions()[len(payload.GetPendingTransactions())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "pendingSignature":
			if len(payload.GetPendingSignatures()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.PendingSignatureQuery.GetFields()), len(payload.GetPendingSignatures()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.PendingSignatureQuery.InsertPendingSignatures(
						payload.GetPendingSignatures()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.PendingSignatureQuery.InsertPendingSignatures(
						payload.GetPendingSignatures()[len(payload.GetPendingSignatures())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}

		case "multisignatureInfo":
			if len(payload.GetMultiSignatureInfos()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.MultisignatureInfoQuery.GetFields()), len(payload.GetMultiSignatureInfos()))
				for i := 0; i < rounds; i++ {
					var ii = i
					musigQ := ss.MultisignatureInfoQuery.InsertMultiSignatureInfos(
						payload.GetMultiSignatureInfos()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, musigQ...)
				}
				if remaining > 0 {
					musigQ := ss.MultisignatureInfoQuery.InsertMultiSignatureInfos(
						payload.GetMultiSignatureInfos()[len(payload.GetMultiSignatureInfos())-remaining:],
					)
					queries = append(queries, musigQ...)
				}
			}
		case "skippedBlocksmith":
			if len(payload.GetSkippedBlocksmiths()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.SkippedBlocksmithQuery.GetFields()), len(payload.GetSkippedBlocksmiths()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.SkippedBlocksmithQuery.InsertSkippedBlocksmiths(
						payload.GetSkippedBlocksmiths()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.SkippedBlocksmithQuery.InsertSkippedBlocksmiths(
						payload.GetSkippedBlocksmiths()[len(payload.GetSkippedBlocksmiths())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}
		case "feeScale":
			if len(payload.GetFeeScale()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.FeeScaleQuery.GetFields()), len(payload.GetFeeScale()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.FeeScaleQuery.InsertFeeScales(
						payload.GetFeeScale()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.FeeScaleQuery.InsertFeeScales(payload.GetFeeScale()[len(payload.GetFeeScale())-remaining:])
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}
		case "feeVoteCommit":
			if len(payload.GetFeeVoteCommitmentVote()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.FeeVoteCommitmentVoteQuery.GetFields()), len(payload.GetFeeScale()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.FeeVoteCommitmentVoteQuery.InsertCommitVotes(
						payload.GetFeeVoteCommitmentVote()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.FeeVoteCommitmentVoteQuery.InsertCommitVotes(
						payload.GetFeeVoteCommitmentVote()[len(payload.GetFeeVoteCommitmentVote())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}
		case "feeVoteReveal":
			if len(payload.GetFeeVoteRevealVote()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.FeeVoteRevealVoteQuery.GetFields()), len(payload.GetFeeVoteRevealVote()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.FeeVoteRevealVoteQuery.InsertRevealVotes(
						payload.GetFeeVoteRevealVote()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.FeeVoteRevealVoteQuery.InsertRevealVotes(
						payload.GetFeeVoteRevealVote()[len(payload.GetFeeVoteRevealVote())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}
		case "liquidPaymentTransaction":
			if len(payload.GetLiquidPayment()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(len(ss.LiquidPaymentTransactionQuery.GetFields()), len(payload.GetLiquidPayment()))
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.LiquidPaymentTransactionQuery.InsertLiquidPaymentTransactions(
						payload.GetLiquidPayment()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.LiquidPaymentTransactionQuery.InsertLiquidPaymentTransactions(
						payload.GetLiquidPayment()[len(payload.GetLiquidPayment())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
			}
		case "nodeAdmissionTimestamp":
			if len(payload.GetNodeAdmissionTimestamp()) > 0 {
				recordsPerPeriod, rounds, remaining = ss.calculateBulkSize(
					len(ss.NodeAdmissionTimestampQuery.GetFields()),
					len(payload.GetNodeAdmissionTimestamp()),
				)
				for i := 0; i < rounds; i++ {
					var ii = i
					qry, args = ss.NodeAdmissionTimestampQuery.InsertNextNodeAdmissions(
						payload.GetNodeAdmissionTimestamp()[ii*recordsPerPeriod : (ii*recordsPerPeriod)+recordsPerPeriod],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
				if remaining > 0 {
					qry, args = ss.NodeAdmissionTimestampQuery.InsertNextNodeAdmissions(
						payload.GetNodeAdmissionTimestamp()[len(payload.GetNodeAdmissionTimestamp())-remaining:],
					)
					queries = append(queries, append([]interface{}{qry}, args...))
				}
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
