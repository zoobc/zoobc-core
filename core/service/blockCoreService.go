package service

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/zoobc/zoobc-core/common/contract"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/core/util"
	core_util "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockServiceInterface interface {
		VerifySeed(seed *big.Int, balance *big.Int, previousBlock *model.Block, timestamp int64) bool
		NewBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte, hash string,
			previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, secretPhrase string) *model.Block
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte,
			hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, smithScale int64, cumulativeDifficulty *big.Int,
			genesisSignature []byte) *model.Block
		PushBlock(previousBlock, block *model.Block) error
		GetLastBlock() (*model.Block, error)
		GetBlocks() ([]*model.Block, error)
		GetGenesisBlock() (*model.Block, error)
		RemoveMempoolTransactions(transactions []*model.Transaction) error
	}

	BlockService struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		MempoolQuery  query.MempoolQueryInterface
		Signature     crypto.SignatureInterface
	}
)

func NewBlockService(chaintype contract.ChainType, queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface, mempoolQuery query.MempoolQueryInterface, signature crypto.SignatureInterface) *BlockService {
	return &BlockService{
		Chaintype:     chaintype,
		QueryExecutor: queryExecutor,
		BlockQuery:    blockQuery,
		MempoolQuery:  mempoolQuery,
		Signature:     signature,
	}
}

// NewBlock generate new block
func (bs *BlockService) NewBlock(version uint32, previousBlockHash, blockSeed, blocksmithID []byte, hash string,
	previousBlockHeight uint32, timestamp, totalAmount, totalFee, totalCoinBase int64, transactions []*model.Transaction,
	payloadHash []byte, secretPhrase string) *model.Block {
	block := &model.Block{
		Version:           version,
		PreviousBlockHash: previousBlockHash,
		BlockSeed:         blockSeed,
		BlocksmithID:      blocksmithID,
		Height:            previousBlockHeight,
		Timestamp:         timestamp,
		TotalAmount:       totalAmount,
		TotalFee:          totalFee,
		TotalCoinBase:     totalCoinBase,
		Transactions:      transactions,
		PayloadHash:       payloadHash,
	}
	blockUnsignedByte, _ := core_util.GetBlockByte(block, false)
	block.BlockSignature = bs.Signature.SignBlock(blockUnsignedByte, secretPhrase)
	return block
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockService) NewGenesisBlock(version uint32, previousBlockHash, blockSeed, blocksmithID []byte,
	hash string, previousBlockHeight uint32, timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction, payloadHash []byte, smithScale int64, cumulativeDifficulty *big.Int,
	genesisSignature []byte) *model.Block {
	block := &model.Block{
		Version:              version,
		PreviousBlockHash:    previousBlockHash,
		BlockSeed:            blockSeed,
		BlocksmithID:         blocksmithID,
		Height:               previousBlockHeight,
		Timestamp:            timestamp,
		TotalAmount:          totalAmount,
		TotalFee:             totalFee,
		TotalCoinBase:        totalCoinBase,
		Transactions:         transactions,
		PayloadHash:          payloadHash,
		SmithScale:           smithScale,
		CumulativeDifficulty: cumulativeDifficulty.String(),
		BlockSignature:       genesisSignature,
	}
	return block
}

// VerifySeed Verify a block can be forged (by a given account, using computed seed value and account balance).
// Can be used to check who's smithing the next block (lastBlock) or if last forged block
// (previousBlock) is acceptable by the network (meaning has been smithed by a valid blocksmith).
func (*BlockService) VerifySeed(seed, balance *big.Int, previousBlock *model.Block, timestamp int64) bool {
	elapsedTime := timestamp - previousBlock.GetTimestamp()
	effectiveSmithScale := new(big.Int).Mul(balance, big.NewInt(previousBlock.GetSmithScale()))
	prevTarget := new(big.Int).Mul(big.NewInt(elapsedTime-1), effectiveSmithScale)
	target := new(big.Int).Add(effectiveSmithScale, prevTarget)
	return seed.Cmp(target) < 0 && (seed.Cmp(prevTarget) >= 0 || elapsedTime > 300)
}

