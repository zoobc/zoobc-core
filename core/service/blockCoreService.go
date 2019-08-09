package service

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockServiceInterface interface {
		VerifySeed(seed *big.Int, balance *big.Int, previousBlock *model.Block, timestamp int64) bool
		NewBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte, hash string,
			previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, payloadLength uint32, secretPhrase string) *model.Block
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed []byte, blocksmithID []byte,
			hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, payloadLength uint32, smithScale int64, cumulativeDifficulty *big.Int,
			genesisSignature []byte) *model.Block
		GenerateBlock(
			previousBlock *model.Block,
			secretPhrase string,
			timestamp int64,
			blockSmithAccountID []byte,
		) (*model.Block, error)
		PushBlock(previousBlock, block *model.Block, needLock bool) error
		GetBlockByID(int64) (*model.Block, error)
		GetBlockByHeight(uint32) (*model.Block, error)
		GetBlocksFromHeight(uint32, uint32) ([]*model.Block, error)
		GetLastBlock() (*model.Block, error)
		GetBlocks() ([]*model.Block, error)
		GetGenesisBlock() (*model.Block, error)
		RemoveMempoolTransactions(transactions []*model.Transaction) error
		AddGenesis() error
		CheckGenesis() bool
		GetChainType() contract.ChainType
		ChainWriteLock()
		ChainWriteUnlock()
		ReceivedBlockListener() observer.Listener
	}

	BlockService struct {
		chainWriteLock      sync.WaitGroup
		Chaintype           contract.ChainType
		QueryExecutor       query.ExecutorInterface
		BlockQuery          query.BlockQueryInterface
		MempoolQuery        query.MempoolQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		Signature           crypto.SignatureInterface
		MempoolService      MempoolServiceInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Observer            *observer.Observer
	}
)

func NewBlockService(
	ct contract.ChainType,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mempoolQuery query.MempoolQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	signature crypto.SignatureInterface,
	mempoolService MempoolServiceInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	obsr *observer.Observer,
) *BlockService {
	return &BlockService{
		Chaintype:           ct,
		QueryExecutor:       queryExecutor,
		BlockQuery:          blockQuery,
		MempoolQuery:        mempoolQuery,
		TransactionQuery:    transactionQuery,
		Signature:           signature,
		MempoolService:      mempoolService,
		ActionTypeSwitcher:  txTypeSwitcher,
		AccountBalanceQuery: accountBalanceQuery,
		Observer:            obsr,
	}
}

// NewBlock generate new block
func (bs *BlockService) NewBlock(
	version uint32,
	previousBlockHash,
	blockSeed,
	blocksmithID []byte,
	hash string,
	previousBlockHeight uint32,
	timestamp,
	totalAmount,
	totalFee,
	totalCoinBase int64,
	transactions []*model.Transaction,
	payloadHash []byte,
	payloadLength uint32,
	secretPhrase string,
) *model.Block {
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
		PayloadLength:     payloadLength,
	}
	blockUnsignedByte, _ := coreUtil.GetBlockByte(block, false)
	block.BlockSignature = bs.Signature.SignByNode(blockUnsignedByte, secretPhrase)
	return block
}

// GetChainType returns the chaintype
func (bs *BlockService) GetChainType() contract.ChainType {
	return bs.Chaintype
}

