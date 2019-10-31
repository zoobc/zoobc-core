package service

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"sync"

	"github.com/dgraph-io/badger"

	"github.com/zoobc/zoobc-core/common/kvdb"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
)

type (
	BlockServiceInterface interface {
		VerifySeed(seed, score *big.Int, previousBlock *model.Block, timestamp int64) bool
		NewBlock(version uint32, previousBlockHash []byte, blockSeed, blockSmithPublicKey []byte,
			previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, blockReceipts []*model.PublishedReceipt, payloadHash []byte, payloadLength uint32,
			secretPhrase string) *model.Block
		NewGenesisBlock(version uint32, previousBlockHash []byte, blockSeed, blockSmithPublicKey []byte,
			previousBlockHeight uint32, timestamp int64, totalAmount int64, totalFee int64, totalCoinBase int64,
			transactions []*model.Transaction, blockReceipts []*model.PublishedReceipt, payloadHash []byte, payloadLength uint32, smithScale int64,
			cumulativeDifficulty *big.Int, genesisSignature []byte) *model.Block
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
		GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		GetGenesisBlock() (*model.Block, error)
		RemoveMempoolTransactions(transactions []*model.Transaction) error
		GenerateGenesisBlock(genesisEntries []constant.MainchainGenesisConfigEntry) (*model.Block, error)
		AddGenesis() error
		CheckGenesis() bool
		GetChainType() chaintype.ChainType
		ChainWriteLock()
		ChainWriteUnlock()
		GetCoinbase() int64
		CoinbaseLotteryWinners() ([]string, error)
		RewardBlocksmithAccountAddresses(blocksmithAccountAddresses []string, totalReward int64, height uint32) error
		GetBlocksmithAccountAddress(block *model.Block) (string, error)
		ReceiveBlock(
			senderPublicKey []byte,
			lastBlock,
			block *model.Block,
			nodeSecretPhrase string,
		) (*model.BatchReceipt, error)
		GetParticipationScore(nodePublicKey []byte) (int64, error)
		GetBlockExtendedInfo(block *model.Block) (*model.BlockExtendedInfo, error)
		GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error)
	}

	BlockService struct {
		sync.WaitGroup
		Chaintype               chaintype.ChainType
		KVExecutor              kvdb.KVExecutorInterface
		QueryExecutor           query.ExecutorInterface
		BlockQuery              query.BlockQueryInterface
		MempoolQuery            query.MempoolQueryInterface
		TransactionQuery        query.TransactionQueryInterface
		MerkleTreeQuery         query.MerkleTreeQueryInterface
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		Signature               crypto.SignatureInterface
		MempoolService          MempoolServiceInterface
		ReceiptService          ReceiptServiceInterface
		NodeRegistrationService NodeRegistrationServiceInterface
		ActionTypeSwitcher      transaction.TypeActionSwitcher
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		Observer                *observer.Observer
		SortedBlocksmiths       *[]model.Blocksmith
		Logger                  *log.Logger
	}
)

func NewBlockService(
	ct chaintype.ChainType,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mempoolQuery query.MempoolQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	signature crypto.SignatureInterface,
	mempoolService MempoolServiceInterface,
	receiptService ReceiptServiceInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	txTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	participationScoreQuery query.ParticipationScoreQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	obsr *observer.Observer,
	sortedBlocksmiths *[]model.Blocksmith,
	logger *log.Logger,
) *BlockService {
	return &BlockService{
		Chaintype:               ct,
		KVExecutor:              kvExecutor,
		QueryExecutor:           queryExecutor,
		BlockQuery:              blockQuery,
		MempoolQuery:            mempoolQuery,
		TransactionQuery:        transactionQuery,
		MerkleTreeQuery:         merkleTreeQuery,
		PublishedReceiptQuery:   publishedReceiptQuery,
		Signature:               signature,
		MempoolService:          mempoolService,
		ReceiptService:          receiptService,
		NodeRegistrationService: nodeRegistrationService,
		ActionTypeSwitcher:      txTypeSwitcher,
		AccountBalanceQuery:     accountBalanceQuery,
		ParticipationScoreQuery: participationScoreQuery,
		NodeRegistrationQuery:   nodeRegistrationQuery,
		Observer:                obsr,
		SortedBlocksmiths:       sortedBlocksmiths,
		Logger:                  logger,
	}
}

