package util

import (
	"bytes"
	"database/sql"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"golang.org/x/crypto/sha3"
)

// GetBlockHash return the block's bytes hash.
// note: the block must be signed, otherwise this function returns an error
func GetBlockHash(block *model.Block, ct chaintype.ChainType) ([]byte, error) {
	var (
		digest     = sha3.New256()
		cloneBlock = *block
	)
	cloneBlock.BlockHash = nil
	// TODO: this error should be managed. for now we leave it because it causes a cascade of failures in unit tests..
	blockByte, _ := GetBlockByte(&cloneBlock, true, ct)
	_, err := digest.Write(blockByte)
	if err != nil {
		return nil, err
	}
	return digest.Sum([]byte{}), nil
}

// GetBlockByte generate value for `Bytes` field if not assigned yet
// return .`Bytes` if value assigned
//TODO: Abstract this method is BlockCoreService or ChainType to decouple business logic from block type
func GetBlockByte(block *model.Block, signed bool, ct chaintype.ChainType) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(ConvertUint32ToBytes(block.GetVersion()))
	buffer.Write(ConvertUint64ToBytes(uint64(block.GetTimestamp())))
	switch ct.(type) {
	case *chaintype.MainChain:
		// added nil check to make sure that transactions for this block have been populated, even if there are none (empty slice).
		// if block object doesn't have transactions populated (GetTransactions() = nil), block hash validation will fail later on
		if block.GetTransactions() == nil {
			return nil, blocker.NewBlocker(blocker.BlockErr, "main block transactions is nil")
		}
		buffer.Write(ConvertIntToBytes(len(block.GetTransactions())))
	case *chaintype.SpineChain:
		// added nil check to make sure that spine public keys for this block have been populated, even if there are none (empty slice).
		// if block object doesn't have spine pub keys populated (GetSpinePublicKeys() = nil), block hash validation will fail later on
		if block.GetSpinePublicKeys() == nil {
			return nil, blocker.NewBlocker(blocker.BlockErr, "spine block public keys is nil")
		}
		buffer.Write(ConvertIntToBytes(len(block.GetSpinePublicKeys())))
	}
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

// GetBlockByHeight TODO: this should be used by services instead of blockService.GetLastBlock
func GetBlockByHeight(
	height uint32,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
) (*model.Block, error) {
	qry := blockQuery.GetBlockByHeight(height)
	rows, err := queryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	var blocks []*model.Block
	blocks, err = blockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed build block into block model")
	}

	if len(blocks) == 0 {
		return nil, blocker.NewBlocker(blocker.DBErr, "BlockNotFound")
	}
	return blocks[0], nil
}

func GetMinRollbackHeight(currentHeight uint32) uint32 {
	if currentHeight < constant.MinRollbackBlocks {
		return 0
	}
	return currentHeight - constant.MinRollbackBlocks
}
