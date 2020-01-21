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
	MegablockServiceInterface interface {
		GetMegablocksFromSpineHeight(spineHeight uint32) ([]*model.Megablock, error)
		GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) (*model.Megablock, error)
		CreateMegablock(fileFullHash []byte, megablockHeight, spineHeight uint32,
			sortedFileChunksHashes [][]byte, lastFileChunkHash []byte, ct chaintype.ChainType,
			mbType model.MegablockType) (*model.Megablock, error)
		GetMegablockBytes(megablock *model.Megablock) []byte
		GetFileChunkBytes(snapshotChunk *model.FileChunk) []byte
		InsertMegablock(megablock *model.Megablock) error
	}

	MegablockService struct {
		QueryExecutor  query.ExecutorInterface
		MegablockQuery query.MegablockQueryInterface
		FileChunkQuery query.FileChunkQueryInterface
		Logger         *log.Logger
	}
)

func NewMegablockService(
	queryExecutor query.ExecutorInterface,
	megablockQuery query.MegablockQueryInterface,
	snapshotChunkQuery query.FileChunkQueryInterface,
	logger *log.Logger,
) *MegablockService {
	return &MegablockService{
		QueryExecutor:  queryExecutor,
		MegablockQuery: megablockQuery,
		FileChunkQuery: snapshotChunkQuery,
		Logger:         logger,
	}
}

// GetMegablocksFromSpineHeight retrieve all megablocks for a given spine height and chain type
// if there is no megablock at this height, return nil
func (ss *MegablockService) GetMegablocksFromSpineHeight(spineHeight uint32) ([]*model.Megablock, error) {
	var (
		megablocks     []*model.Megablock
		snapshotChunks []*model.FileChunk
	)
	qry := ss.MegablockQuery.GetMegablocksBySpineBlockHeight(spineHeight)
	rows, err := ss.QueryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	megablocks, err = ss.MegablockQuery.BuildModel(megablocks, rows)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		return nil, nil
	}

	// populate snapshotChunks
	for _, megablock := range megablocks {
		sqlStr := ss.FileChunkQuery.GetFileChunksByMegablockID(megablock.ID)
		rows, err := ss.QueryExecutor.ExecuteSelect(sqlStr, false)
		if err != nil {
			return nil, err
		}
		rows.Close()
		snapshotChunks, err = ss.FileChunkQuery.BuildModel(snapshotChunks, rows)
		if err != nil {
			return nil, err
		}
		megablock.FileChunks = snapshotChunks
	}
	return megablocks, nil
}

