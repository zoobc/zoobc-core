package service

import (
	"bytes"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	SnapshotServiceInterface interface {
		GetMegablockFromSpineHeight(spineHeight uint32) (*model.Megablock, error)
		GetLastMegablock() (*model.Megablock, error)
		CreateMegablock(snapshotHash []byte, mainHeight, spineHeight uint32,
			sortedSnapshotChunksHashes [][]byte, lastSnapshotChunkHash []byte) (*model.Megablock, error)
		GetNextSnapshotHeight(mainHeight uint32) uint32
		GenerateSnapshot(mainHeight uint32) (*model.Megablock, error)
		GetMegablockBytes(megablock *model.Megablock) ([]byte, error)
		GetSnapshotChunkBytes(snapshotChunk *model.SnapshotChunk) []byte
		StartSnapshotListener() observer.Listener
		InsertMegablock(megablock *model.Megablock) error
	}

	SnapshotService struct {
		QueryExecutor      query.ExecutorInterface
		MegablockQuery     query.MegablockQueryInterface
		SpineBlockQuery    query.BlockQueryInterface
		MainBlockQuery     query.BlockQueryInterface
		SnapshotChunkQuery query.SnapshotChunkQueryInterface
		Logger             *log.Logger
		// below fields are for better code testability
		Spinechain                chaintype.ChainType
		Mainchain                 chaintype.ChainType
		SnapshotInterval          int64
		SnapshotGenerationTimeout int64
	}
)

func NewSnapshotService(
	queryExecutor query.ExecutorInterface,
	mainBlockQuery, spineBlockQuery query.BlockQueryInterface,
	megablockQuery query.MegablockQueryInterface,
	snapshotChunkQuery query.SnapshotChunkQueryInterface,
	logger *log.Logger,
) *SnapshotService {
	return &SnapshotService{
		QueryExecutor:             queryExecutor,
		MegablockQuery:            megablockQuery,
		SpineBlockQuery:           spineBlockQuery,
		MainBlockQuery:            mainBlockQuery,
		SnapshotChunkQuery:        snapshotChunkQuery,
		Spinechain:                &chaintype.SpineChain{},
		Mainchain:                 &chaintype.MainChain{},
		SnapshotInterval:          constant.SnapshotInterval,
		SnapshotGenerationTimeout: constant.SnapshotGenerationTimeout,
		Logger:                    logger,
	}
}

