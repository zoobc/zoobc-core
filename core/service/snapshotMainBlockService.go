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

	SnapshotMainBlockQueryServiceInterface interface {
		GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error)
		GetNodeRegistrations(fromHeight, toHeight uint32) ([]*model.NodeRegistration, error)
		GetAccountDatasets(fromHeight, toHeight uint32) ([]*model.AccountDataset, error)
		GetParticipationScores(fromHeight, toHeight uint32) ([]*model.ParticipationScore, error)
		GetPublishedReceipts(fromHeight, toHeight, limit uint32) ([]*model.PublishedReceipt, error)
		GetEscrowTransactions(fromHeight, toHeight uint32) ([]*model.Escrow, error)
		InsertSnapshotPayloadToDb(payload SnapshotPayload) error
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
		PublishedReceiptQuery     query.PublishedReceiptQueryInterface
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
		},
		FileService: fileService,
	}
}

// GetAccountBalances get account balances for snapshot (wrapper function around account balance query)
func (smbq *SnapshotMainBlockQueryService) GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error) {
	qry := query.GetSnapshotQuery(&chaintype.MainChain{})["accountBalance"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.AccountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetNodeRegistrations get node registrations for snapshot (wrapper function around node registration query)
func (smbq *SnapshotMainBlockQueryService) GetNodeRegistrations(fromHeight, toHeight uint32) ([]*model.NodeRegistration, error) {
	qry := query.GetSnapshotQuery(&chaintype.MainChain{})["nodeRegistration"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetAccountDatasets get account datasets  for snapshot (wrapper function around account dataset query)
func (smbq *SnapshotMainBlockQueryService) GetAccountDatasets(fromHeight, toHeight uint32) ([]*model.AccountDataset, error) {
	qry := query.GetSnapshotQuery(&chaintype.MainChain{})["accountDataset"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.AccountDatasetQuery.BuildModel([]*model.AccountDataset{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ParticipationScores get participation scores  for snapshot (wrapper function around participationscore query)
func (smbq *SnapshotMainBlockQueryService) GetParticipationScores(fromHeight, toHeight uint32) ([]*model.ParticipationScore, error) {
	qry := query.GetSnapshotQuery(&chaintype.MainChain{})["participationScore"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.ParticipationScoreQuery.BuildModel([]*model.ParticipationScore{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetPublishedReceipts get published Receipts for snapshot (wrapper function around published receipts query)
func (smbq *SnapshotMainBlockQueryService) GetPublishedReceipts(fromHeight, toHeight, limit uint32) ([]*model.PublishedReceipt, error) {
	// limit number of blocks to scan for receipts
	if toHeight-fromHeight > limit {
		fromHeight = toHeight - limit
	}
	qry := query.GetSnapshotQuery(&chaintype.MainChain{})["publishedReceipt"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.PublishedReceiptQuery.BuildModel([]*model.PublishedReceipt{}, rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetEscrowTransactions get escrowtransactions for snapshot (wrapper function around escrow transaction query)
func (smbq *SnapshotMainBlockQueryService) GetEscrowTransactions(fromHeight, toHeight uint32) ([]*model.Escrow, error) {
	qry := query.GetSnapshotQuery(&chaintype.MainChain{})["escrowTransaction"].SelectDataForSnapshot(fromHeight, toHeight)
	rows, err := smbq.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res, err := smbq.EscrowTransactionQuery.BuildModels(rows)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// InsertSnapshotPayloadToDb insert snapshot data to db
func (smbq *SnapshotMainBlockQueryService) InsertSnapshotPayloadToDb(payload SnapshotPayload) error {
	var (
		queries [][]interface{}
	)

	err := smbq.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}

	for _, rec := range payload.AccountBalances {
		qry, args := smbq.AccountBalanceQuery.InsertAccountBalance(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)

	}

	for _, rec := range payload.NodeRegistrations {
		qry, args := smbq.NodeRegistrationQuery.InsertNodeRegistration(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.PublishedReceipts {
		qry, args := smbq.PublishedReceiptQuery.InsertPublishedReceipt(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.ParticipationScores {
		qry, args := smbq.ParticipationScoreQuery.InsertParticipationScore(rec)
		queries = append(queries,
			append(
				[]interface{}{qry}, args...),
		)
	}

	for _, rec := range payload.EscrowTransactions {
		qryArgs := smbq.EscrowTransactionQuery.InsertEscrowTransaction(rec)
		queries = append(queries, qryArgs...)
	}

	for _, rec := range payload.AccountDatasets {
		qryArgs := smbq.AccountDatasetQuery.AddDataset(rec)
		queries = append(queries, qryArgs...)
	}

	err = smbq.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		rollbackErr := smbq.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			smbq.Logger.Error(rollbackErr.Error())
		}
		return blocker.NewBlocker(blocker.AppErr, "fail to insert snapshot into db")
	}
	err = smbq.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}

// NewSnapshotFile creates a new snapshot file (or multiple file chunks) and return the snapshotFileInfo
func (ss *SnapshotMainBlockService) NewSnapshotFile(block *model.Block, chunkSizeBytes int64) (*model.SnapshotFileInfo, error) {
	var (
		fileChunkHashes             = make([][]byte, 0)
		snapshotPayload             = new(SnapshotPayload)
		err                         error
		snapshotExpirationTimestamp = block.Timestamp + int64(ss.chainType.GetSnapshotGenerationTimeout().Seconds())
		// (safe) height to get snapshot's data from
		snapshotPayloadHeight = block.Height - constant.MinRollbackBlocks
	)

	if snapshotPayloadHeight <= 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr,
			fmt.Sprintf("invalid snapshot height: %d", snapshotPayloadHeight))
	}

	snapshotPayload.AccountBalances, err = ss.QueryService.GetAccountBalances(0, snapshotPayloadHeight)
	if err != nil {
		return nil, err
	}
	snapshotPayload.NodeRegistrations, err = ss.QueryService.GetNodeRegistrations(0, snapshotPayloadHeight)
	if err != nil {
		return nil, err
	}
	snapshotPayload.AccountDatasets, err = ss.QueryService.GetAccountDatasets(0, snapshotPayloadHeight)
	if err != nil {
		return nil, err
	}
	snapshotPayload.ParticipationScores, err = ss.QueryService.GetParticipationScores(0, snapshotPayloadHeight)
	if err != nil {
		return nil, err
	}
	snapshotPayload.PublishedReceipts, err = ss.QueryService.GetPublishedReceipts(0, snapshotPayloadHeight, constant.LinkedReceiptBlocksLimit)
	if err != nil {
		return nil, err
	}
	snapshotPayload.EscrowTransactions, err = ss.QueryService.GetEscrowTransactions(0, snapshotPayloadHeight)
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
		} else if height == constant.MinRollbackBlocks {
			return true
		}
		return (constant.MinRollbackBlocks+height)%snapshotInterval == 0
	}
	return height%snapshotInterval == 0

}