// NewBlock generate new block
func (bs *BlockService) NewBlock(
	version uint32,
	previousBlockHash,
	blockSeed, blockSmithPublicKey []byte,
	previousBlockHeight uint32,
	timestamp,
	totalAmount,
	totalFee,
	totalCoinBase int64,
	transactions []*model.Transaction,
	publishedReceipts []*model.PublishedReceipt,
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
		PublishedReceipts:   publishedReceipts,
		PayloadHash:         payloadHash,
		PayloadLength:       payloadLength,
	}
	blockUnsignedByte, err := util.GetBlockByte(block, false)
	if err != nil {
		bs.Logger.Error(err.Error())
	}
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
	previousBlockHeight uint32,
	timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction,
	publishedReceipts []*model.PublishedReceipt,
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
		PublishedReceipts:    publishedReceipts,
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
	seed, score *big.Int,
	previousBlock *model.Block,
	timestamp int64,
) bool {
	elapsedTime := timestamp - previousBlock.GetTimestamp()
	if elapsedTime <= 0 {
		return false
	}
	effectiveSmithScale := new(big.Int).Mul(score, big.NewInt(previousBlock.GetSmithScale()))
	prevTarget := new(big.Int).Mul(big.NewInt(elapsedTime-1), effectiveSmithScale)
	target := new(big.Int).Add(effectiveSmithScale, prevTarget)
	return seed.Cmp(target) < 0 && (seed.Cmp(prevTarget) >= 0 || elapsedTime > 3600)
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockService) ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	if block.GetTimestamp() > curTime+constant.GenerateBlockTimeoutSec {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidTimestamp")
	}
	if coreUtil.GetBlockID(block) == 0 {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidID")
	}
	// Verify Signature
	blockByte, err := commonUtils.GetBlockByte(block, false)
	if err != nil {
		return err
	}

	if !bs.Signature.VerifyNodeSignature(
		blockByte,
		block.BlockSignature,
		block.BlocksmithPublicKey,
	) {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSignature")
	}
	// Verify previous block hash
	previousBlockHash, err := util.GetBlockHash(previousLastBlock)
	if err != nil {
		return err
	}
	if !bytes.Equal(previousBlockHash, block.PreviousBlockHash) {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidPreviousBlockHash")
	}
	// if the same block height is already in the database compare cummulative difficulty.
	if err := bs.validateBlockHeight(block); err != nil {
		return err
	}
	return nil
}

// validateBlockAtHeight Check if the same block height is already in the database compare cummulative difficulty.
// and return error if current block's cumulative difficulty is lower than the one in db
func (bs *BlockService) validateBlockHeight(block *model.Block) error {
	var (
		bl                                                 []*model.Block
		refCumulativeDifficulty, blockCumulativeDifficulty *big.Int
		ok                                                 bool
	)
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByHeight(block.Height), false)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()
	bl, err = bs.BlockQuery.BuildModel(bl, rows)
	if err != nil {
		return err
	}
	if len(bl) > 0 {
		refBlock := bl[0]
		if refCumulativeDifficulty, ok = new(big.Int).SetString(refBlock.CumulativeDifficulty, 10); !ok {
			return err
		}
		if blockCumulativeDifficulty, ok = new(big.Int).SetString(block.CumulativeDifficulty, 10); !ok {
			return err
		}
		// if cumulative difficulty of the referece block is > of the one of the (new) block, new block is invalid
		if refCumulativeDifficulty.Cmp(blockCumulativeDifficulty) > 0 {
			return blocker.NewBlocker(blocker.BlockErr, "InvalidCumulativeDifficulty")
		}
	}
	return nil
}

