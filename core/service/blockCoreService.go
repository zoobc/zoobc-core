package service

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/query"

	"github.com/zoobc/zoobc-core/common/model"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

type (
	BlockServiceInterface interface {
		VerifySeed(seed *big.Int, balance *big.Int, previousBlock *model.Block, timestamp int64) bool
		NewBlock(version uint32, previousBlockHash []byte, blockSeed, blockSmithPublicKey []byte, hash string,
			previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, payloadLength uint32, secretPhrase string) *model.Block
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed, blockSmithPublicKey []byte,
			hash string, previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, payloadHash []byte, payloadLength uint32, smithScale int64, cumulativeDifficulty *big.Int,
			genesisSignature []byte) *model.Block
		GenerateBlock(
			previousBlock *model.Block,
			secretPhrase string,
			timestamp int64,
		) (*model.Block, error)
		ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error
		PushBlock(previousBlock, block *model.Block, needLock, broadcast bool) error
		GetBlockByID(int64) (*model.Block, error)
		GetBlockByHeight(uint32) (*model.Block, error)
		GetBlocksFromHeight(uint32, uint32) ([]*model.Block, error)
		GetLastBlock() (*model.Block, error)
		GetBlocks() ([]*model.Block, error)
		GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error)
		GetGenesisBlock() (*model.Block, error)
		RemoveMempoolTransactions(transactions []*model.Transaction) error
		AddGenesis() error
		CheckGenesis() bool
		GetChainType() chaintype.ChainType
		ChainWriteLock()
		ChainWriteUnlock()
		ReceiveBlock(
			senderPublicKey []byte,
			lastBlock,
			block *model.Block,
			nodeSecretPhrase string,
		) (*model.Receipt, error)
		GetParticipationScore(nodePublicKey []byte) (int64, error)
	}

	BlockService struct {
		sync.WaitGroup
		Chaintype               chaintype.ChainType
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		Observer                *observer.Observer
	}
)

func NewBlockService(
	ct chaintype.ChainType,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mempoolQuery query.MempoolQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	signature crypto.SignatureInterface,
	mempoolService MempoolServiceInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	obsr *observer.Observer,
) *BlockService {
	return &BlockService{
		Chaintype:               ct,
		QueryExecutor:           queryExecutor,
		BlockQuery:              blockQuery,
		MempoolQuery:            mempoolQuery,
		TransactionQuery:        transactionQuery,
		Signature:               signature,
		MempoolService:          mempoolService,
		ActionTypeSwitcher:      txTypeSwitcher,
		AccountBalanceQuery:     accountBalanceQuery,
		ParticipationScoreQuery: participationScoreQuery,
		Observer:                obsr,
	}
}

// NewBlock generate new block
func (bs *BlockService) NewBlock(
	version uint32,
	previousBlockHash,
	blockSeed, blockSmithPublicKey []byte,
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
		Version:             version,
		PreviousBlockHash:   previousBlockHash,
		BlockSeed:           blockSeed,
		BlocksmithPublicKey: blockSmithPublicKey,
		Height:              previousBlockHeight,
		Timestamp:           timestamp,
		TotalAmount:         totalAmount,
		TotalFee:            totalFee,
		TotalCoinBase:       totalCoinBase,
		Transactions:        transactions,
		PayloadHash:         payloadHash,
		PayloadLength:       payloadLength,
	}
	blockUnsignedByte, _ := util.GetBlockByte(block, false)
	block.BlockSignature = bs.Signature.SignByNode(blockUnsignedByte, secretPhrase)
	return block
}

// GetChainType returns the chaintype
func (bs *BlockService) GetChainType() chaintype.ChainType {
	return bs.Chaintype
}

