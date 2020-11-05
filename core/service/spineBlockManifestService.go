package service

import (
	"bytes"
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	SpineBlockManifestServiceInterface interface {
		GetSpineBlockManifestID(spineBlockManifest *model.SpineBlockManifest) (int64, error)
		GetSpineBlockManifestsForSpineBlock(spineHeight uint32, spineTimestamp int64) ([]*model.SpineBlockManifest, error)
		GetLastSpineBlockManifest(ct chaintype.ChainType, mbType model.SpineBlockManifestType) (*model.SpineBlockManifest, error)
		CreateSpineBlockManifest(fullFileHash []byte, megablockHeight uint32, expirationTimestamp int64, sortedFileChunksHashes [][]byte,
			ct chaintype.ChainType, mbType model.SpineBlockManifestType) (*model.SpineBlockManifest, error)
		GetSpineBlockManifestBytes(spineBlockManifest *model.SpineBlockManifest) []byte
		InsertSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) error
		GetSpineBlockManifestBySpineBlockHeight(spineBlockHeight uint32) (
			[]*model.SpineBlockManifest, error,
		)
		GetSpineBlockManifestsFromSpineBlockHeight(spineBlockHeight uint32) (
			[]*model.SpineBlockManifest, error,
		)
		GetSpineBlockManifestsByManifestReferenceHeightRange(fromHeight, toHeight uint32) (manifests []*model.SpineBlockManifest, err error)
	}

	SpineBlockManifestService struct {
		QueryExecutor           query.ExecutorInterface
		SpineBlockManifestQuery query.SpineBlockManifestQueryInterface
		SpineBlockQuery         query.BlockQueryInterface
		Logger                  *log.Logger
	}
)

func NewSpineBlockManifestService(
	queryExecutor query.ExecutorInterface,
	megablockQuery query.SpineBlockManifestQueryInterface,
	spineBlockQuery query.BlockQueryInterface,
	logger *log.Logger,
) *SpineBlockManifestService {
	return &SpineBlockManifestService{
		QueryExecutor:           queryExecutor,
		SpineBlockManifestQuery: megablockQuery,
		SpineBlockQuery:         spineBlockQuery,
		Logger:                  logger,
	}
}