// PushBlock push block into blockchain, to broadcast the block after pushing to own node, switch the
// broadcast flag to `true`, and `false` otherwise
func (bs *BlockService) PushBlock(previousBlock, block *model.Block, needLock, broadcast bool) error {
	var (
		err error
	)
	// needLock indicates the push block needs to be protected
	if needLock {
		bs.Wait()
	}

	if !coreUtil.IsGenesis(previousBlock.GetID(), block) {
		block.Height = previousBlock.GetHeight() + 1
		block, err = coreUtil.CalculateSmithScale(
			previousBlock, block, bs.Chaintype.GetSmithingPeriod(), bs.BlockQuery, bs.QueryExecutor,
		)
		if err != nil {
			return err
		}
	}
	// start db transaction here
	err = bs.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	blockInsertQuery, blockInsertValue := bs.BlockQuery.InsertBlock(block)
	err = bs.QueryExecutor.ExecuteTransaction(blockInsertQuery, blockInsertValue...)
	if err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	// apply transactions and remove them from mempool
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
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			rows.Close()
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		if rows.Next() {
			// undo unconfirmed
			err = txType.UndoApplyUnconfirmed()
			if err != nil {
				rows.Close()
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
		}
		rows.Close()
		if block.Height > 0 {
			err = txType.Validate(true)
			if err != nil {
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
		}
		// validate tx body and apply/perform transaction-specific logic
		err = txType.ApplyConfirmed()
		if err == nil {
			transactionInsertQuery, transactionInsertValue := bs.TransactionQuery.InsertTransaction(tx)
			err := bs.QueryExecutor.ExecuteTransaction(transactionInsertQuery, transactionInsertValue...)
			if err != nil {
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
		} else {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	if block.Height != 0 {
		if err := bs.RemoveMempoolTransactions(block.GetTransactions()); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	linkedCount, err := bs.processPublishedReceipts(block)
	if err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	if block.Height > 0 {
		// this is to manage the edge case when the blocksmith array has not been initialized yet:
		// when start smithing from a block with height > 0, since SortedBlocksmiths are computed  after a block is pushed,
		// for the first block that is pushed, we don't know who are the blocksmith to be rewarded
		if len(*bs.SortedBlocksmiths) == 0 {
			blocksmiths, err := bs.GetBlocksmiths(block)
			if err != nil {
				if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
					bs.Logger.Error(rollbackErr.Error())
				}
				return err
			}
			tmpBlocksmiths := make([]model.Blocksmith, 0)
			// copy the nextBlocksmiths pointers array into an array of blocksmiths
			for _, blocksmith := range blocksmiths {
				tmpBlocksmiths = append(tmpBlocksmiths, *blocksmith)
			}
			*bs.SortedBlocksmiths = tmpBlocksmiths
		}
		popScore, err := commonUtils.CalculateParticipationScore(
			uint32(linkedCount),
			uint32(len(block.GetPublishedReceipts())-linkedCount),
			uint32(len(*bs.SortedBlocksmiths)-1),
		)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		err = bs.updatePopScore(popScore, block)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}

		// selecting multiple account to be rewarded and split the total coinbase + totalFees evenly between them
		totalReward := block.TotalFee + block.TotalCoinBase
		lotteryAccounts, err := bs.CoinbaseLotteryWinners()
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		if err := bs.RewardBlocksmithAccountAddresses(lotteryAccounts, totalReward, block.Height); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}

	// admit/expel nodes from registry at genesis and regular intervals
	if block.Height == 0 || block.Height%bs.NodeRegistrationService.GetNodeAdmittanceCycle() == 0 {
		if err := bs.updateNodeRegistry(block); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}

	err = bs.QueryExecutor.CommitTx()
	if err != nil { // commit automatically unlock executor and close tx
		return err
	}
	// broadcast block
	if broadcast {
		bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
	}
	bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)
	return nil
}

// updateNodeRegistry seelct and admit/expel nodes from node registry
func (bs *BlockService) updateNodeRegistry(block *model.Block) error {
	// select n (= MaxNodeAdmittancePerCycle) queued nodes with the highest locked balance from node registry
	nodeRegistrations, err := bs.NodeRegistrationService.SelectNodesToBeAdmitted(constant.MaxNodeAdmittancePerCycle)
	if err != nil {
		return err
	}
	if len(nodeRegistrations) > 0 {
		err = bs.NodeRegistrationService.AdmitNodes(nodeRegistrations, block.Height)
		if err != nil {
			return err
		}
	}
	// expel nodes with zero score from node registry
	nodeRegistrations, err = bs.NodeRegistrationService.SelectNodesToBeExpelled()
	if err != nil {
		return err
	}
	if len(nodeRegistrations) > 0 {
		err = bs.NodeRegistrationService.ExpelNodes(nodeRegistrations, block.Height)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockService) updatePopScore(popScore int64, block *model.Block) error {
	var blocksmithNode model.NodeRegistration
	blocksmithNodeIDQ := bs.NodeRegistrationQuery.GetNodeRegistrationByNodePublicKey()
	row := bs.QueryExecutor.ExecuteSelectRow(blocksmithNodeIDQ, block.BlocksmithPublicKey)
	err := bs.NodeRegistrationQuery.Scan(&blocksmithNode, row)
	if err != nil {
		return err
	}
	addParticipationScoreQueries := bs.ParticipationScoreQuery.AddParticipationScore(
		blocksmithNode.NodeID, popScore, block.Height)
	err = bs.QueryExecutor.ExecuteTransactions(addParticipationScoreQueries)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BlockService) processPublishedReceipts(block *model.Block) (int, error) {
	var linkedCount int
	if len(block.GetPublishedReceipts()) > 0 {
		for index, rc := range block.GetPublishedReceipts() {
			// validate the receipts
			unsignedBytes := util.GetUnsignedBatchReceiptBytes(rc.BatchReceipt)
			if !bs.Signature.VerifyNodeSignature(
				unsignedBytes,
				rc.BatchReceipt.RecipientSignature,
				rc.BatchReceipt.RecipientPublicKey,
			) {
				// rollback
				return 0, blocker.NewBlocker(
					blocker.ValidationErr,
					"InvalidReceiptSignature",
				)
			}
			// check if linked
			if rc.IntermediateHashes != nil && len(rc.IntermediateHashes) > 0 {
				var publishedReceipt = &model.PublishedReceipt{
					BatchReceipt:       &model.BatchReceipt{},
					IntermediateHashes: nil,
					BlockHeight:        0,
					ReceiptIndex:       0,
				}
				merkle := &commonUtils.MerkleRoot{}
				rcByte := util.GetSignedBatchReceiptBytes(rc.BatchReceipt)
				rcHash := sha3.Sum256(rcByte)
				root, err := merkle.GetMerkleRootFromIntermediateHashes(
					rcHash[:],
					rc.ReceiptIndex,
					merkle.RestoreIntermediateHashes(rc.IntermediateHashes),
				)
				if err != nil {
					return 0, err
				}
				// look up root in published_receipt table
				rcQ, rcArgs := bs.PublishedReceiptQuery.GetPublishedReceiptByLinkedRMR(root)
				row := bs.QueryExecutor.ExecuteSelectRow(rcQ, rcArgs...)
				err = bs.PublishedReceiptQuery.Scan(publishedReceipt, row)
				if err != nil {
					return 0, err
				}
				// add to linked receipt count for calculation later
				linkedCount++
			}
			// store in database
			// assign index and height, index is the order of the receipt in the block,
			// it's different with receiptIndex which is used to validate merkle root.
			rc.BlockHeight, rc.PublishedIndex = block.Height, uint32(index)
			insertPublishedReceiptQ, insertPublishedReceiptArgs := bs.PublishedReceiptQuery.InsertPublishedReceipt(
				rc,
			)
			err := bs.QueryExecutor.ExecuteTransaction(insertPublishedReceiptQ, insertPublishedReceiptArgs...)
			if err != nil {
				return 0, err
			}
		}
	}
	return linkedCount, nil
}

// CoinbaseLotteryWinners get the current list of blocksmiths, duplicate it (to not change the original one)
// and sort it using the NodeOrder algorithm. The first n (n = constant.MaxNumBlocksmithRewards) in the newly ordered list
// are the coinbase lottery winner (the blocksmiths that will be rewarded for the current block)
func (bs *BlockService) CoinbaseLotteryWinners() ([]string, error) {
	var (
		selectedAccounts []string
	)
	// copy the pointer array to not change original order
	blocksmiths := make([]model.Blocksmith, len(*bs.SortedBlocksmiths))
	copy(blocksmiths, *bs.SortedBlocksmiths)

	// sort blocksmiths by NodeOrder
	sort.SliceStable(blocksmiths, func(i, j int) bool {
		bi, bj := blocksmiths[i], blocksmiths[j]
		res := bi.NodeOrder.Cmp(bj.NodeOrder)
		if res == 0 {
			// compare node ID
			nodePKI := new(big.Int).SetUint64(uint64(bi.NodeID))
			nodePKJ := new(big.Int).SetUint64(uint64(bj.NodeID))
			res = nodePKI.Cmp(nodePKJ)
		}
		// ascending sort
		return res < 0
	})

	for idx, sortedBlockSmith := range blocksmiths {
		if idx > constant.MaxNumBlocksmithRewards-1 {
			break
		}
		// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
		qry, args := bs.NodeRegistrationQuery.GetNodeRegistrationByID(sortedBlockSmith.NodeID)
		rows, err := bs.QueryExecutor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		nr, err := bs.NodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		if (err != nil) || len(nr) == 0 {
			rows.Close()
			return nil, blocker.NewBlocker(blocker.DBErr, "CoinbaseLotteryNodeRegistrationNotFound")
		}
		selectedAccounts = append(selectedAccounts, nr[0].AccountAddress)
		rows.Close()
	}
	return selectedAccounts, nil
}

// RewardBlocksmithAccountAddresses accrue the block total fees + total coinbase to selected list of accounts
func (bs *BlockService) RewardBlocksmithAccountAddresses(
	blocksmithAccountAddresses []string,
	totalReward int64,
	height uint32) error {
	queries := make([][]interface{}, 0)
	if len(blocksmithAccountAddresses) == 0 {
		return blocker.NewBlocker(blocker.AppErr, "NoAccountToBeRewarded")
	}
	blocksmithReward := totalReward / int64(len(blocksmithAccountAddresses))
	for _, blocksmithAccountAddress := range blocksmithAccountAddresses {
		accountBalanceRecipientQ := bs.AccountBalanceQuery.AddAccountBalance(
			blocksmithReward,
			map[string]interface{}{
				"account_address": blocksmithAccountAddress,
				"block_height":    height,
			},
		)
		queries = append(queries, accountBalanceRecipientQ...)
	}
	if err := bs.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}

// GetBlockByID return the last pushed block
func (bs *BlockService) GetBlockByID(id int64) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByID(id), false)
	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				bs.Logger.Error(err.Error())
			}
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

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
			if err := rows.Close(); err != nil {
				bs.Logger.Error(err.Error())
			}
		}
	}()
	if err != nil {
		return []*model.Block{}, err
	}
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

	return blocks, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetLastBlock() (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetLastBlock(), false)
	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				bs.Logger.Error(err.Error())
			}
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

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
	return bs.TransactionQuery.BuildModel(transactions, rows)
}