// ChainWriteLock locks the chain
func (bs *BlockService) ChainWriteLock() {
	bs.Add(1)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockService) ChainWriteUnlock() {
	bs.Done()
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockService) NewGenesisBlock(
	version uint32,
	previousBlockHash, blockSeed, blockSmithPublicKey []byte,
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
		BlocksmithPublicKey:  blockSmithPublicKey,
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

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockService) ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	if block.GetTimestamp() > curTime+15 {
		return blocker.NewBlocker(blocker.BlockErr, "invalid timestamp")
	}
	if coreUtil.GetBlockID(block) == 0 {
		return blocker.NewBlocker(blocker.BlockErr, "invalid ID")
	}
	// Verify Signature
	sig := new(crypto.Signature)
	blockByte, err := commonUtils.GetBlockByte(block, false)
	if err != nil {
		return err
	}

	if !sig.VerifyNodeSignature(
		blockByte,
		block.BlockSignature,
		block.BlocksmithPublicKey,
	) {
		return blocker.NewBlocker(blocker.BlockErr, "invalid signature")
	}
	// Verify previous block hash
	previousBlockIDFromHash := new(big.Int)
	previousBlockIDFromHashInt := previousBlockIDFromHash.SetBytes([]byte{
		block.PreviousBlockHash[7],
		block.PreviousBlockHash[6],
		block.PreviousBlockHash[5],
		block.PreviousBlockHash[4],
		block.PreviousBlockHash[3],
		block.PreviousBlockHash[2],
		block.PreviousBlockHash[1],
		block.PreviousBlockHash[0],
	}).Int64()
	if previousLastBlock.ID != previousBlockIDFromHashInt {
		return blocker.NewBlocker(blocker.BlockErr, "invalid previous block hash")
	}
	return nil
}

// PushBlock push block into blockchain, to broadcast the block after pushing to own node, switch the
// broadcast flag to `true`, and `false` otherwise
func (bs *BlockService) PushBlock(previousBlock, block *model.Block, needLock, broadcast bool) error {
	// needLock indicates the push block needs to be protected
	if needLock {
		bs.Wait()
	}
	if previousBlock.GetID() != -1 && block.CumulativeDifficulty == "" && block.SmithScale == 0 {
		block.Height = previousBlock.GetHeight() + 1
		block = coreUtil.CalculateSmithScale(previousBlock, block, bs.Chaintype.GetChainSmithingDelayTime())
	}
	// start db transaction here
	_ = bs.QueryExecutor.BeginTx()
	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	err := bs.QueryExecutor.ExecuteTransaction(blockInsertQuery, blockInsertValue...)
	if err != nil {
		_ = bs.QueryExecutor.RollbackTx()
		return err
	}
	// apply transactions and remove them from mempool
	transactions := block.GetTransactions()
	if len(transactions) > 0 {
		for index, tx := range block.GetTransactions() {
			// assign block id and block height to tx
			tx.BlockID = block.ID
			tx.Height = block.Height
			tx.TransactionIndex = uint32(index) + 1
			// validate tx here
			// check if is in mempool : if yes, undo unconfirmed
			rows, err := bs.QueryExecutor.ExecuteSelect(bs.MempoolQuery.GetMempoolTransaction(), false, tx.ID)
			if err != nil {
				rows.Close()
				_ = bs.QueryExecutor.RollbackTx()
				return err
			}
			txType := bs.ActionTypeSwitcher.GetTransactionType(tx)
			if rows.Next() {
				// undo unconfirmed
				err = txType.UndoApplyUnconfirmed()
				if err != nil {
					rows.Close()
					_ = bs.QueryExecutor.RollbackTx()
					return err
				}
			}
			rows.Close()
			if block.Height > 0 {
				err = txType.Validate(true)
				if err != nil {
					_ = bs.QueryExecutor.RollbackTx()
					return err
				}
			}
			// validate tx body and apply/perform transaction-specific logic
			err = txType.ApplyConfirmed()
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
				_ = bs.QueryExecutor.RollbackTx()
				return err
			}
		}
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	// broadcast block
	if block.Height > 0 && broadcast {
		bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)
	}
	return nil
}

