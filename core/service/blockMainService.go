package service

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	BlockServiceMainInterface interface {
		NewMainBlock(
			version uint32,
			previousBlockHash, blockSeed, blockSmithPublicKey []byte,
			previousBlockHeight uint32,
			timestamp, totalAmount, totalFee, totalCoinBase int64,
			transactions []*model.Transaction,
			blockReceipts []*model.PublishedReceipt,
			payloadHash []byte,
			payloadLength uint32,
			secretPhrase string,
		) (*model.Block, error)
		GetCoinbase() int64
		CoinbaseLotteryWinners(sortedBlocksmith []*model.Blocksmith) ([]string, error)
		RewardBlocksmithAccountAddresses(blocksmithAccountAddresses []string, totalReward int64, height uint32) error
		GetBlocksmithAccountAddress(block *model.Block) (string, error)
		GetParticipationScore(nodePublicKey []byte) (int64, error)
		GetTransactionsByBlockID(blockID int64) ([]*model.Transaction, error)
		GetPublishedReceiptsByBlockHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		RemoveMempoolTransactions(transactions []*model.Transaction) error
		CleanTheTimedoutBlock()
	}

	//TODO: rename to BlockMainService
	BlockService struct {
		sync.RWMutex
		Chaintype                    chaintype.ChainType
		KVExecutor                   kvdb.KVExecutorInterface
		QueryExecutor                query.ExecutorInterface
		BlockQuery                   query.BlockQueryInterface
		MempoolQuery                 query.MempoolQueryInterface
		TransactionQuery             query.TransactionQueryInterface
		MerkleTreeQuery              query.MerkleTreeQueryInterface
		PublishedReceiptQuery        query.PublishedReceiptQueryInterface
		SkippedBlocksmithQuery       query.SkippedBlocksmithQueryInterface
		SpinePublicKeyQuery          query.SpinePublicKeyQueryInterface
		Signature                    crypto.SignatureInterface
		MempoolService               MempoolServiceInterface
		ReceiptService               ReceiptServiceInterface
		NodeRegistrationService      NodeRegistrationServiceInterface
		ActionTypeSwitcher           transaction.TypeActionSwitcher
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		AccountLedgerQuery           query.AccountLedgerQueryInterface
		BlocksmithStrategy           strategy.BlocksmithStrategyInterface
		WaitingTransactionBlockQueue WaitingTransactionBlockQueue
		Observer                     *observer.Observer
		Logger                       *log.Logger
	}
	// TransactionIDs reperesent a list of transaction id will used by queued block
	TransactionIDsMap map[int64]int
	BlockIDsMap       map[int64]bool
	// BlockWithMetaData is incoming block with some information while waiting transaction
	BlockWithMetaData struct {
		Block     *model.Block
		Timestamp int64
	}
	// WaitingTransactionBlockQueue reperesent a list of incoming blocks while waiting their transaction
	WaitingTransactionBlockQueue struct {
		// map of block ID with the blocks that have been received but waiting transactions to be completed
		WaitingTxBlocks map[int64]*BlockWithMetaData
		// map of blockID with an array of transactionIds it requires
		BlockRequiringTransactionsMap map[int64]TransactionIDsMap
		// map of transactionIds with blockIds that requires them
		TransactionsRequiredMap map[int64]BlockIDsMap
		BlockMutex              sync.Mutex
	}
)

// NewMainBlock generate new mainchain block
func (bs *BlockService) NewMainBlock(
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
) (*model.Block, error) {
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
	blockUnsignedByte, err := util.GetBlockByte(block, false, bs.Chaintype)
	if err != nil {
		bs.Logger.Error(err.Error())
	}
	block.BlockSignature = bs.Signature.SignByNode(blockUnsignedByte, secretPhrase)
	blockHash, err := util.GetBlockHash(block, bs.Chaintype)
	if err != nil {
		return nil, err
	}
	block.BlockHash = blockHash
	return block, nil
}

// GetChainType returns the chaintype
func (bs *BlockService) GetChainType() chaintype.ChainType {
	return bs.Chaintype
}

func (bs *BlockService) GetBlocksmithStrategy() strategy.BlocksmithStrategyInterface {
	return bs.BlocksmithStrategy
}

// ChainWriteLock locks the chain
func (bs *BlockService) ChainWriteLock(actionType int) {
	monitoring.IncrementStatusLockCounter(actionType)
	bs.Lock()
	monitoring.SetBlockchainStatus(bs.Chaintype.GetTypeInt(), actionType)
}

// ChainWriteUnlock unlocks the chain
func (bs *BlockService) ChainWriteUnlock(actionType int) {
	monitoring.SetBlockchainStatus(bs.Chaintype.GetTypeInt(), constant.BlockchainStatusIdle)
	monitoring.DecrementStatusLockCounter(actionType)
	bs.Unlock()
}