func (bs *BlockService) GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	var publishedReceipts []*model.PublishedReceipt

	// get published receipts of the block
	publishedReceiptQ, publishedReceiptArg := bs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(blockHeight)
	rows, err := bs.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArg...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	publishedReceipts, err = bs.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return nil, err
	}
	return publishedReceipts, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetBlockByHeight(height uint32) (*model.Block, error) {
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockByHeight(height), false)
	defer func() {
		if rows != nil {
			if err := rows.Close(); err != nil {
				bs.Logger.Error(err.Error())
			}
		}
	}()
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	var blocks []*model.Block
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

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
			if err := rows.Close(); err != nil {
				bs.Logger.Error(err.Error())
			}
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
			if err := rows.Close(); err != nil {
				bs.Logger.Error(err.Error())
			}
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
	bs.Logger.Infof("mempool transaction with IDs = %s deleted", idsStr)
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
		payloadLength                        uint32
		// only for mainchain
		sortedTransactions  []*model.Transaction
		publishedReceipts   []*model.PublishedReceipt
		payloadHash         []byte
		err                 error
		digest              = sha3.New256()
		blockSmithPublicKey = util.GetPublicKeyFromSeed(secretPhrase)
	)
	newBlockHeight := previousBlock.Height + 1
	if _, ok := bs.Chaintype.(*chaintype.MainChain); ok {
		totalCoinbase = bs.GetCoinbase()
		sortedTransactions, err = bs.MempoolService.SelectTransactionsFromMempool(timestamp)
		if err != nil {
			return nil, errors.New("MempoolReadError")
		}
		for _, tx := range sortedTransactions {
			if _, err := digest.Write(tx.TransactionHash); err != nil {
				return nil, err
			}
			txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
			if err != nil {
				return nil, err
			}
			totalAmount += txType.GetAmount()
			totalFee += tx.Fee
			payloadLength += txType.GetSize()
		}
		publishedReceipts, err = bs.ReceiptService.SelectReceipts(timestamp, constant.ReceiptNumberToPick, previousBlock.Height)
		if err != nil {
			return nil, err
		}
		for _, br := range publishedReceipts {
			_, err = digest.Write(util.GetSignedBatchReceiptBytes(br.BatchReceipt))
			if err != nil {
				return nil, err
			}
		}
		payloadHash = digest.Sum([]byte{})
	}
	// loop through transaction to build block hash
	digest.Reset() // reset the digest
	if _, err := digest.Write(previousBlock.GetBlockSeed()); err != nil {
		return nil, err
	}

	previousSeedHash := digest.Sum([]byte{})
	blockSeed := bs.Signature.SignByNode(previousSeedHash, secretPhrase)
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
		newBlockHeight,
		timestamp,
		totalAmount,
		totalFee,
		totalCoinbase,
		sortedTransactions,
		publishedReceipts,
		payloadHash,
		payloadLength,
		secretPhrase,
	)
	return block, nil
}