// GetMegablockFromSpineHeight retrieve a megablock for a given spine height
// if there is no megablock at this height, return nil
func (ss *SnapshotService) GetMegablockFromSpineHeight(spineHeight uint32) (*model.Megablock, error) {
	var (
		megablock model.Megablock
		snapshotChunks []*model.SnapshotChunk
	)
	qry := ss.MegablockQuery.GetMegablocksByBlockHeight(spineHeight, ss.Spinechain)
	row, err := ss.QueryExecutor.ExecuteSelectRow(qry, false)
	if err != nil {
		return nil, err
	}
	err = ss.MegablockQuery.Scan(&megablock, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		return nil, nil
	}
	
	// populate snapshotChunks
	sqlStr := ss.SnapshotChunkQuery.GetSnapshotChunksByBlockHeight(spineHeight)
	rows, err := ss.QueryExecutor.ExecuteSelect(sqlStr, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	snapshotChunks, err = ss.SnapshotChunkQuery.BuildModel(snapshotChunks, rows)
	if err != nil {
		return nil, err
	}
	// never return nil for snapshotChunks, otherwise GetMegablockBytes will return error
	if snapshotChunks == nil {
		snapshotChunks = make([]*model.SnapshotChunk,0)
	}
	megablock.SnapshotChunks = snapshotChunks
	return &megablock, nil
}

// GetMegablockFromSpineHeight retrieve a megablock for a given spine height
func (ss *SnapshotService) GetLastMegablock() (*model.Megablock, error) {
	var (
		megablock model.Megablock
	)
	qry := ss.MegablockQuery.GetLastMegablock()
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
// snapshotHash: hash of the full snapshot content
// mainHeight: mainchain height at which the snapshot has started (note: this is not the captured snapshot's height,
// which should be = mainHeight - minRollbackHeight)
// sortedSnapshotChunksHashes all snapshot chunks hashes for this megablock (already sorted from first to last chunk)
// lastSnapshotChunkHash last available snapshot chunk hash (from db)
func (ss *SnapshotService) CreateMegablock(snapshotHash []byte, mainHeight, spineHeight uint32,
	sortedSnapshotChunksHashes [][]byte, lastSnapshotChunkHash []byte) (*model.Megablock, error) {
	var (
		previousChunkHash    []byte
		sortedSnapshotChunks = make([]*model.SnapshotChunk, 0)
	)

	// build the snapshot chunks
	previousChunkHash = lastSnapshotChunkHash
	for idx, chunkHash := range sortedSnapshotChunksHashes {
		snapshotChunk := &model.SnapshotChunk{
			ChunkHash:         chunkHash,
			ChunkIndex:        uint32(idx),
			PreviousChunkHash: previousChunkHash,
			SpineBlockHeight:  spineHeight,
		}
		sortedSnapshotChunks = append(sortedSnapshotChunks, snapshotChunk)
		previousChunkHash = chunkHash
	}

	// build the megablock
	megablock := &model.Megablock{
		FullSnapshotHash: snapshotHash,
		MainBlockHeight:  mainHeight,
		SpineBlockHeight: spineHeight,
		ChunksCount:      uint32(len(sortedSnapshotChunks)),
		SnapshotChunks:   sortedSnapshotChunks,
	}
	if err := ss.InsertMegablock(megablock); err != nil {
		return nil, err
	}
	return megablock, nil
}

// InsertMegablock persist a megablock to db (query wrapper)
func (ss *SnapshotService) InsertMegablock(megablock *model.Megablock) error {
	var (
		queries              [][]interface{}
	)
	if megablock.SnapshotChunks == nil {
		return blocker.NewBlocker(blocker.AppErr, "SnapshotChunksNil")
	}
	insertMegablockQ, insertMegablockArgs := ss.MegablockQuery.InsertMegablock(megablock)
	insertMegablockQry := append([]interface{}{insertMegablockQ}, insertMegablockArgs...)
	queries = append(queries, insertMegablockQry)

	for _, snapshotChunk := range megablock.SnapshotChunks {
		// add chunk to db transaction
		insertSnapshotChunkQ, insertSnapshotChunkArgs := ss.SnapshotChunkQuery.InsertSnapshotChunk(snapshotChunk)
		insertSnapshotChunkQry := append([]interface{}{insertSnapshotChunkQ}, insertSnapshotChunkArgs...)
		queries = append(queries, insertSnapshotChunkQry)
	}

	err := ss.QueryExecutor.ExecuteTransactions(queries)
	if err != nil {
		return err
	}
	return nil
}

// GetNextSnapshotHeight calculate next snapshot (main block) height given an arbitrary main block height
func (ss *SnapshotService) GetNextSnapshotHeight(mainHeight uint32) uint32 {
	// first snapshot cannot be taken before minRollBack height
	if mainHeight < constant.MinRollbackBlocks {
		mainHeight = constant.MinRollbackBlocks
	}
	avgBlockTime := ss.Mainchain.GetSmithingPeriod() + ss.Mainchain.GetChainSmithingDelayTime()
	avgBlockInterval := ss.SnapshotInterval / avgBlockTime
	return uint32(util.GetNextStep(int64(mainHeight), avgBlockInterval))
}

// GenerateSnapshot compute and persist a snapshot to file
// Note: First iteration will save a single chunk, for simplicity, but in future we should be able to split the file into multiple parts
func (ss *SnapshotService) GenerateSnapshot(mainHeight uint32) (*model.Megablock, error) {
	var (
		lastMainBlock, lastSpineBlock model.Block
		lastSnapshotChunk             model.SnapshotChunk
		firstValidSpineHeight         uint32
		lastSnapshotChunkHash         []byte
	)
	// get the last main block
	row, err := ss.QueryExecutor.ExecuteSelectRow(ss.MainBlockQuery.GetLastBlock(), false)
	if err != nil {
		return nil, err
	}
	err = ss.MainBlockQuery.Scan(&lastMainBlock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// get the last spine block
	row, err = ss.QueryExecutor.ExecuteSelectRow(ss.SpineBlockQuery.GetLastBlock(), false)
	if err != nil {
		return nil, err
	}
	err = ss.MainBlockQuery.Scan(&lastSpineBlock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	// get the last snapshot chunk's hash (to attach to the first chunk of the megablock, as previous chunk hash)
	row, err = ss.QueryExecutor.ExecuteSelectRow(ss.SnapshotChunkQuery.GetLastSnapshotChunk(), false)
	if err != nil {
		return nil, err
	}
	err = ss.SnapshotChunkQuery.Scan(&lastSnapshotChunk, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		lastSnapshotChunkHash = nil
	} else {
		lastSnapshotChunkHash = lastSnapshotChunk.ChunkHash
	}

	// calculate first valid spine block height for the snapshot (= megablock) to be included in.
	// spine blocks have discrete timing,
	// so we can calculate accurately next spine timestamp and give enough time to all nodes to complete their snapshot
	spinechainInterval := ss.Spinechain.GetSmithingPeriod() + ss.Spinechain.GetChainSmithingDelayTime()
	// lastMainBlock.Timestamp is the timestamp at which the snapshot started to be computed
	nextMinimumSpineBlockTimestamp := lastMainBlock.Timestamp + ss.SnapshotGenerationTimeout
	firstValidTime := util.GetNextStep(nextMinimumSpineBlockTimestamp, spinechainInterval)
	firstValidSpineHeight = uint32((firstValidTime - constant.SpinechainGenesisBlockTimestamp) / spinechainInterval)
	// don't allow megablocks to reference past spine blocks
	if firstValidSpineHeight < lastSpineBlock.Height {
		firstValidSpineHeight = lastSpineBlock.Height + 1
	}

	// TODO: call here the function that compute the snapshot and returns:
	//  the snapshot chunks' hashes
	//  the snapshot full hash
	var snapshotChunkHashes = make([][]byte, 0)
	var snapshotFullHash = make([]byte, 64)

	return ss.CreateMegablock(snapshotFullHash, mainHeight, firstValidSpineHeight, snapshotChunkHashes, lastSnapshotChunkHash)
}

// GetBodyBytes translate tx body to bytes representation
func (ss *SnapshotService) GetMegablockBytes(megablock *model.Megablock) ([]byte, error) {
	var (
		snapshotChunksBytes []byte
	)
	if megablock.SnapshotChunks == nil {
		return nil, blocker.NewBlocker(blocker.AppErr, "MegablockSnapshotChunksNil")
	}
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(megablock.FullSnapshotHash)
	buffer.Write(util.ConvertUint32ToBytes(megablock.ChunksCount))
	buffer.Write(util.ConvertUint32ToBytes(megablock.SpineBlockHeight))
	buffer.Write(util.ConvertUint32ToBytes(megablock.MainBlockHeight))
	// snapshot chunks
	for _, sc := range megablock.SnapshotChunks {
		snapshotChunksBytes = append(snapshotChunksBytes, ss.GetSnapshotChunkBytes(sc)...)
	}
	buffer.Write(snapshotChunksBytes)
	return buffer.Bytes(), nil
}

func (ss *SnapshotService) GetSnapshotChunkBytes(snapshotChunk *model.SnapshotChunk) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(snapshotChunk.ChunkHash)
	buffer.Write(util.ConvertUint32ToBytes(snapshotChunk.ChunkIndex))
	buffer.Write(snapshotChunk.PreviousChunkHash)
	buffer.Write(util.ConvertUint32ToBytes(snapshotChunk.SpineBlockHeight))
	return buffer.Bytes()
}

// StartSnapshotListener setup listener for transaction to the list peer
func (ss *SnapshotService) StartSnapshotListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			b := block.(*model.Block)
			if b.Height == ss.GetNextSnapshotHeight(b.Height) {
				go func() {
					// TODO: implement some process management,
					//  such as controlling if there is another snapshot running before starting to compute a new one (
					//  or compute the new one and kill the one already running...)
					if _, err := ss.GenerateSnapshot(b.Height); err != nil {
						ss.Logger.Errorf("Snapshot at main block "+
							"height %d terminated with errors %s", b.Height, err)
					}
					ss.Logger.Infof("Snapshot at main block "+
						"height %d terminated successfully", b.Height)
				}()
			}
		},
	}
}
