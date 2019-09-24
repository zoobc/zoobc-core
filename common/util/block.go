package util

import (
	"bytes"
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"golang.org/x/crypto/sha3"
)

// GetBlockIdFromHash returns blockID from given hash
func GetBlockIDFromHash(blockHash []byte) int64 {
	// mainly for testing
	if len(blockHash) < 8 {
		return -1
	}
	res := new(big.Int)
	return res.SetBytes([]byte{
		blockHash[7],
		blockHash[6],
		blockHash[5],
		blockHash[4],
		blockHash[3],
		blockHash[2],
		blockHash[1],
		blockHash[0],
	}).Int64()
}

// GetBlockHash return the block's bytes hash.
// note: the block must be signed, otherwise this function returns an error
func GetBlockHash(block *model.Block) ([]byte, error) {
	digest := sha3.New256()
	blockByte, _ := GetBlockByte(block, true)
	_, err := digest.Write(blockByte)
	if err != nil {
		return nil, err
	}
	return digest.Sum([]byte{}), nil
}

// GetBlockByte generate value for `Bytes` field if not assigned yet
// return .`Bytes` if value assigned
func GetBlockByte(block *model.Block, signed bool) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(ConvertUint32ToBytes(block.GetVersion()))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTimestamp())))
	buffer.Write(ConvertIntToBytes(len(block.GetTransactions())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTotalAmount())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTotalFee())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTotalCoinBase())))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetPayloadLength())))
	buffer.Write(block.PayloadHash)

	buffer.Write(block.BlocksmithPublicKey)
	// FIXME: remove this comment after making sure the one below is a repetition and can be deleted
	// buffer.Write(ConvertUint32ToBytes(uint32(len([]byte(block.BlocksmithPublicKey)))))
	// buffer.Write([]byte(block.GetBlocksmithAddress()))

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
func GetLastBlock(queryExecutor query.ExecutorInterface, blockQuery query.BlockQueryInterface) (*model.Block, error) {
	qry := blockQuery.GetLastBlock()
	rows, err := queryExecutor.ExecuteSelect(qry, false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var (
		blocks []*model.Block
	)
	blocks = blockQuery.BuildModel(blocks, rows)
	if len(blocks) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "LastBlockNotFound")
	}
	return blocks[0], nil
}

// GetBlockByHeight TODO: this should be used by services instead of blockService.GetLastBlock
func GetBlockByHeight(height uint32, queryExecutor query.ExecutorInterface, blockQuery query.BlockQueryInterface) (*model.Block, error) {
	qry := blockQuery.GetBlockByHeight(height)
	rows, err := queryExecutor.ExecuteSelect(qry, false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks = blockQuery.BuildModel(blocks, rows)
	if len(blocks) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "BlockNotFound")
	}
	return blocks[0], nil
}