// GenerateGenesisBlock generate and return genesis block from a given template (see constant/genesis.go)
func (bs *BlockService) GenerateGenesisBlock(genesisEntries []constant.MainchainGenesisConfigEntry) (*model.Block, error) {
	var (
		totalAmount, totalFee, totalCoinBase int64
		blockTransactions                    []*model.Transaction
		payloadLength                        uint32
		digest                               = sha3.New256()
	)
	genesisTransactions, err := GetGenesisTransactions(bs.Chaintype, genesisEntries)
	if err != nil {
		return nil, err
	}
	for index, tx := range genesisTransactions {
		if _, err := digest.Write(tx.TransactionHash); err != nil {
			return nil, err
		}
		if tx.TransactionType == util.ConvertBytesToUint32([]byte{1, 0, 0, 0}) { // if type = send money
			totalAmount += tx.GetSendMoneyTransactionBody().Amount
		}
		txType, err := bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}
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
		0,
		constant.MainchainGenesisBlockTimestamp,
		totalAmount,
		totalFee,
		totalCoinBase,
		blockTransactions,
		[]*model.PublishedReceipt{},
		payloadHash,
		payloadLength,
		constant.InitialSmithScale,
		big.NewInt(0),
		constant.MainchainGenesisBlockSignature,
	)
	// assign genesis block id
	block.ID = coreUtil.GetBlockID(block)
	return block, nil
}