// GetSpineBlockManifestBySpineBlockHeight return all manifests published in spine block
func (ss *SpineBlockManifestService) GetSpineBlockManifestBySpineBlockHeight(spineBlockHeight uint32) (
	[]*model.SpineBlockManifest, error,
) {
	var (
		spineBlockManifests = make([]*model.SpineBlockManifest, 0)
	)
	qry := ss.SpineBlockManifestQuery.GetManifestBySpineBlockHeight(spineBlockHeight)
	rows, err := ss.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	spineBlockManifests, err = ss.SpineBlockManifestQuery.BuildModel(spineBlockManifests, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return spineBlockManifests, err
}

// GetSpineBlockManifestsFromSpineBlockHeight return all manifest where height > spineBlockHeight
func (ss *SpineBlockManifestService) GetSpineBlockManifestsFromSpineBlockHeight(spineBlockHeight uint32) (
	[]*model.SpineBlockManifest, error,
) {
	var (
		spineBlockManifests = make([]*model.SpineBlockManifest, 0)
	)
	qry := ss.SpineBlockManifestQuery.GetManifestsFromSpineBlockHeight(spineBlockHeight)
	rows, err := ss.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	spineBlockManifests, err = ss.SpineBlockManifestQuery.BuildModel(spineBlockManifests, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return spineBlockManifests, err
}

// GetSpineBlockManifestsForSpineBlock retrieve all spineBlockManifests for a given spine height
// if there are no spineBlockManifest at this height, return nil
// spineHeight height of the spine block we want to fetch the spineBlockManifests for
// spineTimestamp timestamp spine block we want to fetch the spineBlockManifests for
func (ss *SpineBlockManifestService) GetSpineBlockManifestsForSpineBlock(spineHeight uint32,
	spineTimestamp int64) ([]*model.SpineBlockManifest, error) {
	var (
		spineBlockManifests = make([]*model.SpineBlockManifest, 0)
		prevSpineBlock      model.Block
	)
	// genesis can never have spineBlockManifests
	if spineHeight == 0 {
		return spineBlockManifests, nil
	}

	qry := ss.SpineBlockQuery.GetBlockByHeight(spineHeight - 1)
	row, err := ss.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = ss.SpineBlockQuery.Scan(&prevSpineBlock, row)
	if err != nil {
		return nil, err
	}

	qry = ss.SpineBlockManifestQuery.GetSpineBlockManifestTimeInterval(prevSpineBlock.Timestamp, spineTimestamp)
	rows, err := ss.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	spineBlockManifests, err = ss.SpineBlockManifestQuery.BuildModel(spineBlockManifests, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	return spineBlockManifests, nil
}

// GetLastSpineBlockManifest retrieve the last available spineBlockManifest for the given chaintype
func (ss *SpineBlockManifestService) GetLastSpineBlockManifest(ct chaintype.ChainType,
	mbType model.SpineBlockManifestType) (*model.SpineBlockManifest, error) {
	var (
		spineBlockManifest model.SpineBlockManifest
	)
	qry := ss.SpineBlockManifestQuery.GetLastSpineBlockManifest(ct, mbType)
	row, err := ss.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = ss.SpineBlockManifestQuery.Scan(&spineBlockManifest, row)
	if err != nil {
		if blockErr, ok := err.(blocker.Blocker); ok && blockErr.Type != blocker.DBRowNotFound {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		// return nil if no spineBlockManifests are found
		return nil, nil
	}
	return &spineBlockManifest, nil
}

// CreateSpineBlockManifest persist a new spineBlockManifest
// fullFileHash: hash of the full (snapshot) file content
// megablockHeight: (mainchain) height at which the (snapshot) file computation has started (note: this is not the captured
// snapshot's height, which should be = mainHeight - minRollbackHeight)
// sortedFileChunksHashes all (snapshot) file chunks hashes for this spineBlockManifest (already sorted from first to last chunk)
// ct the spineBlockManifest's chain type (eg. mainchain)
// ct the spineBlockManifest's type (eg. snapshot)
func (ss *SpineBlockManifestService) CreateSpineBlockManifest(fullFileHash []byte, megablockHeight uint32,
	expirationTimestamp int64, sortedFileChunksHashes [][]byte, ct chaintype.ChainType,
	mbType model.SpineBlockManifestType) (*model.SpineBlockManifest,
	error) {
	var (
		megablockFileHashes = make([]byte, 0)
	)

	// build the spineBlockManifest's payload (ordered sequence of file hashes been referenced by the spineBlockManifest)
	for _, chunkHash := range sortedFileChunksHashes {
		megablockFileHashes = append(megablockFileHashes, chunkHash...)
	}

	// build the spineBlockManifest
	spineBlockManifest := &model.SpineBlockManifest{
		// we store SpineBlockManifest ID as little endian of fullFileHash so that we can join the spineBlockManifest and
		// FileChunks tables if needed
		FullFileHash:            fullFileHash,
		FileChunkHashes:         megablockFileHashes,
		ManifestReferenceHeight: megablockHeight,
		ChainType:               ct.GetTypeInt(),
		SpineBlockManifestType:  mbType,
		ExpirationTimestamp:     expirationTimestamp,
	}
	megablockID, err := ss.GetSpineBlockManifestID(spineBlockManifest)
	if err != nil {
		return nil, err
	}
	spineBlockManifest.ID = megablockID
	if err := ss.QueryExecutor.BeginTx(); err != nil {
		return nil, err
	}
	if err := ss.InsertSpineBlockManifest(spineBlockManifest); err != nil {
		if rollbackErr := ss.QueryExecutor.RollbackTx(); rollbackErr != nil {
			ss.Logger.Error(rollbackErr.Error())
		}
		return nil, err
	}
	err = ss.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}
	return spineBlockManifest, nil
}

// InsertSpineBlockManifest persist a spineBlockManifest to db (query wrapper)
func (ss *SpineBlockManifestService) InsertSpineBlockManifest(spineBlockManifest *model.SpineBlockManifest) error {
	var (
		queries = make([][]interface{}, 0)
	)
	insertSpineBlockManifestQ, insertSpineBlockManifestArgs := ss.SpineBlockManifestQuery.InsertSpineBlockManifest(spineBlockManifest)
	insertSpineBlockManifestQry := append([]interface{}{insertSpineBlockManifestQ}, insertSpineBlockManifestArgs...)
	queries = append(queries, insertSpineBlockManifestQry)
	err := ss.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

// GetSpineBlockManifestBytes translate tx body to bytes representation
func (ss *SpineBlockManifestService) GetSpineBlockManifestBytes(spineBlockManifest *model.SpineBlockManifest) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(spineBlockManifest.ID)))
	buffer.Write(spineBlockManifest.FullFileHash)
	// spineBlockManifest payload = all file chunks' entities bytes
	buffer.Write(spineBlockManifest.FileChunkHashes)
	buffer.Write(util.ConvertUint32ToBytes(spineBlockManifest.ManifestReferenceHeight))
	buffer.Write(util.ConvertUint32ToBytes(spineBlockManifest.ManifestSpineBlockHeight))
	buffer.Write(util.ConvertUint32ToBytes(uint32(spineBlockManifest.ChainType)))
	buffer.Write(util.ConvertUint64ToBytes(uint64(spineBlockManifest.ExpirationTimestamp)))
	return buffer.Bytes()
}

// GetSpineBlockManifestID hash the spineBlockManifest bytes and return its little endian representation
func (ss *SpineBlockManifestService) GetSpineBlockManifestID(spineBlockManifest *model.SpineBlockManifest) (int64, error) {
	digest := sha3.New256()
	_, err := digest.Write(ss.GetSpineBlockManifestBytes(spineBlockManifest))
	if err != nil {
		return -1, err
	}
	megablockHash := digest.Sum([]byte{})
	return int64(util.ConvertBytesToUint64(megablockHash)), nil

}

func (ss *SpineBlockManifestService) GetSpineBlockManifestsByManifestReferenceHeightRange(
	fromHeight, toHeight uint32,
) (manifests []*model.SpineBlockManifest, err error) {
	var (
		rows      *sql.Rows
		qry, args = ss.SpineBlockManifestQuery.GetManifestsFromManifestReferenceHeightRange(fromHeight, toHeight)
	)

	rows, err = ss.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	manifests, err = ss.SpineBlockManifestQuery.BuildModel(manifests, rows)
	if err != nil {
		return nil, err
	}

	return manifests, nil
}
