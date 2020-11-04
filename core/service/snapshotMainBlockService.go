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
		MultisignatureParticipantQuery query.MultiSignatureParticipantQueryInterface
		SkippedBlocksmithQuery         query.SkippedBlocksmithQueryInterface
		BlockQuery                     query.BlockQueryInterface
		FeeScaleQuery                  query.FeeScaleQueryInterface
		FeeVoteCommitmentVoteQuery     query.FeeVoteCommitmentVoteQueryInterface
		FeeVoteRevealVoteQuery         query.FeeVoteRevealVoteQueryInterface
		LiquidPaymentTransactionQuery  query.LiquidPaymentTransactionQueryInterface
		NodeAdmissionTimestampQuery    query.NodeAdmissionTimestampQueryInterface
		SnapshotQueries                map[string]query.SnapshotQuery
		BlocksmithSafeQuery            map[string]bool
		DerivedQueries                 []query.DerivedQuery
		BlockMainService               BlockServiceInterface
		NodeRegistrationService        NodeRegistrationServiceInterface
		ScrambleNodeService            ScrambleNodeServiceInterface
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
	multisignatureParticipantQuery query.MultiSignatureParticipantQueryInterface,
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
	blockMainService BlockServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
) *SnapshotMainBlockService {
	return &SnapshotMainBlockService{
		SnapshotPath:                   snapshotPath,
		chainType:                      &chaintype.MainChain{},
		Logger:                         logger,
		SnapshotBasicChunkStrategy:     snapshotChunkStrategy,
		QueryExecutor:                  queryExecutor,
		AccountBalanceQuery:            accountBalanceQuery,
		NodeRegistrationQuery:          nodeRegistrationQuery,
		AccountDatasetQuery:            accountDatasetQuery,
		ParticipationScoreQuery:        participationScoreQuery,
		EscrowTransactionQuery:         escrowTransactionQuery,
		PublishedReceiptQuery:          publishedReceiptQuery,
		PendingTransactionQuery:        pendingTransactionQuery,
		PendingSignatureQuery:          pendingSignatureQuery,
		MultisignatureInfoQuery:        multisignatureInfoQuery,
		MultisignatureParticipantQuery: multisignatureParticipantQuery,
		SkippedBlocksmithQuery:         skippedBlocksmithQuery,
		FeeScaleQuery:                  feeScaleQuery,
		FeeVoteCommitmentVoteQuery:     feeVoteCommitmentVoteQuery,
		FeeVoteRevealVoteQuery:         feeVoteRevealVoteQuery,
		LiquidPaymentTransactionQuery:  liquidPaymentTransactionQuery,
		NodeAdmissionTimestampQuery:    nodeAdmissionTimestampQuery,
		BlockQuery:                     blockQuery,
		SnapshotQueries:                snapshotQueries,
		BlocksmithSafeQuery:            blocksmithSafeQueries,
		DerivedQueries:                 derivedQueries,
		TransactionUtil:                transactionUtil,
		TypeActionSwitcher:             typeSwitcher,
		BlockMainService:               blockMainService,
		NodeRegistrationService:        nodeRegistrationService,
		ScrambleNodeService:            scrambleNodeService,
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
				fromHeight    uint32
				rows          *sql.Rows
				multisigInfos []*model.MultiSignatureInfo
			)
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
				multisigInfos, err = ss.MultisignatureInfoQuery.BuildModel([]*model.MultiSignatureInfo{}, rows)
				for idx, multisigInfo := range multisigInfos {
					err = func(idx int, multisigInfos []*model.MultiSignatureInfo) error {
						qry, args := ss.MultisignatureParticipantQuery.GetMultiSignatureParticipantsByMultisigAddressAndHeightRange(
							multisigInfo.GetMultisigAddress(),
							fromHeight,
							snapshotPayloadHeight,
						)
						rows2, err := ss.QueryExecutor.ExecuteSelect(qry, false, args...)
						if err != nil {
							return err
						}
						defer rows2.Close()
						participants, err := ss.MultisignatureParticipantQuery.BuildModel(rows2)
						if err != nil {
							return err
						}
						for _, participant := range participants {
							multisigInfos[idx].Addresses = append(multisigInfos[idx].Addresses, participant.GetAccountAddress())
						}
						return nil
					}(idx, multisigInfos)
					if err != nil {
						return
					}
				}
				snapshotPayload.MultiSignatureInfos = multisigInfos
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
	snapshotFileHash, fileChunkHashes, err = ss.SnapshotBasicChunkStrategy.GenerateSnapshotChunks(snapshotPayload)
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
	currentBlock, err = ss.BlockMainService.GetLastBlock()
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
	// update or clear all cache storage
	err = ss.ScrambleNodeService.InitializeScrambleCache(currentBlock.GetHeight())
	if err != nil {
		return err
	}
	err = ss.NodeRegistrationService.InitializeCache()
	if err != nil {
		return err
	}
	err = ss.NodeRegistrationService.UpdateNextNodeAdmissionCache(nil)
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
		queries      [][]interface{}
		highestBlock *model.Block
	)

	for qryRepoName, snapshotQuery := range ss.SnapshotQueries {
		var (
			qry = snapshotQuery.TrimDataBeforeSnapshot(0, height)
		)
		queries = append(queries, []interface{}{qry})

		switch qryRepoName {
		case "block":
			if len(payload.GetBlocks()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetBlocks())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "accountBalance":
			if len(payload.GetAccountBalances()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetAccountBalances())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "nodeRegistration":
			if len(payload.GetNodeRegistrations()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetNodeRegistrations())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "accountDataset":
			if len(payload.GetAccountDatasets()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetAccountDatasets())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "participationScore":
			if len(payload.GetParticipationScores()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetParticipationScores())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "publishedReceipt":
			if len(payload.GetPublishedReceipts()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetPublishedReceipts())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "escrowTransaction":
			if len(payload.GetEscrowTransactions()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetEscrowTransactions())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "pendingTransaction":
			if len(payload.GetPendingTransactions()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetPendingTransactions())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "pendingSignature":
			if len(payload.GetPendingSignatures()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetPendingSignatures())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}

		case "multisignatureInfo":
			if len(payload.GetMultiSignatureInfos()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetMultiSignatureInfos())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "skippedBlocksmith":
			if len(payload.GetSkippedBlocksmiths()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetSkippedBlocksmiths())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "feeScale":
			if len(payload.GetFeeScale()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetFeeScale())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "feeVoteCommit":
			if len(payload.GetFeeVoteCommitmentVote()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetFeeVoteCommitmentVote())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "feeVoteReveal":
			if len(payload.GetFeeVoteRevealVote()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetFeeVoteRevealVote())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "liquidPaymentTransaction":
			if len(payload.GetLiquidPayment()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetLiquidPayment())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		case "nodeAdmissionTimestamp":
			if len(payload.GetNodeAdmissionTimestamp()) > 0 {
				q, err := snapshotQuery.ImportSnapshot(payload.GetNodeAdmissionTimestamp())
				if err != nil {
					return err
				}
				queries = append(queries, q...)
			}
		default:
			return blocker.NewBlocker(blocker.ParserErr, fmt.Sprintf("Invalid Snapshot Query Repository: %s", qryRepoName))
		}
		// recalibrate the versioned table to get rid of multiple `latest = true` rows.
		recalibrateQuery := snapshotQuery.RecalibrateVersionedTable()
		if len(recalibrateQuery) > 0 {
			for _, s := range recalibrateQuery {
				queries = append(queries, []interface{}{s})
			}
		}
	}
	err := ss.QueryExecutor.BeginTx()
	if err != nil {
		return err
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

	// update or clear all cache storage
	err = ss.BlockMainService.UpdateLastBlockCache(nil)
	if err != nil {
		return err
	}
	err = ss.NodeRegistrationService.UpdateNextNodeAdmissionCache(nil)
	if err != nil {
		return err
	}

	err = ss.BlockMainService.InitializeBlocksCache()
	if err != nil {
		return err
	}
	highestBlock, err = ss.BlockMainService.GetLastBlock()
	if err != nil {
		return err
	}
	err = ss.ScrambleNodeService.InitializeScrambleCache(highestBlock.GetHeight())
	if err != nil {
		return err
	}
	err = ss.NodeRegistrationService.InitializeCache()
	if err != nil {
		return err
	}
	monitoring.SetLastBlock(ss.chainType, highestBlock)
	return nil
}

// DeleteFileByChunkHashes delete the files included in the file chunk hashes.
func (ss *SnapshotMainBlockService) DeleteFileByChunkHashes(fileChunkHashes []byte) error {
	return ss.SnapshotBasicChunkStrategy.DeleteFileByChunkHashes(fileChunkHashes)
}