// AddGenesis generate and add (push) genesis block to db
func (bs *BlockService) AddGenesis() error {
	block, err := bs.GenerateGenesisBlock(constant.MainChainGenesisConfig)
	if err != nil {
		return err
	}
	err = bs.PushBlock(&model.Block{ID: -1, Height: 0}, block, true, false)
	if err != nil {
		bs.Logger.Fatal("PushGenesisBlock:fail ", err)
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
		bs.Logger.Fatalf("Genesis ID does not match, expect: %d, get: %d", bs.Chaintype.GetGenesisBlockID(), genesisBlock.ID)
	}
	return true
}

// ReceiveBlock handle the block received from connected peers
func (bs *BlockService) ReceiveBlock(
	senderPublicKey []byte,
	lastBlock, block *model.Block,
	nodeSecretPhrase string,
) (*model.BatchReceipt, error) {
	// make sure block has previous block hash
	if block.GetPreviousBlockHash() == nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			"last block hash does not exist",
		)
	}
	// check signature of the incoming block
	blockUnsignedByte, err := util.GetBlockByte(block, false)
	if err != nil {
		return nil, err
	}
	if !bs.Signature.VerifyNodeSignature(blockUnsignedByte, block.BlockSignature, block.BlocksmithPublicKey) {
		return nil, blocker.NewBlocker(
			blocker.ValidationErr,
			"block signature invalid")
	}
	blockHash, err := util.GetBlockHash(block)
	if err != nil {
		return nil, err
	}
	// check previous block hash
	lastBlockByte, err := util.GetBlockByte(lastBlock, true)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			err.Error(),
		)
	}
	lastBlockHash := sha3.Sum256(lastBlockByte)
	receiptKey, err := commonUtils.GetReceiptKey(
		blockHash, senderPublicKey,
	)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			err.Error(),
		)
	}
	//  check equality last block hash with previous block hash from received block
	if !bytes.Equal(lastBlockHash[:], block.PreviousBlockHash) {
		// check if already broadcast receipt to this node
		_, err := bs.KVExecutor.Get(constant.KVdbTableBlockReminderKey + string(receiptKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
					blockHash,
					lastBlock,
					senderPublicKey,
					nodeSecretPhrase,
					constant.KVdbTableBlockReminderKey+string(receiptKey),
					constant.ReceiptDatumTypeBlock,
					bs.Signature,
					bs.QueryExecutor,
					bs.KVExecutor,
				)
				if err != nil {
					return nil, err
				}
				return batchReceipt, nil
			}
			return nil, blocker.NewBlocker(
				blocker.DBErr,
				err.Error(),
			)
		}
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			"previous block hash does not match with last block hash",
		)
	}
	// check if the block broadcaster is the valid blocksmith
	index := -1 // use index to determine if is in list, and who to punish
	for i, bs := range *bs.SortedBlocksmiths {
		if reflect.DeepEqual(bs.NodePublicKey, block.BlocksmithPublicKey) {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, blocker.NewBlocker(
			blocker.BlockErr, "invalid blocksmith")
	}
	// base on index we can calculate punishment and reward
	err = bs.PushBlock(lastBlock, block, true, true)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}
	// generate receipt and return as response
	batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
		blockHash,
		lastBlock,
		senderPublicKey,
		nodeSecretPhrase,
		constant.KVdbTableBlockReminderKey+string(receiptKey),
		constant.ReceiptDatumTypeBlock,
		bs.Signature,
		bs.QueryExecutor,
		bs.KVExecutor,
	)
	if err != nil {
		return nil, err
	}
	return batchReceipt, nil
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
	participationScores, err = bs.ParticipationScoreQuery.BuildModel(participationScores, rows)
	// if there aren't participation scores for this address/node, return 0
	if (err != nil) || len(participationScores) == 0 {
		return 0, nil
	}
	return participationScores[0].Score, nil
}