// ChainWriteLock locks the chain
func (bs *BlockService) ChainWriteLock() {
	bs.chainWriteLock.Add(1)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockService) ChainWriteUnlock() {
	bs.chainWriteLock.Done()
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockService) NewGenesisBlock(
	version uint32,
	previousBlockHash, blockSeed, blocksmithID []byte,
	hash string,
	previousBlockHeight uint32,
	timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction,
	payloadHash []byte,
	payloadLength uint32,
	smithScale int64,
	cumulativeDifficulty *big.Int,
	genesisSignature []byte,
) *model.Block {
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
		PayloadLength:        payloadLength,
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
func (*BlockService) VerifySeed(
	seed, balance *big.Int,
	previousBlock *model.Block,
	timestamp int64,
) bool {
	elapsedTime := timestamp - previousBlock.GetTimestamp()
	effectiveSmithScale := new(big.Int).Mul(balance, big.NewInt(previousBlock.GetSmithScale()))
	prevTarget := new(big.Int).Mul(big.NewInt(elapsedTime-1), effectiveSmithScale)
	target := new(big.Int).Add(effectiveSmithScale, prevTarget)
	return seed.Cmp(target) < 0 && (seed.Cmp(prevTarget) >= 0 || elapsedTime > 300)
}

// PushBlock push block into blockchain
func (bs *BlockService) PushBlock(previousBlock, block *model.Block, needLock bool) error {
	// needLock indicates the push block needs to be protected
	if needLock {
		bs.chainWriteLock.Wait()
	}
	if previousBlock.GetID() != -1 {
		block.Height = previousBlock.GetHeight() + 1
		block = coreUtil.CalculateSmithScale(previousBlock, block, bs.Chaintype.GetChainSmithingDelayTime())
	}
	// start db transaction here
	_ = bs.QueryExecutor.BeginTx()
	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	err := bs.QueryExecutor.ExecuteTransaction(blockInsertQuery, blockInsertValue...)
	if err != nil {
		return err
	}
	// apply transactions and remove them from mempool
	transactions := block.GetTransactions()
	if len(transactions) > 0 {
		for _, tx := range block.GetTransactions() {
			// assign block id and block height to tx
			tx.BlockID = block.ID
			tx.Height = block.Height
			// validate the transaction
			if err := util.ValidateTransaction(tx, bs.QueryExecutor, bs.AccountBalanceQuery, true); err != nil {
				return err
			}
			// validate tx body and apply/perform transaction-specific logic
			err := bs.ActionTypeSwitcher.GetTransactionType(tx).ApplyConfirmed() // todo: make this mockable
			if err == nil {
				transactionInsertQuery, transactionInsertValue := bs.TransactionQuery.InsertTransaction(tx)
				err := bs.QueryExecutor.ExecuteTransaction(transactionInsertQuery, transactionInsertValue...)
				if err != nil {
					_ = bs.QueryExecutor.RollbackTx()
					return err
				}
			} else {
				_ = bs.QueryExecutor.RollbackTx()
				return err
			}
		}
		if block.Height != 0 {
			if err := bs.RemoveMempoolTransactions(transactions); err != nil {
				log.Errorf("Can't delete Mempool Transactions: %s", err)
				return err
			}
		}
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	// broadcast block
	bs.Observer.Notify(observer.BlockPushed, block, nil)
	return nil

}

// GetBlockByID return the last pushed block
func (bs *BlockService) GetBlockByID(id int64) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByID(id))
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

func (bs *BlockService) GetBlocksFromHeight(startHeight uint32, limit uint32) ([]*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockFromHeight(startHeight, limit))
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return []*model.Block{}, err
	}
	var blocks []*model.Block
	blocks = bs.BlockQuery.BuildModel(blocks, rows)
	return blocks, nil
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

// GetLastBlock return the last pushed block
func (bs *BlockService) GetBlockByHeight(height uint32) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByHeight(height))
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
		err = rows.Scan(
			&lastBlock.ID,
			&lastBlock.PreviousBlockHash,
			&lastBlock.Height,
			&lastBlock.Timestamp,
			&lastBlock.BlockSeed,
			&lastBlock.BlockSignature,
			&lastBlock.CumulativeDifficulty,
			&lastBlock.SmithScale,
			&lastBlock.PayloadLength,
			&lastBlock.PayloadHash,
			&lastBlock.BlocksmithID,
			&lastBlock.TotalAmount,
			&lastBlock.TotalFee,
			&lastBlock.TotalCoinBase,
			&lastBlock.Version,
		)
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
	var idsStr []string
	for _, tx := range transactions {
		idsStr = append(idsStr, strconv.FormatInt(tx.ID, 10))
	}
	err := bs.QueryExecutor.ExecuteTransaction(bs.MempoolQuery.DeleteMempoolTransactions(), strings.Join(idsStr, ","))
	if err != nil {
		return err
	}
	log.Printf("mempool transaction with IDs = %s deleted", idsStr)
	return nil
}

// GenerateBlock generate block from transactions in mempool
func (bs *BlockService) GenerateBlock(
	previousBlock *model.Block,
	secretPhrase string,
	timestamp int64,
	blockSmithAccountID []byte,
) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinbase int64
		//TODO: missing coinbase calculation
		payloadLength uint32
		// only for mainchain
		sortedTx    []*model.Transaction
		payloadHash []byte
		digest      = sha3.New512()
	)

	newBlockHeight := previousBlock.Height + 1

	if _, ok := bs.Chaintype.(*chaintype.MainChain); ok {
		mempoolTransactions, err := bs.MempoolService.SelectTransactionsFromMempool(timestamp)
		if err != nil {
			return nil, errors.New("MempoolReadError")
		}
		for _, mpTx := range mempoolTransactions {
			tx, err := util.ParseTransactionBytes(mpTx.TransactionBytes, true)
			if err != nil {
				return nil, err
			}

			sortedTx = append(sortedTx, tx)
			_, _ = digest.Write(mpTx.TransactionBytes)
			txType := bs.ActionTypeSwitcher.GetTransactionType(tx)
			totalAmount += txType.GetAmount()
			totalFee += tx.Fee
			payloadLength += txType.GetSize()
		}
		payloadHash = digest.Sum([]byte{})
	}

	// loop through transaction to build block hash
	hash := digest.Sum([]byte{})
	digest.Reset() // reset the digest
	_, _ = digest.Write(previousBlock.GetBlockSeed())
	_, _ = digest.Write(blockSmithAccountID)
	blockSeed := digest.Sum([]byte{})
	digest.Reset() // reset the digest
	previousBlockByte, _ := coreUtil.GetBlockByte(previousBlock, true)
	previousBlockHash := sha3.Sum512(previousBlockByte)
	block := bs.NewBlock(
		1,
		previousBlockHash[:],
		blockSeed,
		blockSmithAccountID,
		string(hash),
		newBlockHeight,
		timestamp,
		totalAmount,
		totalFee,
		totalCoinbase,
		sortedTx,
		payloadHash,
		payloadLength,
		secretPhrase,
	)
	log.Printf("block forged: fee %d\n", totalFee)
	return block, nil
}