// NewGenesisBlock create new block that is fixed in the value of cumulative difficulty, smith scale, and the block signature
func (bs *BlockService) NewGenesisBlock(
	version uint32,
	previousBlockHash, blockSeed, blockSmithPublicKey []byte,
	previousBlockHeight uint32,
	timestamp, totalAmount, totalFee, totalCoinBase int64,
	transactions []*model.Transaction,
	publishedReceipts []*model.PublishedReceipt,
	spinePublicKeys []*model.SpinePublicKey,
	payloadHash []byte,
	payloadLength uint32,
	cumulativeDifficulty *big.Int,
	genesisSignature []byte,
) (*model.Block, error) {
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
		SpinePublicKeys:      spinePublicKeys,
		PublishedReceipts:    publishedReceipts,
		PayloadLength:        payloadLength,
		PayloadHash:          payloadHash,
		CumulativeDifficulty: cumulativeDifficulty.String(),
		BlockSignature:       genesisSignature,
	}
	blockHash, err := util.GetBlockHash(block, bs.Chaintype)
	if err != nil {
		return nil, err
	}
	block.BlockHash = blockHash
	return block, nil
}

// PreValidateBlock valdiate block without without it's transactions
func (bs *BlockService) PreValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	// check block timestamp
	if block.GetTimestamp() > curTime+constant.GenerateBlockTimeoutSec {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidTimestamp")
	}
	// check if blocksmith can smith at the time
	blocksmithsMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousLastBlock)
	blocksmithIndex := blocksmithsMap[string(block.BlocksmithPublicKey)]
	if blocksmithIndex == nil {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidBlocksmith")
	}
	// check smithtime
	blocksmithTime := bs.BlocksmithStrategy.GetSmithTime(*blocksmithIndex, previousLastBlock)
	if blocksmithTime > block.GetTimestamp() {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidSmithTime")
	}
	return nil
}

// ValidateBlock validate block to be pushed into the blockchain
func (bs *BlockService) ValidateBlock(block, previousLastBlock *model.Block, curTime int64) error {
	err := bs.PreValidateBlock(block, previousLastBlock, curTime)
	if err != nil {
		return err
	}
	if coreUtil.GetBlockID(block, bs.Chaintype) == 0 {
		return blocker.NewBlocker(blocker.BlockErr, "InvalidID")
	}
	// Verify Signature
	blockByte, err := commonUtils.GetBlockByte(block, false, bs.Chaintype)
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
	previousBlockHash, err := util.GetBlockHash(previousLastBlock, bs.Chaintype)
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

		// if cumulative difficulty of the reference block is > of the one of the (new) block, new block is invalid
		if refCumulativeDifficulty.Cmp(blockCumulativeDifficulty) > 0 {
			return blocker.NewBlocker(blocker.BlockErr, "InvalidCumulativeDifficulty")
		}
	}
	return nil
}