// GetParticipationScore handle received block from another node
func (bs *BlockService) GetBlockExtendedInfo(block *model.Block) (*model.BlockExtendedInfo, error) {
	var (
		blExt = &model.BlockExtendedInfo{}
		err   error
	)
	blExt.Block = block
	// block extra (computed) info
	if block.Height > 0 {
		blExt.BlocksmithAccountAddress, err = bs.GetBlocksmithAccountAddress(block)
		if err != nil {
			return nil, err
		}
	} else {
		blExt.BlocksmithAccountAddress = constant.MainchainGenesisAccountAddress
	}
	// Total number of receipts at a block height
	// STEF: do we need to get all receipts that have reference_block_height <= block.height
	blExt.TotalReceipts = 99
	//TODO: from @barton: Receipt value will be the "score" of all the receipts in a block added together
	// STEF: how to compute the receipt score?
	blExt.ReceiptValue = 99
	// once we have the receipt for this blExt we should be able to calculate this using util.CalculateParticipationScore
	blExt.PopChange = -20

	return blExt, nil
}

func (bs *BlockService) GetBlocksmithAccountAddress(block *model.Block) (string, error) {
	var (
		nr []*model.NodeRegistration
	)
	// get node registration related to current BlockSmith to retrieve the node's owner account at the block's height
	qry, args := bs.NodeRegistrationQuery.GetLastVersionedNodeRegistrationByPublicKey(block.BlocksmithPublicKey, block.Height)
	rows, err := bs.QueryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return "", blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	nr, err = bs.NodeRegistrationQuery.BuildModel(nr, rows)
	if (err != nil) || len(nr) == 0 {
		return "", blocker.NewBlocker(blocker.DBErr, "VersionedNodeRegistrationNotFound")
	}
	return nr[0].AccountAddress, nil
}

func (*BlockService) GetCoinbase() int64 {
	return 50 * constant.OneZBC
}

// GetBlocksmiths select the blocksmiths for a given block and calculate the SmithOrder (for smithing) and NodeOrder (for block rewards)
func (bs *BlockService) GetBlocksmiths(block *model.Block) ([]*model.Blocksmith, error) {
	var (
		activeBlocksmiths, blocksmiths []*model.Blocksmith
	)
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.NodeRegistrationQuery.GetActiveNodeRegistrations(), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activeBlocksmiths = bs.NodeRegistrationQuery.BuildBlocksmith(activeBlocksmiths, rows)
	// add smithorder and nodeorder to be used to select blocksmith and coinbase rewards
	blockSeed := new(big.Int).SetBytes(block.BlockSeed)
	for _, blocksmith := range activeBlocksmiths {
		blocksmith.SmithOrder = coreUtil.CalculateSmithOrder(blocksmith.Score, blockSeed, blocksmith.NodeID)
		blocksmith.NodeOrder = coreUtil.CalculateNodeOrder(blocksmith.Score, blockSeed, blocksmith.NodeID)
		blocksmith.BlockSeed = blockSeed
		blocksmiths = append(blocksmiths, blocksmith)
	}
	return blocksmiths, nil
}