// AddGenesis add genesis block of chain to the chain
func (bs *BlockService) AddGenesis() error {
	var (
		totalAmount, totalFee, totalCoinBase int64
		blockTransactions                    []*model.Transaction
		payloadLength                        uint32
		digest                               = sha3.New512()
	)
	for _, tx := range GetGenesisTransactions(bs.Chaintype) {
		txBytes, _ := util.GetTransactionBytes(tx, true)
		_, _ = digest.Write(txBytes)
		if tx.TransactionType == util.ConvertBytesToUint32([]byte{1, 0, 0, 0}) { // if type = send money
			totalAmount += tx.GetSendMoneyTransactionBody().Amount
		}
		txType := bs.ActionTypeSwitcher.GetTransactionType(tx)
		totalAmount += txType.GetAmount()
		totalFee += tx.Fee
		payloadLength += txType.GetSize()
		blockTransactions = append(blockTransactions, tx)
	}
	payloadHash := digest.Sum([]byte{})
	block := bs.NewGenesisBlock(
		1,
		nil,
		make([]byte, 64),
		[]byte{},
		"",
		0,
		constant.GenesisBlockTimestamp,
		totalAmount,
		totalFee,
		totalCoinBase,
		blockTransactions,
		payloadHash,
		payloadLength,
		constant.InitialSmithScale,
		big.NewInt(0),
		constant.GenesisBlockSignature,
	)
	// assign genesis block id
	block.ID = coreUtil.GetBlockID(block)
	fmt.Printf("\n\ngenesis block: %v\n\n ", block)
	err := bs.PushBlock(&model.Block{ID: -1, Height: 0}, block, true)
	if err != nil {
		panic("PushGenesisBlock:fail")
	}
	return nil
}

// CheckGenesis check if genesis has been added
func (bs *BlockService) CheckGenesis() bool {
	genesisBlock, err := bs.GetGenesisBlock()
	if err != nil { // Genesis is not in the blockchain yet
		return false
	}
	if genesisBlock.ID != bs.Chaintype.GetGenesisBlockID() {
		log.Fatalf("Genesis ID does not match, expect: %d, get: %d", bs.Chaintype.GetGenesisBlockID(), genesisBlock.ID)
	}
	return true
}

// CheckSignatureBlock check signature of block
func (bs *BlockService) CheckSignatureBlock(block *model.Block) bool {
	if block.GetBlockSignature() != nil {
		blockUnsignedByte, err := coreUtil.GetBlockByte(block, false)
		if err != nil {
			return false
		}
		accountAddress, err := util.GetAddressFromPublicKey(block.GetBlocksmithID())
		if err != nil {
			return false
		}
		return bs.Signature.VerifySignature(blockUnsignedByte, block.GetBlockSignature(), 0, accountAddress)
	}
	return false
}

// ReceivedBlockListener handle received block from another node
func (bs *BlockService) ReceivedBlockListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			receivedBlock := block.(*model.Block)
			// make sure block has previous block hash
			if receivedBlock.GetPreviousBlockHash() != nil {
				if bs.CheckSignatureBlock(receivedBlock) {
					lastBlock, err := bs.GetLastBlock()
					if err != nil {
						return
					}

					lastBlockByte, _ := coreUtil.GetBlockByte(lastBlock, true)
					lastBlockHash := sha3.Sum512(lastBlockByte)

					//  check equality last block hash with previous block hash from received block
					if bytes.Equal(lastBlockHash[:], receivedBlock.GetPreviousBlockHash()) {
						err := bs.PushBlock(lastBlock, receivedBlock, true)
						if err != nil {
							return
						}
					}
				}
			}
		},
	}
}