// PushBlock push block into blockchain, to broadcast the block after pushing to own node, switch the
// broadcast flag to `true`, and `false` otherwise
func (bs *BlockService) PushBlock(previousBlock, block *model.Block, broadcast bool) error {
	var (
		err error
	)
	if !coreUtil.IsGenesis(previousBlock.GetID(), block) {
		block.Height = previousBlock.GetHeight() + 1
		sortedBlocksmithMap := bs.BlocksmithStrategy.GetSortedBlocksmithsMap(previousBlock)
		blocksmithIndex := sortedBlocksmithMap[string(block.GetBlocksmithPublicKey())]
		if blocksmithIndex == nil {
			return blocker.NewBlocker(blocker.BlockErr, "BlocksmithNotInSmithingList")
		}
		blockCumulativeDifficulty, err := coreUtil.CalculateCumulativeDifficulty(
			previousBlock, *blocksmithIndex,
		)
		if err != nil {
			return err
		}
		block.CumulativeDifficulty = blockCumulativeDifficulty
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
	var transactionIDs = make([]int64, len(block.GetTransactions()))
	// apply transactions and remove them from mempool
	for index, tx := range block.GetTransactions() {
		// assign block id and block height to tx
		tx.BlockID = block.ID
		tx.Height = block.Height
		tx.TransactionIndex = uint32(index) + 1
		transactionIDs[index] = tx.GetID()
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
		err = txType.ApplyConfirmed(block.GetTimestamp())
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

	// Mainchain specific:
	// - Compute and update popscore
	// - Block reward
	// - Admit/Expel nodes to/from registry
	// - Build scrambled node registry
	if block.Height > 0 {
		// this is to manage the edge case when the blocksmith array has not been initialized yet:
		// when start smithing from a block with height > 0, since SortedBlocksmiths are computed  after a block is pushed,
		// for the first block that is pushed, we don't know who are the blocksmith to be rewarded
		// sort blocksmiths for current block
		popScore, err := commonUtils.CalculateParticipationScore(
			uint32(linkedCount),
			uint32(len(block.GetPublishedReceipts())-linkedCount),
			coreUtil.GetNumberOfMaxReceipts(len(bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))),
		)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		err = bs.updatePopScore(popScore, previousBlock, block)
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}

		// selecting multiple account to be rewarded and split the total coinbase + totalFees evenly between them
		totalReward := block.TotalFee + block.TotalCoinBase
		lotteryAccounts, err := bs.CoinbaseLotteryWinners(bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))
		if err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		if err := bs.RewardBlocksmithAccountAddresses(
			lotteryAccounts,
			totalReward,
			block.GetTimestamp(),
			block.Height,
		); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	// admit nodes from registry at genesis and regular intervals
	// expel nodes from node registry as soon as they reach zero participation score
	if err := bs.expelNodes(block); err != nil {
		if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
			bs.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	if block.Height == 0 || block.Height%bs.NodeRegistrationService.GetNodeAdmittanceCycle() == 0 {
		if err := bs.admitNodes(block); err != nil {
			if rollbackErr := bs.QueryExecutor.RollbackTx(); rollbackErr != nil {
				bs.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}
	// building scrambled node registry
	if block.GetHeight() == bs.NodeRegistrationService.GetBlockHeightToBuildScrambleNodes(block.GetHeight()) {
		err = bs.NodeRegistrationService.BuildScrambledNodes(block)
		if err != nil {
			bs.Logger.Error(err.Error())
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
	bs.Logger.Debugf("%s Block Pushed ID: %d", bs.Chaintype.GetName(), block.GetID())
	// sort blocksmiths for next block
	bs.BlocksmithStrategy.SortBlocksmiths(block)
	// add transactionIDs and remove transaction before broadcast
	block.TransactionIDs = transactionIDs
	block.Transactions = []*model.Transaction{}
	// broadcast block
	if broadcast {
		bs.Observer.Notify(observer.BroadcastBlock, block, bs.Chaintype)
	}
	bs.Observer.Notify(observer.BlockPushed, block, bs.Chaintype)
	monitoring.SetLastBlock(bs.Chaintype.GetTypeInt(), block)
	return nil
}

// adminNodes seelct and admit nodes from node registry
func (bs *BlockService) admitNodes(block *model.Block) error {
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

	return nil
}

// expelNodes seelct and expel nodes from node registry
func (bs *BlockService) expelNodes(block *model.Block) error {
	nodeRegistrations, err := bs.NodeRegistrationService.SelectNodesToBeExpelled()
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

func (bs *BlockService) updatePopScore(popScore int64, previousBlock, block *model.Block) error {
	var (
		blocksmithNode  *model.Blocksmith
		blocksmithIndex = -1
		err             error
	)
	for i, bsm := range bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock) {
		if reflect.DeepEqual(block.BlocksmithPublicKey, bsm.NodePublicKey) {
			blocksmithIndex = i
			blocksmithNode = bsm
			break
		}
	}
	if blocksmithIndex < 0 {
		return blocker.NewBlocker(blocker.BlockErr, "BlocksmithNotInBlocksmithList")
	}
	// punish the skipped (index earlier than current blocksmith) blocksmith
	for i, bsm := range (bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))[:blocksmithIndex] {
		skippedBlocksmith := &model.SkippedBlocksmith{
			BlocksmithPublicKey: bsm.NodePublicKey,
			POPChange:           constant.ParticipationScorePunishAmount,
			BlockHeight:         block.Height,
			BlocksmithIndex:     int32(i),
		}
		// store to skipped_blocksmith table
		qStr, args := bs.SkippedBlocksmithQuery.InsertSkippedBlocksmith(
			skippedBlocksmith,
		)
		err = bs.QueryExecutor.ExecuteTransaction(qStr, args...)
		if err != nil {
			return err
		}
		// punish score
		_, err = bs.NodeRegistrationService.AddParticipationScore(bsm.NodeID, constant.ParticipationScorePunishAmount, block.Height, true)
		if err != nil {
			return err
		}
	}
	_, err = bs.NodeRegistrationService.AddParticipationScore(blocksmithNode.NodeID, popScore, block.Height, true)
	if err != nil {
		return err
	}
	return nil
}

// processPublishedReceipts process the receipt received in a block
// todo: this should be moved to PublishedReceiptService
func (bs *BlockService) processPublishedReceipts(block *model.Block) (int, error) {
	var (
		linkedCount int
		err         error
	)
	if len(block.GetPublishedReceipts()) > 0 {
		for index, rc := range block.GetPublishedReceipts() {
			// validate sender and recipient of receipt
			err = bs.ReceiptService.ValidateReceipt(rc.BatchReceipt)
			if err != nil {
				return 0, err
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
				row, _ := bs.QueryExecutor.ExecuteSelectRow(rcQ, false, rcArgs...)
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
func (bs *BlockService) CoinbaseLotteryWinners(blocksmiths []*model.Blocksmith) ([]string, error) {
	var (
		selectedAccounts []string
	)
	// copy the pointer array to not change original order

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
	totalReward, blockTimestamp int64,
	height uint32,
) error {
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

		accountLedgerQ, accountLedgerArgs := bs.AccountLedgerQuery.InsertAccountLedger(&model.AccountLedger{
			AccountAddress: blocksmithAccountAddress,
			BalanceChange:  blocksmithReward,
			BlockHeight:    height,
			EventType:      model.EventType_EventReward,
			Timestamp:      uint64(blockTimestamp),
		})

		accountLedgerArgs = append([]interface{}{accountLedgerQ}, accountLedgerArgs...)
		queries = append(queries, accountLedgerArgs)
	}
	if err := bs.QueryExecutor.ExecuteTransactions(queries); err != nil {
		return err
	}
	return nil
}

// GetBlockByID return a block by its ID
// withAttachedData if true returns extra attached data for the block (transactions)
func (bs *BlockService) GetBlockByID(id int64, withAttachedData bool) (*model.Block, error) {
	var (
		block model.Block
	)
	row, err := bs.QueryExecutor.ExecuteSelectRow(bs.BlockQuery.GetBlockByID(id), false)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if err = bs.BlockQuery.Scan(&block, row); err != nil {
		if err == sql.ErrNoRows {
			return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, err.Error())
		}
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

	if block.ID != 0 {
		if withAttachedData {
			transactions, err := bs.GetTransactionsByBlockID(block.ID)
			if err != nil {
				return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
			}
			block.Transactions = transactions
		}
		return &block, nil
	}
	return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("block %v is not found", id))
}

// GetBlocksFromHeight get all blocks from a given height till last block (or a given limit is reached).
// Note: this only returns main block data, it doesn't populate attached data (transactions, receipts)
func (bs *BlockService) GetBlocksFromHeight(startHeight, limit uint32) ([]*model.Block, error) {
	var blocks []*model.Block
	rows, err := bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlockFromHeight(startHeight, limit), false)
	if err != nil {
		return []*model.Block{}, err
	}
	defer rows.Close()
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, "failed to build model")
	}

	return blocks, nil
}

// GetLastBlock return the last pushed block
func (bs *BlockService) GetLastBlock() (*model.Block, error) {
	lastBlock, err := commonUtils.GetLastBlock(bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	transactions, err := bs.GetTransactionsByBlockID(lastBlock.ID)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	lastBlock.Transactions = transactions
	return lastBlock, nil
}

// GetBlockHash return block's hash (makes sure always include transactions)
func (bs *BlockService) GetBlockHash(block *model.Block) ([]byte, error) {
	transactions, err := bs.GetTransactionsByBlockID(block.ID)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.Transactions = transactions
	return commonUtils.GetBlockHash(block, bs.GetChainType())

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
	defer rows.Close()

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
	block, err := commonUtils.GetBlockByHeight(height, bs.QueryExecutor, bs.BlockQuery)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	transactions, err := bs.GetTransactionsByBlockID(block.ID)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.Transactions = transactions

	return block, nil
}

// GetGenesis return the last pushed block
func (bs *BlockService) GetGenesisBlock() (*model.Block, error) {
	var (
		lastBlock model.Block
		row, _    = bs.QueryExecutor.ExecuteSelectRow(bs.BlockQuery.GetGenesisBlock(), false)
	)
	if row == nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
	}
	err := bs.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.BlockNotFoundErr, "genesis block is not found")
	}
	return &lastBlock, nil
}

// GetBlocks return all pushed blocks
func (bs *BlockService) GetBlocks() ([]*model.Block, error) {
	var (
		blocks    []*model.Block
		rows, err = bs.QueryExecutor.ExecuteSelect(bs.BlockQuery.GetBlocks(0, 100), false)
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks, err = bs.BlockQuery.BuildModel(blocks, rows)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// PopulateBlockData add transactions and published receipts to model.Block instance
func (bs *BlockService) PopulateBlockData(block *model.Block) error {
	txs, err := bs.GetTransactionsByBlockID(block.ID)
	if err != nil {
		bs.Logger.Errorln(err)
		return blocker.NewBlocker(blocker.BlockErr, "error getting block transactions")
	}
	prs, err := bs.GetPublishedReceiptsByBlockHeight(block.Height)
	if err != nil {
		bs.Logger.Errorln(err)
		return blocker.NewBlocker(blocker.BlockErr, "error getting block published receipts")
	}
	block.Transactions = txs
	block.PublishedReceipts = prs
	return nil
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
	// calculate total coinbase to be added to the block
	totalCoinbase = bs.GetCoinbase()
	sortedTransactions, err = bs.MempoolService.SelectTransactionsFromMempool(timestamp)
	if err != nil {
		return nil, errors.New("MempoolReadError")
	}
	// select transactions from mempool to be added to the block
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
	// select published receipts to be added to the block
	publishedReceipts, err = bs.ReceiptService.SelectReceipts(
		timestamp, coreUtil.GetNumberOfMaxReceipts(
			len(bs.BlocksmithStrategy.GetSortedBlocksmiths(previousBlock))),
		previousBlock.Height,
	)
	// FIXME: add published receipts to block payload length

	if err != nil {
		return nil, err
	}
	// filter only good receipt
	for _, br := range publishedReceipts {
		_, err = digest.Write(util.GetSignedBatchReceiptBytes(br.BatchReceipt))
		if err != nil {
			return nil, err
		}
	}

	payloadHash = digest.Sum([]byte{})
	// loop through transaction to build block hash
	digest.Reset() // reset the digest
	if _, err := digest.Write(previousBlock.GetBlockSeed()); err != nil {
		return nil, err
	}

	previousSeedHash := digest.Sum([]byte{})
	blockSeed := bs.Signature.SignByNode(previousSeedHash, secretPhrase)
	digest.Reset() // reset the digest
	// compute the previous block hash
	previousBlockHash, err := util.GetBlockHash(previousBlock, bs.Chaintype)
	if err != nil {
		return nil, err
	}
	block, err := bs.NewMainBlock(
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
	if err != nil {
		return nil, err
	}
	return block, nil
}

// GenerateGenesisBlock generate and return genesis block from a given template (see constant/genesis.go)
func (bs *BlockService) GenerateGenesisBlock(genesisEntries []constant.GenesisConfigEntry) (*model.Block, error) {
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
	block, err := bs.NewGenesisBlock(
		1,
		nil,
		bs.Chaintype.GetGenesisBlockSeed(),
		bs.Chaintype.GetGenesisNodePublicKey(),
		0,
		bs.Chaintype.GetGenesisBlockTimestamp(),
		totalAmount,
		totalFee,
		totalCoinBase,
		blockTransactions,
		[]*model.PublishedReceipt{},
		nil,
		payloadHash,
		payloadLength,
		big.NewInt(0),
		bs.Chaintype.GetGenesisBlockSignature(),
	)
	if err != nil {
		return nil, err
	}
	// assign genesis block id
	block.ID = coreUtil.GetBlockID(block, bs.Chaintype)
	if block.ID == 0 {
		return nil, blocker.NewBlocker(blocker.BlockErr, fmt.Sprintf("Invalid %s Genesis Block ID", bs.Chaintype.GetName()))
	}
	return block, nil
}

// AddGenesis generate and add (push) genesis block to db
func (bs *BlockService) AddGenesis() error {
	block, err := bs.GenerateGenesisBlock(constant.GenesisConfig)
	if err != nil {
		return err
	}
	err = bs.PushBlock(&model.Block{ID: -1, Height: 0}, block, false)
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
// argument lastBlock is the lastblock in this node
// argument block is the in coming block from peer
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
	receiptKey, err := commonUtils.GetReceiptKey(
		block.GetBlockHash(), senderPublicKey,
	)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.BlockErr,
			err.Error(),
		)
	}
	var isQueued bool
	//  check equality last block hash with previous block hash from received block
	if !bytes.Equal(lastBlock.GetBlockHash(), block.GetPreviousBlockHash()) {
		// when incoming block is better than last block that having same prevoius block hash
		if bytes.Equal(lastBlock.GetPreviousBlockHash(), block.PreviousBlockHash) &&
			block.Timestamp < lastBlock.Timestamp {
			var previousBlock, err = commonUtils.GetBlockByHeight(lastBlock.Height-1, bs.QueryExecutor, bs.BlockQuery)
			if err != nil {
				return nil, status.Error(codes.Internal, "FailGetBlock")
			}
			// pre validation block
			if err = bs.PreValidateBlock(block, previousBlock, time.Now().Unix()); err != nil {
				return nil, status.Error(codes.InvalidArgument, "InvalidBlock")
			}
			isQueued, err = bs.ProcessQueuedBlock(block)
			if err != nil {
				return nil, err
			}
			// when isQueued is false and err is nil it means need to process completed block
			if !isQueued {
				err = bs.ProcessCompletedBlock(block)
				if err != nil {
					return nil, err
				}
			}
		}

		// check if already broadcast receipt to this node
		_, err = bs.KVExecutor.Get(constant.KVdbTableBlockReminderKey + string(receiptKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				blockHash, err := commonUtils.GetBlockHash(block, bs.Chaintype)
				if err != nil {
					return nil, err
				}
				if !bytes.Equal(blockHash, lastBlock.GetBlockHash()) {
					// invalid block hash don't send receipt to client
					return nil, status.Error(codes.InvalidArgument, "InvalidBlockHash")
				}
				batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
					bs.Chaintype,
					block.GetBlockHash(),
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
					return nil, status.Error(codes.Internal, err.Error())
				}
				return batchReceipt, nil
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument,
			"previousBlockHashDoesNotMatchWithLastBlockHash",
		)
	}

	// pre validation block
	if err = bs.PreValidateBlock(block, lastBlock, time.Now().Unix()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "InvalidBlock")
	}
	isQueued, err = bs.ProcessQueuedBlock(block)
	if err != nil {
		return nil, err
	}
	// precess block when block don't have transaction
	if !isQueued {
		err = bs.ProcessCompletedBlock(block)
		if err != nil {
			return nil, err
		}
	}

	// generate receipt and return as response
	batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
		bs.Chaintype,
		block.GetBlockHash(),
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
		return nil, status.Error(codes.Internal, err.Error())
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
	defer rows.Close()
	participationScores, err = bs.ParticipationScoreQuery.BuildModel(participationScores, rows)
	// if there aren't participation scores for this address/node, return 0
	if (err != nil) || len(participationScores) == 0 {
		return 0, nil
	}
	return participationScores[0].Score, nil
}

// GetParticipationScore handle received block from another node
func (bs *BlockService) GetBlockExtendedInfo(block *model.Block, includeReceipts bool) (*model.BlockExtendedInfo, error) {
	var (
		blExt                         = &model.BlockExtendedInfo{}
		skippedBlocksmiths            []*model.SkippedBlocksmith
		publishedReceipts             []*model.PublishedReceipt
		nodeRegistryAtHeight          []*model.NodeRegistration
		linkedPublishedReceiptCount   uint32
		unLinkedPublishedReceiptCount uint32
		err                           error
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
	skippedBlocksmithsQuery := bs.SkippedBlocksmithQuery.GetSkippedBlocksmithsByBlockHeight(block.Height)
	skippedBlocksmithsRows, err := bs.QueryExecutor.ExecuteSelect(skippedBlocksmithsQuery, false)
	if err != nil {
		return nil, err
	}
	defer skippedBlocksmithsRows.Close()
	blExt.SkippedBlocksmiths, err = bs.SkippedBlocksmithQuery.BuildModel(skippedBlocksmiths, skippedBlocksmithsRows)
	if err != nil {
		return nil, err
	}
	publishedReceiptQ, publishedReceiptArgs := bs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(block.Height)
	publishedReceiptRows, err := bs.QueryExecutor.ExecuteSelect(publishedReceiptQ, false, publishedReceiptArgs...)
	if err != nil {
		return nil, err
	}
	defer publishedReceiptRows.Close()
	publishedReceipts, err = bs.PublishedReceiptQuery.BuildModel(publishedReceipts, publishedReceiptRows)
	if err != nil {
		return nil, err
	}
	blExt.TotalReceipts = int64(len(publishedReceipts))
	for _, pr := range publishedReceipts {
		if pr.IntermediateHashes != nil {
			linkedPublishedReceiptCount++
		} else {
			unLinkedPublishedReceiptCount++
		}
	}
	nodeRegistryAtHeightQ := bs.NodeRegistrationQuery.GetNodeRegistryAtHeight(block.Height)
	nodeRegistryAtHeightRows, err := bs.QueryExecutor.ExecuteSelect(nodeRegistryAtHeightQ, false)
	if err != nil {
		return nil, err
	}
	defer nodeRegistryAtHeightRows.Close()
	nodeRegistryAtHeight, err = bs.NodeRegistrationQuery.BuildModel(nodeRegistryAtHeight, nodeRegistryAtHeightRows)
	if err != nil {
		return nil, err
	}
	blExt.ReceiptValue = commonUtils.GetReceiptValue(linkedPublishedReceiptCount, unLinkedPublishedReceiptCount)
	blExt.PopChange, err = util.CalculateParticipationScore(
		linkedPublishedReceiptCount,
		unLinkedPublishedReceiptCount,
		coreUtil.GetNumberOfMaxReceipts(len(nodeRegistryAtHeight)),
	)
	if err != nil {
		return nil, err
	}

	if includeReceipts {
		blExt.Block.PublishedReceipts = publishedReceipts
	}

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

func (bs *BlockService) PopOffToBlock(commonBlock *model.Block) ([]*model.Block, error) {
	var (
		mempoolsBackupBytes *bytes.Buffer
		mempoolsBackup      []*model.MempoolTransaction
		err                 error
	)
	// if current blockchain Height is lower than minimal height of the blockchain that is allowed to rollback
	lastBlock, err := bs.GetLastBlock()
	if err != nil {
		return []*model.Block{}, err
	}
	minRollbackHeight := util.GetMinRollbackHeight(lastBlock.Height)

	if commonBlock.Height < minRollbackHeight {
		// TODO: handle it appropriately and analyze the effect if this returning empty element in the further processfork process
		bs.Logger.Warn("the node blockchain detects hardfork, please manually delete the database to recover")
		return []*model.Block{}, nil
	}

	_, err = bs.GetBlockByID(commonBlock.ID, false)
	if err != nil {
		return []*model.Block{}, blocker.NewBlocker(blocker.BlockNotFoundErr, fmt.Sprintf("the common block is not found %v", commonBlock.ID))
	}

	var poppedBlocks []*model.Block
	block := lastBlock

	// TODO:
	// Need to refactor this codes with better solution in the future
	// https://github.com/zoobc/zoobc-core/pull/514#discussion_r355297318
	publishedReceipts, err := bs.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	block.PublishedReceipts = publishedReceipts

	for block.ID != commonBlock.ID && block.ID != bs.Chaintype.GetGenesisBlockID() {
		poppedBlocks = append(poppedBlocks, block)
		block, err = bs.GetBlockByHeight(block.Height - 1)
		if err != nil {
			return nil, err
		}
		publishedReceipts, err := bs.ReceiptService.GetPublishedReceiptsByHeight(block.GetHeight())
		if err != nil {
			return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
		}
		block.PublishedReceipts = publishedReceipts
	}

	// Backup existing transactions from mempool before rollback
	mempoolsBackup, err = bs.MempoolService.GetMempoolTransactionsWantToBackup(commonBlock.Height)
	if err != nil {
		return nil, err
	}
	bs.Logger.Warnf("mempool tx backup %d in total with block_height %d", len(mempoolsBackup), commonBlock.GetHeight())
	derivedQueries := query.GetDerivedQuery(bs.Chaintype)
	err = bs.QueryExecutor.BeginTx()
	if err != nil {
		return []*model.Block{}, err
	}

	for _, dQuery := range derivedQueries {
		queries := dQuery.Rollback(commonBlock.Height)
		err = bs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			_ = bs.QueryExecutor.RollbackTx()
			return []*model.Block{}, err
		}
	}

	mempoolsBackupBytes = bytes.NewBuffer([]byte{})

	for _, mempool := range mempoolsBackup {
		var (
			tx     *model.Transaction
			txType transaction.TypeAction
		)
		tx, err := transaction.ParseTransactionBytes(mempool.GetTransactionBytes(), true)
		if err != nil {
			return nil, err
		}
		txType, err = bs.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}

		err = txType.UndoApplyUnconfirmed()
		if err != nil {
			return nil, err
		}

		/*
			mempoolsBackupBytes format is
			[...{4}byteSize,{bytesSize}transactionBytes]
		*/
		sizeMempool := uint32(len(mempool.GetTransactionBytes()))
		mempoolsBackupBytes.Write(util.ConvertUint32ToBytes(sizeMempool))
		mempoolsBackupBytes.Write(mempool.GetTransactionBytes())
	}
	err = bs.QueryExecutor.CommitTx()
	if err != nil {
		return nil, err
	}

	if mempoolsBackupBytes.Len() > 0 {
		kvdbMempoolsBackupKey := commonUtils.GetKvDbMempoolDBKey(bs.GetChainType())
		err = bs.KVExecutor.Insert(kvdbMempoolsBackupKey, mempoolsBackupBytes.Bytes(), int(constant.KVDBMempoolsBackupExpiry))
		if err != nil {
			return nil, err
		}
	}
	// remove peer memoization
	bs.NodeRegistrationService.ResetScrambledNodes()
	return poppedBlocks, nil
}

// ProcessCompletedBlock to precess block that already having all needed transactions
func (bs *BlockService) ProcessCompletedBlock(block *model.Block) error {
	// pushBlock closure to release lock as soon as block pushed
	// Securing receive block process
	bs.ChainWriteLock(constant.BlockchainStatusReceivingBlock)
	defer bs.ChainWriteUnlock(constant.BlockchainStatusReceivingBlock)

	// making sure get last block after paused process
	var lastBlock, err = bs.GetLastBlock()
	if err != nil {
		return status.Error(codes.Internal,
			"fail to get last block",
		)
	}
	// when incoming block is better than last block that having same prevoius block hash
	if bytes.Equal(lastBlock.GetPreviousBlockHash(), block.PreviousBlockHash) &&
		block.Timestamp < lastBlock.Timestamp {
		var previousBlock, err = commonUtils.GetBlockByHeight(lastBlock.Height-1, bs.QueryExecutor, bs.BlockQuery)
		if err != nil {
			return status.Error(codes.Internal,
				"fail to get last block",
			)
		}
		// Pop off last block to trying push incoming block
		lastBlocks, err := bs.PopOffToBlock(previousBlock)
		if err != nil {
			return err
		}
		err = bs.ValidateBlock(block, previousBlock, time.Now().Unix())
		if err != nil {
			errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], false)
			if errPushBlock != nil {
				bs.Logger.Errorf("pushing back popped off block fail: %v", errPushBlock)
				return status.Error(codes.InvalidArgument, "InvalidBlock")
			}

			bs.Logger.Info("pushing back popped off block")
			return status.Error(codes.InvalidArgument, "InvalidBlock")
		}
		err = bs.PushBlock(previousBlock, block, true)
		if err != nil {
			errPushBlock := bs.PushBlock(previousBlock, lastBlocks[0], true)
			if errPushBlock != nil {
				bs.Logger.Errorf("pushing back popped off block fail: %v", errPushBlock)
				return status.Error(codes.InvalidArgument, "InvalidBlock")
			}
			bs.Logger.Info("pushing back popped off block")
			return status.Error(codes.InvalidArgument, "InvalidBlock")
		}

		return nil
	}

	// normal process to validate and push block
	err = bs.ValidateBlock(block, lastBlock, time.Now().Unix())
	if err != nil {
		return status.Error(codes.InvalidArgument, "InvalidBlock")
	}
	err = bs.PushBlock(lastBlock, block, true)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