// GetLastMegablock retrieve the last available megablock for the given chaintype
func (ss *MegablockService) GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) (*model.Megablock, error) {
	var (
		megablock model.Megablock
	)
	qry := ss.MegablockQuery.GetLastMegablock(ct, mbType)
	row, err := ss.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = ss.MegablockQuery.Scan(&megablock, row)
	if err != nil {
		if blockErr, ok := err.(blocker.Blocker); ok && blockErr.Type != blocker.DBRowNotFound {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		// return nil if no megablocks are found
		return nil, nil
	}
	return &megablock, nil
}

// CreateMegablock persist a new megablock
// fullFileHash: hash of the full (snapshot) file content
// megablockHeight: (mainchain) height at which the (snapshot) file computation has started (note: this is not the captured
// snapshot's height, which should be = mainHeight - minRollbackHeight)
// sortedFileChunksHashes all (snapshot) file chunks hashes for this megablock (already sorted from first to last chunk)
// lastFileChunkHash last available (snapshot) file chunk hash (from db)
// ct the megablock's chain type
func (ss *MegablockService) CreateMegablock(fullFileHash []byte, megablockHeight, spineHeight uint32,
	sortedFileChunksHashes [][]byte, lastFileChunkHash []byte, ct chaintype.ChainType, mbType model.MegablockType) (*model.Megablock, error) {
	var (
		previousChunkHash, fileChunksBytes, fileChunksHash []byte
		sortedFileChunks                                   = make([]*model.FileChunk, 0)
		megablockID                                        = int64(util.ConvertBytesToUint64(fullFileHash))
	)

	// build the snapshot chunks
	previousChunkHash = lastFileChunkHash
	for idx, chunkHash := range sortedFileChunksHashes {
		fileChunk := &model.FileChunk{
			MegablockID:       megablockID,
			ChunkHash:         chunkHash,
			ChunkIndex:        uint32(idx),
			PreviousChunkHash: previousChunkHash,
			SpineBlockHeight:  spineHeight,
		}
		sortedFileChunks = append(sortedFileChunks, fileChunk)
		fileChunksBytes = append(fileChunksBytes, ss.GetFileChunkBytes(fileChunk)...)
		previousChunkHash = chunkHash
	}
	digest := sha3.New512()
	_, err := digest.Write(fileChunksBytes)
	if err != nil {
		return nil, err
	}
	fileChunksHash = digest.Sum([]byte{})

	// build the megablock
	megablock := &model.Megablock{
		// we store Megablock ID as little endian of fullFileHash so that we can join the megablock and FileChunks tables if needed
		ID:                     megablockID,
		FullFileHash:           fullFileHash,
		MegablockHeight:        megablockHeight,
		SpineBlockHeight:       spineHeight,
		MegablockPayloadLength: uint32(len(fileChunksBytes)),
		MegablockPayloadHash:   fileChunksHash,
		ChainType:              ct.GetTypeInt(),
		MegablockType:          mbType,
		FileChunks:             sortedFileChunks,
	}
	if err := ss.QueryExecutor.BeginTx(); err != nil {
		return nil, err
	}
	if err := ss.InsertMegablock(megablock); err != nil {
		if rollbackErr := ss.QueryExecutor.RollbackTx(); rollbackErr != nil {
			ss.Logger.Error(rollbackErr.Error())
		}
		return nil, err
	}
	err = ss.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}
	return megablock, nil
}

// InsertMegablock persist a megablock to db (query wrapper)
func (ss *MegablockService) InsertMegablock(megablock *model.Megablock) error {
	var (
		queries = make([][]interface{}, 0)
	)
	if megablock.FileChunks == nil {
		return blocker.NewBlocker(blocker.AppErr, "FileChunksNil")
	}
	insertMegablockQ, insertMegablockArgs := ss.MegablockQuery.InsertMegablock(megablock)
	insertMegablockQry := append([]interface{}{insertMegablockQ}, insertMegablockArgs...)
	queries = append(queries, insertMegablockQry)

	for _, snapshotChunk := range megablock.FileChunks {
		// add chunk to db transaction
		insertFileChunkQ, insertFileChunkArgs := ss.FileChunkQuery.InsertFileChunk(snapshotChunk)
		insertFileChunkQry := append([]interface{}{insertFileChunkQ}, insertFileChunkArgs...)
		queries = append(queries, insertFileChunkQry)
	}
	err := ss.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

// GetBodyBytes translate tx body to bytes representation
func (ss *MegablockService) GetMegablockBytes(megablock *model.Megablock) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(uint64(megablock.ID)))
	buffer.Write(megablock.FullFileHash)
	// megablock payload = all file chunks' entities bytes
	buffer.Write(util.ConvertUint32ToBytes(megablock.MegablockPayloadLength))
	buffer.Write(megablock.MegablockPayloadHash)
	buffer.Write(util.ConvertUint32ToBytes(megablock.SpineBlockHeight))
	buffer.Write(util.ConvertUint32ToBytes(megablock.MegablockHeight))
	buffer.Write(util.ConvertUint32ToBytes(uint32(megablock.ChainType)))
	return buffer.Bytes()
}

func (ss *MegablockService) GetFileChunkBytes(snapshotChunk *model.FileChunk) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(snapshotChunk.ChunkHash)
	buffer.Write(util.ConvertUint64ToBytes(uint64(snapshotChunk.MegablockID)))
	buffer.Write(util.ConvertUint32ToBytes(snapshotChunk.ChunkIndex))
	buffer.Write(snapshotChunk.PreviousChunkHash)
	buffer.Write(util.ConvertUint32ToBytes(snapshotChunk.SpineBlockHeight))
	buffer.Write(util.ConvertUint32ToBytes(uint32(snapshotChunk.ChainType)))
	return buffer.Bytes()
}
