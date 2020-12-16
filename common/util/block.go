package util

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"golang.org/x/crypto/sha3"
)

// GetBlockHash return the block's bytes hash.
// note: the block must be signed, otherwise this function returns an error
func GetBlockHash(block *model.Block, ct chaintype.ChainType) ([]byte, error) {
	var (
		digest = sha3.New256()
	)
	// TODO: this error should be managed. for now we leave it because it causes a cascade of failures in unit tests..
	blockByte, _ := GetBlockByte(block, true, ct)
	_, err := digest.Write(blockByte)
	if err != nil {
		return nil, err
	}
	return digest.Sum([]byte{}), nil
}

// GetBlockByte generate value for `Bytes` field if not assigned yet
// return .`Bytes` if value assigned
// TODO: Abstract this method in BlockCoreService or ChainType to decouple business logic from block type
func GetBlockByte(block *model.Block, signed bool, ct chaintype.ChainType) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(ConvertUint32ToBytes(block.GetVersion()))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTimestamp())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTotalAmount())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTotalFee())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTotalCoinBase())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetPayloadLength())))
	buffer.Write(block.PayloadHash)
	buffer.Write(block.BlocksmithPublicKey)
	buffer.Write(block.GetBlockSeed())
	buffer.Write(block.GetPreviousBlockHash())
	if signed {
		if block.BlockSignature == nil {
			return nil, blocker.NewBlocker(blocker.BlockErr, "invalid signature")
		}
		buffer.Write(block.BlockSignature)
	}
	return buffer.Bytes(), nil
}

func IsBlockIDExist(blockIds []int64, expectedBlockID int64) bool {
	for _, blockID := range blockIds {
		if blockID == expectedBlockID {
			return true
		}
	}
	return false
}

// GetLastBlock TODO: this should be used by services instead of blockService.GetLastBlock
func GetLastBlock(
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) (*model.Block, error) {

	var (
		qry   = blockQuery.GetLastBlock()
		block model.Block
		row   *sql.Row
		err   error
	)

	// note: no need to check for the error here, since dbTx is false
	row, _ = queryExecutor.ExecuteSelectRow(qry, false)
	err = blockQuery.Scan(&block, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		return nil, blocker.NewBlocker(blocker.DBErr, "LastBlockNotFound")
	}

	return &block, nil
}

// GetBlockByHeight  get block at the height provided
// TODO: this should be used by services instead of blockService.GetLastBlock
func GetBlockByHeight(
	height uint32,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) (*model.Block, error) {
	var (
		block model.Block
		row   *sql.Row
		err   error
	)
	row, err = queryExecutor.ExecuteSelectRow(blockQuery.GetBlockByHeight(height), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = blockQuery.Scan(&block, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, "BlockScanErr, ", err.Error())
		}
		return nil, blocker.NewBlocker(blocker.DBRowNotFound, "BlockNotFound")
	}
	return &block, nil
}

// GetBlockByID get block at the ID provided
func GetBlockByID(
	id int64,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) (*model.Block, error) {
	if id == 0 {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "BlockIDZeroNotFound")
	}
	var (
		block    model.Block
		row, err = queryExecutor.ExecuteSelectRow(blockQuery.GetBlockByID(id), false)
	)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if err = blockQuery.Scan(&block, row); err != nil {
		if err != sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.DBErr, "BlockByIDScanErr, ", err.Error())
		}
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "BlockNotFound")
	}
	return &block, nil
}

func GetBlockByHeightUseBlocksCache(
	height uint32,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	blocksCacheStorage storage.CacheStackStorageInterface,
) (*storage.BlockCacheObject, error) {
	var (
		blockCacheObject storage.BlockCacheObject
		err              = blocksCacheStorage.GetAtIndex(height, &blockCacheObject)
	)
	if err == nil {
		return &blockCacheObject, nil
	}
	block, err := GetBlockByHeight(height, queryExecutor, blockQuery)
	if err != nil {
		return nil, err
	}
	blockCacheObject = BlockConvertToCacheFormat(block)
	return &blockCacheObject, nil
}

func GetMinRollbackHeight(currentHeight uint32) uint32 {
	if currentHeight < constant.MinRollbackBlocks {
		return 0
	}
	return currentHeight - constant.MinRollbackBlocks
}

// GetSpinePublicKeyBytes convert a model.SpinePublicKey to []byte
func GetSpinePublicKeyBytes(spinePublicKey *model.SpinePublicKey) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(ConvertUint32ToBytes(spinePublicKey.MainBlockHeight))
	buffer.Write(spinePublicKey.NodePublicKey)
	buffer.Write(ConvertUint32ToBytes(uint32(spinePublicKey.PublicKeyAction)))
	buffer.Write(ConvertUint32ToBytes(spinePublicKey.Height))
	return buffer.Bytes()
}

func BlockConvertToCacheFormat(block *model.Block) storage.BlockCacheObject {
	var bHash = make([]byte, len(block.BlockHash))
	copy(bHash, block.BlockHash)
	return storage.BlockCacheObject{
		ID:        block.ID,
		Height:    block.Height,
		Timestamp: block.Timestamp,
		BlockHash: bHash,
	}
}