// ProcessQueuedBlock process to queue block when waiting their transactions
// block will directly into completed precess block when isQueued is false and err is nil
func (bs *BlockService) ProcessQueuedBlock(block *model.Block) (isQueued bool, err error) {
	// check block having transactions or not
	if len(block.TransactionIDs) == 0 {
		return false, nil
	}
	bs.WaitingTransactionBlockQueue.BlockMutex.Lock()
	defer bs.WaitingTransactionBlockQueue.BlockMutex.Unlock()
	// check block already queued or not
	if bs.WaitingTransactionBlockQueue.WaitingTxBlocks[block.ID] != nil {
		return true, nil
	}
	var (
		txRequiredByBlock     = make(TransactionIDsMap)
		txRequiredByBlockArgs []interface{}
	)
	block.Transactions = make([]*model.Transaction, len(block.GetTransactionIDs()))
	for idx, txID := range block.TransactionIDs {
		// check if the transaction already exists in the blockTransactionsCandidate
		var txFound = bs.MempoolService.GetBlockTxCached(txID)
		if txFound != nil {
			// add transaction into block
			block.Transactions[idx] = txFound
			continue
		}
		// save transaction ID when transaction not found
		if bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[txID] == nil {
			bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[txID] = make(BlockIDsMap)
		}
		bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[txID][block.GetID()] = true
		txRequiredByBlock[txID] = idx
		// used as argument when quermockBlockDataying in mempool
		txRequiredByBlockArgs = append(txRequiredByBlockArgs, txID)
	}
	// process when needed trasacntions are completed
	if len(txRequiredByBlock) == 0 {
		err := bs.ProcessCompletedBlock(block)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// looking rest of needed transactions in mempool
	var (
		caseQuery    = query.NewCaseQuery()
		mempoolQuery = query.NewMempoolQuery(bs.Chaintype)
		mempools     []*model.MempoolTransaction
	)
	// build query to select transaction in mempool transaction
	caseQuery.Select(mempoolQuery.TableName, mempoolQuery.Fields...)
	caseQuery.Where(caseQuery.In("id", txRequiredByBlockArgs...))
	selectQuery, args := caseQuery.Build()
	rows, err := bs.QueryExecutor.ExecuteSelect(selectQuery, false, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	mempools, err = mempoolQuery.BuildModel(mempools, rows)
	if err != nil {
		return false, err
	}
	for _, mempool := range mempools {
		tx, err := transaction.ParseTransactionBytes(mempool.TransactionBytes, true)
		if err != nil {
			continue
		}
		block.Transactions[txRequiredByBlock[tx.GetID()]] = tx
		delete(bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[tx.GetID()], block.GetID())
		delete(txRequiredByBlock, tx.GetID())
	}
	// process when needed trasacntions are completed
	if len(txRequiredByBlock) == 0 {
		err := bs.ProcessCompletedBlock(block)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// saving temporary map ID transactions and block
	bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap[block.ID] = txRequiredByBlock
	bs.WaitingTransactionBlockQueue.WaitingTxBlocks[block.ID] = &BlockWithMetaData{
		Block:     block,
		Timestamp: time.Now().Unix(),
	}
	bs.RequestBlockTransactions(block.GetTransactionIDs())
	return true, nil
}

// RequestBlockTransactions request transactons of incoming block
func (bs *BlockService) RequestBlockTransactions(txIds []int64) {
	// TODO: chunks requested transaction
	bs.Observer.Notify(observer.BlockRequestTransactions, txIds, bs.Chaintype)
}

// ReceiveTransactionListener will received transaction for queued block
func (bs *BlockService) ReceiveTransactionListener(transaction *model.Transaction) {
	bs.WaitingTransactionBlockQueue.BlockMutex.Lock()
	defer bs.WaitingTransactionBlockQueue.BlockMutex.Unlock()
	for blockID := range bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[transaction.GetID()] {
		// check if waiting block don't have required this transaction
		if bs.WaitingTransactionBlockQueue.WaitingTxBlocks[blockID] == nil ||
			bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap[blockID] == nil {
			continue
		}
		var (
			txs     = bs.WaitingTransactionBlockQueue.WaitingTxBlocks[blockID].Block.GetTransactions()
			txIndex = bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap[blockID][transaction.GetID()]
		)
		// joining new transaction into existing transactions
		txs[txIndex] = transaction
		bs.WaitingTransactionBlockQueue.WaitingTxBlocks[blockID].Block.Transactions = txs
		delete(bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap[blockID], transaction.GetID())
		// process block when all transactions are completed
		if len(bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap[blockID]) == 0 {
			err := bs.ProcessCompletedBlock(bs.WaitingTransactionBlockQueue.WaitingTxBlocks[blockID].Block)
			if err != nil {
				continue
			}
			// remove waited block and list of transaction ID map when block already pushed
			delete(bs.WaitingTransactionBlockQueue.WaitingTxBlocks, blockID)
			delete(bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap, blockID)
		}
	}
	// removing transaction ID and transaction candidate when it's not needed by any block
	delete(bs.WaitingTransactionBlockQueue.TransactionsRequiredMap, transaction.GetID())
	bs.MempoolService.DeleteBlockTxCandidate(transaction.GetID(), false)
}

// CleanTheTimedoutBlock will remove waited block when block waiting too long
func (bs *BlockService) CleanTheTimedoutBlock() {
	bs.WaitingTransactionBlockQueue.BlockMutex.Lock()
	defer bs.WaitingTransactionBlockQueue.BlockMutex.Unlock()
	for blockID, blockWithMeta := range bs.WaitingTransactionBlockQueue.WaitingTxBlocks {
		// check waiting time block
		if blockWithMeta.Timestamp <= time.Now().Unix()-constant.TimeOutBlockWaitingTransactions {
			for _, transactionID := range blockWithMeta.Block.GetTransactionIDs() {
				delete(bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[transactionID], blockID)
				// removing transaction candidate when it's not needed by any block
				if len(bs.WaitingTransactionBlockQueue.TransactionsRequiredMap[transactionID]) == 0 {
					bs.MempoolService.DeleteBlockTxCandidate(transactionID, false)
				}
			}
			delete(bs.WaitingTransactionBlockQueue.WaitingTxBlocks, blockID)
			delete(bs.WaitingTransactionBlockQueue.BlockRequiringTransactionsMap, blockID)

		}
	}
}