// GetBlockByID return the last pushed block
func (bs *BlockService) GetBlockByID(id int64) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByID(id), false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks = bs.BlockQuery.BuildModel(blocks, rows)
	if len(blocks) > 0 {
		return blocks[0], nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block %v is not found", id))
}

func (bs *BlockService) GetBlocksFromHeight(startHeight, limit uint32) ([]*model.Block, error) {
	var blocks []*model.Block
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockFromHeight(startHeight, limit), false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return []*model.Block{}, err
	}
	blocks = bs.BlockQuery.BuildModel(blocks, rows)
	return blocks, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetLastBlock() (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetLastBlock(), false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks = bs.BlockQuery.BuildModel(blocks, rows)
	if len(blocks) > 0 {
		transactions, err := bs.GetTransactionsByBlockID(blocks[0].ID)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		blocks[0].Transactions = transactions
		return blocks[0], nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "last block is not found")
}

// GetTransactionsByBlockID get transactions of the block
func (bs *BlockService) GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error) {
	var transactions []*model.Transaction

	// get transaction of the block
	transactionQ, transactionArg := bs.TransactionQuery.GetTransactionsByBlockID(blockID)
	rows, err := bs.QueryExecutor.ExecuteSelect(transactionQ, false, transactionArg...)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	return bs.TransactionQuery.BuildModel(transactions, rows), nil
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetBlockByHeight(height uint32) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByHeight(height), false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks = bs.BlockQuery.BuildModel(blocks, rows)
	if len(blocks) > 0 {
		return blocks[0], nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block with height %v is not found", height))

}

// GetGenesis return the last pushed block
func (bs *BlockService) GetGenesisBlock() (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetGenesisBlock(), false)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
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
			&lastBlock.BlocksmithPublicKey,
			&lastBlock.TotalAmount,
			&lastBlock.TotalFee,
			&lastBlock.TotalCoinBase,
			&lastBlock.Version,
		)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
		}
		return &lastBlock, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")

}

