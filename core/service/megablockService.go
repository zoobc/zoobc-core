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
		GetMegablockID(megablock *model.Megablock) (int64, error)
		GetMegablocksForSpineBlock(spineHeight uint32, spineTimestamp int64) ([]*model.Megablock, error)
		GetLastMegablock(ct chaintype.ChainType, mbType model.MegablockType) (*model.Megablock, error)
		CreateMegablock(fullFileHash []byte, megablockHeight uint32, expirationTimestamp int64, sortedFileChunksHashes [][]byte,
			ct chaintype.ChainType, mbType model.MegablockType) (*model.Megablock, error)
		GetMegablockBytes(megablock *model.Megablock) []byte
		InsertMegablock(megablock *model.Megablock) error
	}

	MegablockService struct {
		QueryExecutor   query.ExecutorInterface
		MegablockQuery  query.MegablockQueryInterface
		SpineBlockQuery query.BlockQueryInterface
		Logger          *log.Logger
	}
)

func NewMegablockService(
	queryExecutor query.ExecutorInterface,
	megablockQuery query.MegablockQueryInterface,
	spineBlockQuery query.BlockQueryInterface,
	logger *log.Logger,
) *MegablockService {
	return &MegablockService{
		QueryExecutor:   queryExecutor,
		MegablockQuery:  megablockQuery,
		SpineBlockQuery: spineBlockQuery,
		Logger:          logger,
	}
}

// GetMegablocksForSpineBlock retrieve all megablocks for a given spine height
// if there are no megablock at this height, return nil
// spineHeight height of the spine block we want to fetch the megablocks for
// spineTimestamp timestamp spine block we want to fetch the megablocks for
func (ss *MegablockService) GetMegablocksForSpineBlock(spineHeight uint32, spineTimestamp int64) ([]*model.Megablock, error) {
	var (
		megablocks     = make([]*model.Megablock, 0)
		prevSpineBlock model.Block
	)
	// genesis can never have megablocks
	if spineHeight == 0 {
		return megablocks, nil
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

	qry = ss.MegablockQuery.GetMegablocksInTimeInterval(prevSpineBlock.Timestamp, spineTimestamp)
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
// ct the megablock's chain type (eg. mainchain)
// ct the megablock's type (eg. snapshot)
func (ss *MegablockService) CreateMegablock(fullFileHash []byte, megablockHeight uint32,
	megablockTimestamp int64, sortedFileChunksHashes [][]byte, ct chaintype.ChainType, mbType model.MegablockType) (*model.Megablock,
	error) {
	var (
		megablockID         = int64(util.ConvertBytesToUint64(fullFileHash))
		megablockFileHashes = make([]byte, 0)
	)

	// build the megablock's payload (ordered sequence of file hashes been referenced by the megablock)
	for _, chunkHash := range sortedFileChunksHashes {
		megablockFileHashes = append(megablockFileHashes, chunkHash...)
	}

	// build the megablock
	megablock := &model.Megablock{
		// we store Megablock ID as little endian of fullFileHash so that we can join the megablock and FileChunks tables if needed
		FullFileHash:        fullFileHash,
		FileChunkHashes:     megablockFileHashes,
		MegablockHeight:     megablockHeight,
		ChainType:           ct.GetTypeInt(),
		MegablockType:       mbType,
		ExpirationTimestamp: megablockTimestamp,
	}
	megablockID, err := ss.GetMegablockID(megablock)
	if err != nil {
		return nil, err
	}
	megablock.ID = megablockID
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
	insertMegablockQ, insertMegablockArgs := ss.MegablockQuery.InsertMegablock(megablock)
	insertMegablockQry := append([]interface{}{insertMegablockQ}, insertMegablockArgs...)
	queries = append(queries, insertMegablockQry)
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
	buffer.Write(megablock.FileChunkHashes)
	buffer.Write(util.ConvertUint32ToBytes(megablock.MegablockHeight))
	buffer.Write(util.ConvertUint32ToBytes(uint32(megablock.ChainType)))
	buffer.Write(util.ConvertUint64ToBytes(uint64(megablock.ExpirationTimestamp)))
	return buffer.Bytes()
}

// GetMegablockID hash the megablock bytes and return its little endian representation
func (ss *MegablockService) GetMegablockID(megablock *model.Megablock) (int64, error) {
	digest := sha3.New256()
	_, err := digest.Write(ss.GetMegablockBytes(megablock))
	if err != nil {
		return -1, err
	}
	megablockHash := digest.Sum([]byte{})
	return int64(util.ConvertBytesToUint64(megablockHash)), nil

}