// PushBlock push block into blockchain
func (bs *BlockService) PushBlock(previousBlock, block *model.Block) error {
	if previousBlock.GetID() != -1 {
		block.Height = previousBlock.GetHeight() + 1
		block = core_util.CalculateSmithScale(previousBlock, block, bs.Chaintype.GetChainSmithingDelayTime())
	}
	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	result, err := bs.QueryExecutor.ExecuteStatement(blockInsertQuery, blockInsertValue...)
	if err != nil {
		return err
	}

	// apply transactions and remove them from mempool
	transactions := block.GetTransactions()
	if len(transactions) > 0 {
		for _, tx := range block.GetTransactions() {
			//TODO: not 100% sure if we need to call ApplyUnconfirmed or ApplyConfirmed
			if err := transaction.GetTransactionType(tx).ApplyUnconfirmed(); err != nil {
				tx.BlockID = block.ID
				tx.Height = block.Height
				//TODO: do we need to recompute txID (in previous prototype we used to do it)?
				//		note the function also checks if tx is signed (only checks that signature field is !nil)
				tx.ID, err = util.GetTransactionID(tx, bs.Chaintype)
				if err != nil {
					log.Error(err)
					continue
				}
				// TODO: save tx to db, in a (sql) database transaction (unless ApplyUnconfirmed already saves tx in db)
			}
		}
		if err := bs.RemoveMempoolTransactions(transactions); err != nil {
			log.Errorf("Can't delete Mempool Transactions: %s", err)
			return err
		}

		//TODO: add db transaction commit here
	}

	// broadcast block

	fmt.Printf("got new block, %v", result)
	return nil
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetLastBlock() (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetLastBlock())
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return &model.Block{
			ID: -1,
		}, err
	}
	var blocks []*model.Block
	blocks = bs.BlockQuery.BuildModel(blocks, rows)
	if len(blocks) > 0 {
		return blocks[0], nil
	}
	return &model.Block{
		ID: -1,
	}, errors.New("BlockNotFound")

}

// GetGenesis return the last pushed block
func (bs *BlockService) GetGenesisBlock() (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetGenesisBlock())
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return &model.Block{
			ID: -1,
		}, err
	}
	var lastBlock model.Block
	if rows.Next() {
		err = rows.Scan(&lastBlock.ID, &lastBlock.PreviousBlockHash, &lastBlock.Height, &lastBlock.Timestamp, &lastBlock.BlockSeed,
			&lastBlock.BlockSignature, &lastBlock.CumulativeDifficulty, &lastBlock.SmithScale, &lastBlock.PayloadLength,
			&lastBlock.PayloadHash, &lastBlock.BlocksmithID, &lastBlock.TotalAmount, &lastBlock.TotalFee, &lastBlock.TotalCoinBase,
			&lastBlock.Version)
		if err != nil {
			return &model.Block{
				ID: -1,
			}, err
		}
		return &lastBlock, nil
	}
	return &model.Block{
		ID: -1,
	}, errors.New("BlockNotFound")

}

// GetBlocks return all pushed blocks
func (bs *BlockService) GetBlocks() ([]*model.Block, error) {
	var blocks []*model.Block
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlocks(0, 100))
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var block model.Block
		err = rows.Scan(&block.ID, &block.PreviousBlockHash, &block.Height, &block.Timestamp, &block.BlockSeed, &block.BlockSignature,
			&block.CumulativeDifficulty, &block.SmithScale, &block.PayloadLength, &block.PayloadHash, &block.BlocksmithID, &block.TotalAmount,
			&block.TotalFee, &block.TotalCoinBase, &block.Version)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}
	return blocks, nil
}

// RemoveMempoolTransactions removes a list of transactions tx from mempool given their Ids
func (bs *BlockService) RemoveMempoolTransactions(transactions []*model.Transaction) error {
	idsStr := []string{}
	for _, tx := range transactions {
		idsStr = append(idsStr, strconv.FormatInt(tx.ID, 10))
	}
	_, err := bs.QueryExecutor.ExecuteStatement(bs.MempoolQuery.DeleteMempoolTransactions(), strings.Join(idsStr, ","))
	if err != nil {
		return err
	}
	log.Printf("mempool transaction with IDs = %s deleted", idsStr)
	return nil
}