// GetBlocks return all pushed blocks
func (bs *BlockService) GetBlocks() ([]*model.Block, error) {
	var blocks []*model.Block
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlocks(0, 100), false)
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
			&block.CumulativeDifficulty, &block.SmithScale, &block.PayloadLength, &block.PayloadHash, &block.BlocksmithPublicKey,
			&block.TotalAmount, &block.TotalFee, &block.TotalCoinBase, &block.Version)
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
		idsStr = append(idsStr, "'"+strconv.FormatInt(tx.ID, 10)+"'")
	}
	err := bs.QueryExecutor.ExecuteTransaction(bs.MempoolQuery.DeleteMempoolTransactions(idsStr))
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
) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinbase int64
		//TODO: missing coinbase calculation
		payloadLength uint32
		// only for mainchain
		sortedTx            []*model.Transaction
		payloadHash         []byte
		digest              = sha3.New512()
		blockSmithPublicKey = util.GetPublicKeyFromSeed(secretPhrase)
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
	_, _ = digest.Write(blockSmithPublicKey)
	blockSeed := digest.Sum([]byte{})
	digest.Reset() // reset the digest
	previousBlockHash, err := coreUtil.GetBlockHash(previousBlock)
	if err != nil {
		return nil, err
	}
	block := bs.NewBlock(
		1,
		previousBlockHash,
		blockSeed,
		blockSmithPublicKey,
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
	for index, tx := range GetGenesisTransactions(bs.Chaintype) {
		txBytes, _ := util.GetTransactionBytes(tx, true)
		_, _ = digest.Write(txBytes)
		if tx.TransactionType == util.ConvertBytesToUint32([]byte{1, 0, 0, 0}) { // if type = send money
			totalAmount += tx.GetSendMoneyTransactionBody().Amount
		}
		txType := bs.ActionTypeSwitcher.GetTransactionType(tx)
		totalAmount += txType.GetAmount()
		totalFee += tx.Fee
		payloadLength += txType.GetSize()
		tx.TransactionIndex = uint32(index) + 1
		blockTransactions = append(blockTransactions, tx)
	}
	payloadHash := digest.Sum([]byte{})
	block := bs.NewGenesisBlock(
		1,
		nil,
		constant.MainchainGenesisBlockSeed,
		constant.MainchainGenesisNodePublicKey,
		"",
		0,
		constant.MainchainGenesisBlockTimestamp,
		totalAmount,
		totalFee,
		totalCoinBase,
		blockTransactions,
		payloadHash,
		payloadLength,
		constant.InitialSmithScale,
		big.NewInt(0),
		constant.MainchainGenesisBlockSignature,
	)
	// assign genesis block id
	block.ID = coreUtil.GetBlockID(block)
	err := bs.PushBlock(&model.Block{ID: -1, Height: 0}, block, true, false)
	if err != nil {
		log.Fatal("PushGenesisBlock:fail ", err)
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

// ReceiveBlock handle the block received from connected peers
func (bs *BlockService) ReceiveBlock(
	senderPublicKey []byte,
	lastBlock, block *model.Block,
	nodeSecretPhrase string,
) (*model.Receipt, error) {
	// make sure block has previous block hash
	if block.GetPreviousBlockHash() != nil {
		blockUnsignedByte, _ := util.GetBlockByte(block, false)
		if bs.Signature.VerifyNodeSignature(blockUnsignedByte, block.BlockSignature, block.BlocksmithPublicKey) {
			lastBlockByte, err := util.GetBlockByte(lastBlock, true)
			if err != nil {
				return nil, blocker.NewBlocker(
					blocker.BlockErr,
					"fail to get last block byte",
				)
			}
			lastBlockHash := sha3.Sum512(lastBlockByte)

			//  check equality last block hash with previous block hash from received block
			if !bytes.Equal(lastBlockHash[:], block.PreviousBlockHash) {
				return nil, blocker.NewBlocker(
					blocker.BlockErr,
					"previous block hash does not match with last block hash",
				)
			}
			err = bs.PushBlock(lastBlock, block, true, true)
			if err != nil {
				return nil, blocker.NewBlocker(blocker.ValidationErr, "invalid block, fail to push block")
			}
			// generate receipt and return as response
			// todo: lastblock last applied block, or incoming block?
			nodePublicKey := util.GetPublicKeyFromSeed(nodeSecretPhrase)
			blockHash, _ := util.GetBlockHash(block)
			receipt, err := util.GenerateReceipt(
				lastBlock,
				senderPublicKey,
				nodePublicKey,
				blockHash,
				constant.ReceiptDatumTypeBlock)
			if err != nil {
				return nil, err
			}
			receipt.RecipientSignature = bs.Signature.SignByNode(
				util.GetUnsignedReceiptBytes(receipt),
				nodeSecretPhrase,
			)
			return receipt, nil
		}
		return nil, blocker.NewBlocker(
			blocker.ValidationErr,
			"block signature invalid")
	}
	return nil, blocker.NewBlocker(
		blocker.BlockErr,
		"last block hash does not exist",
	)
}

// GetParticipationScore handle received block from another node
func (bs *BlockService) GetParticipationScore(nodePublicKey []byte) (int64, error) {
	var (
		participationScores []*model.ParticipationScore
	)
	participationScoreQ, args := bs.ParticipationScoreQuery.GetParticipationScoreByNodePublicKey(nodePublicKey)
	rows, err := bs.QueryExecutor.ExecuteSelect(participationScoreQ, false, args...)
	if err != nil {
		return 0, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	participationScores = bs.ParticipationScoreQuery.BuildModel(participationScores, rows)
	// if there aren't participation scores for this address/node, return 0
	if len(participationScores) == 0 {
		return 0, nil
	}
	return participationScores[0].Score, nil
}
