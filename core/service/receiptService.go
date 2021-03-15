// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package service

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"

	"sort"
	"time"

	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptServiceInterface interface {
		Initialize() error
		SelectReceipts(
			numberOfReceipt, blockHeight uint32,
			blockSeed []byte,
		) ([][]*model.PublishedReceipt, error)
		SelectLinkedReceipts(
			numberOfUnlinkedReceipts, numberOfReceipt, blockHeight uint32,
			blockSeed []byte,
		) ([]*model.PublishedReceipt, error)
		SelectUnlinkedReceipts(
			numberOfReceipt, blockHeight uint32,
			blockSeed []byte,
		) ([]*model.PublishedReceipt, error)
		ValidateUnlinkedReceipts(
			receiptsToValidate []*model.PublishedReceipt,
			blockToValidate *model.Block,
		) (validReceipts []*model.PublishedReceipt, err error)
		ValidateLinkedReceipts(
			receiptsToValidate []*model.PublishedReceipt,
			blockToValidate *model.Block,
			maxLookBackwardSteps int32,
		) (validReceipts []*model.PublishedReceipt, err error)
		FlattenSelectedReceipts(
			selectedReceiptsMatrix [][]*model.PublishedReceipt,
		) []*model.PublishedReceipt
		GenerateReceiptsMerkleRoot(lastBlock *model.Block) error
		GenerateReceiptsMerkleRootListener() observer.Listener
		// ValidateReceipt to validating *model.BatchReceipt when send block or send transaction and also when want to publishing receipt
		ValidateReceipt(
			receipt *model.Receipt,
			validateRefBlock bool,
		) error
		GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error)
		// GenerateReceiptWithReminder generating batch receipt and store to reminder also
		GenerateReceiptWithReminder(
			ct chaintype.ChainType,
			receivedDatumHash []byte,
			lastBlock *storage.BlockCacheObject,
			senderPublicKey []byte,
			nodeSecretPhrase string,
			datumType uint32,
		) (*model.Receipt, error)
		// CheckDuplication to check duplication of *model.BatchReceipt when get response from send block and send transaction
		CheckDuplication(publicKey []byte, datumHash []byte) (err error)
		StoreReceipt(receipt *model.Receipt, senderPublicKey []byte, chaintype chaintype.ChainType) error
		ClearCache()
	}

	ReceiptService struct {
		NodeReceiptQuery         query.BatchReceiptQueryInterface
		MerkleTreeQuery          query.MerkleTreeQueryInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		BlockQuery               query.BlockQueryInterface
		TransactionQuery         query.TransactionQueryInterface
		QueryExecutor            query.ExecutorInterface
		NodeRegistrationService  NodeRegistrationServiceInterface
		Signature                crypto.SignatureInterface
		PublishedReceiptQuery    query.PublishedReceiptQueryInterface
		ReceiptUtil              coreUtil.ReceiptUtilInterface
		MainBlockStateStorage    storage.CacheStorageInterface
		ScrambleNodeService      ScrambleNodeServiceInterface
		ReceiptReminderStorage   storage.CacheStorageInterface
		BatchReceiptCacheStorage storage.CacheStorageInterface
		MainBlocksStorage        storage.CacheStackStorageInterface
		NodeConfiguration        NodeConfigurationServiceInterface
		// local cache
		LastMerkleRoot []byte
		MerkleRootUtil util.MerkleRootInterface
		Logger         *log.Logger
	}
)

func (rs *ReceiptService) GenerateReceiptsMerkleRootListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args ...interface{}) {
			var (
				b         *model.Block
				chainType chaintype.ChainType
				ok        bool
			)
			b, ok = block.(*model.Block)
			if !ok {
				rs.Logger.Fatalln("Block casting failures in SendBlockListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				rs.Logger.Fatalln("chainType casting failures in SendBlockListener")
			}

			if chainType.GetTypeInt() == (&chaintype.MainChain{}).GetTypeInt() {
				if err := rs.GenerateReceiptsMerkleRoot(b); err != nil {
					rs.Logger.Error(err)
				}
			}
		},
	}
}

func NewReceiptService(
	nodeReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	blockQuery query.BlockQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	queryExecutor query.ExecutorInterface,
	nodeRegistrationService NodeRegistrationServiceInterface,
	signature crypto.SignatureInterface,
	publishedReceiptQuery query.PublishedReceiptQueryInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
	mainBlockStateStorage, receiptReminderStorage, batchReceiptCacheStorage storage.CacheStorageInterface,
	scrambleNodeService ScrambleNodeServiceInterface,
	mainBlocksStorage storage.CacheStackStorageInterface,
	merkleRootUtil util.MerkleRootInterface,
	nodeConfiguration NodeConfigurationServiceInterface,
	logger *log.Logger,
) *ReceiptService {
	return &ReceiptService{
		NodeReceiptQuery:         nodeReceiptQuery,
		MerkleTreeQuery:          merkleTreeQuery,
		NodeRegistrationQuery:    nodeRegistrationQuery,
		BlockQuery:               blockQuery,
		TransactionQuery:         transactionQuery,
		QueryExecutor:            queryExecutor,
		NodeRegistrationService:  nodeRegistrationService,
		Signature:                signature,
		PublishedReceiptQuery:    publishedReceiptQuery,
		ReceiptUtil:              receiptUtil,
		MainBlockStateStorage:    mainBlockStateStorage,
		ScrambleNodeService:      scrambleNodeService,
		ReceiptReminderStorage:   receiptReminderStorage,
		BatchReceiptCacheStorage: batchReceiptCacheStorage,
		MainBlocksStorage:        mainBlocksStorage,
		LastMerkleRoot:           nil,
		MerkleRootUtil:           merkleRootUtil,
		NodeConfiguration:        nodeConfiguration,
		Logger:                   logger,
	}
}

func (rs *ReceiptService) Initialize() error {
	lastRmrQ := rs.MerkleTreeQuery.GetLastMerkleRoot()
	row, _ := rs.QueryExecutor.ExecuteSelectRow(lastRmrQ, false)

	lastMerkleRoot, err := rs.MerkleTreeQuery.ScanRoot(row)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}
	rs.LastMerkleRoot = lastMerkleRoot
	return nil
}

// getPriorityPeersAtHeight compute the set of assigned receivers (priority peers) at a given block height
func (rs *ReceiptService) getPriorityPeersAtHeight(nodePubKey []byte, blockHeight uint32) (map[string]*model.Peer, error) {
	// the node looks up its position in scramble nodes at the look back height,
	// and computes its set of assigned receivers at that time (gets its priority peers at that height).
	var (
		scrambleNodes *model.ScrambledNodes
		err           error
	)
	scrambleNodes, err = rs.ScrambleNodeService.GetScrambleNodesByHeight(blockHeight)
	if err != nil {
		return nil, err
	}
	if scrambleNodes.NodePublicKeyToIDMap == nil {
		return nil, errors.New("NodePublicKeyToIDMapEmpty")
	}
	return rs.ReceiptUtil.GetPriorityPeersAtHeight(nodePubKey, scrambleNodes)
}

// SelectUnlinkedReceipts select receipts received from node's priority peers at a given block height (from either a transaction or block
// broadcast to them by current node
func (rs *ReceiptService) SelectUnlinkedReceipts(
	numberOfReceipt, blockHeight uint32,
	blockSeed []byte,
) ([]*model.PublishedReceipt, error) {
	var (
		err                               error
		qryStr                            string
		batchMerkleRoot                   []byte
		batchReceipts, validBatchReceipts []*model.BatchReceipt
		lookBackBlock                     *model.Block
		unlinkedBatchReceiptList          []*model.BatchReceipt
		unlinkedPublishedReceipts         []*model.PublishedReceipt
		lookBackHeight                    = blockHeight - constant.BatchReceiptLookBackHeight
		emptyReceipts                     = make([]*model.PublishedReceipt, 0)
	)

	// possible no connected node || lastblock height is too low to select receipts
	if numberOfReceipt == 0 || blockHeight < constant.BatchReceiptLookBackHeight {
		return emptyReceipts, nil
	}

	// get the reference block's and transactions to select batch receipts from
	lookBackBlock, err = util.GetBlockByHeight(
		lookBackHeight,
		rs.QueryExecutor,
		rs.BlockQuery,
	)
	if err != nil {
		return nil, err
	}
	// get block transactions and published receipts
	err = util.PopulateBlockData(lookBackBlock, rs.QueryExecutor, rs.TransactionQuery, rs.PublishedReceiptQuery)
	if err != nil {
		return nil, err
	}

	// get merkle root for this reference previousBlock (previousBlock at lookBackHeight)
	qryStr = rs.MerkleTreeQuery.SelectMerkleTreeAtHeight(lookBackHeight)
	mrRow, err := rs.QueryExecutor.ExecuteSelectRow(qryStr, false)
	if err != nil {
		return nil, err
	}
	batchMerkleRoot, err = rs.MerkleTreeQuery.ScanRoot(mrRow)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyReceipts, nil
		}
		return nil, err
	}

	// roll a random number based on lastBlock's previousBlock seed to select specific data from the looked up previousBlock
	// (lastblock height - BatchReceiptLookBackHeight)
	var (
		rndDatumType uint32
		hashList     [][]byte
		rndDatumHash []byte
	)
	hashList, err = rs.ReceiptUtil.BuildBlockDatumHashes(lookBackBlock, rs.QueryExecutor, rs.TransactionQuery)
	if err != nil {
		return nil, err
	}
	rndDatumHash, rndDatumType, err = rs.ReceiptUtil.GetRandomDatumHash(hashList, blockSeed)
	if err != nil {
		return nil, err
	}

	// get all batch receipts for selected merkle root and datum_hash ordered by recipient_public_key, reference_block_height
	batchReceipts, err = func() ([]*model.BatchReceipt, error) {
		var receipts []*model.BatchReceipt
		batchReceiptsQ, rootArgs := rs.NodeReceiptQuery.GetReceiptsByRootAndDatumHash(batchMerkleRoot, rndDatumHash, rndDatumType)
		rows, err := rs.QueryExecutor.ExecuteSelect(batchReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return rs.NodeReceiptQuery.BuildModel(receipts, rows)
	}()
	if err != nil {
		return nil, err
	}
	for _, br := range batchReceipts {
		if err = rs.ValidateReceipt(br.Receipt, true); err == nil {
			validBatchReceipts = append(validBatchReceipts, br)
		}
	}
	if len(validBatchReceipts) == 0 {
		return nil, nil
	}
	// the node looks up its position in scramble nodes at the look back height,
	// and computes its set of assigned receivers at that time (gets its priority peers at that height).
	nodePubKey := rs.NodeConfiguration.GetNodePublicKey()
	priorityPeersAtHeight, err := rs.getPriorityPeersAtHeight(nodePubKey, lookBackHeight)
	if err != nil || len(priorityPeersAtHeight) == 0 {
		if err != nil {
			rs.Logger.Error(err)
		}
		// TODO: lower this to debug once tested on alpha network to not pollute logs
		rs.Logger.Error("no priority peers for node at block height: ", lookBackHeight)
		return emptyReceipts, nil
	}

	// select all batch (node) receipts, collected at look back height, from one of the priority peers (at that height),
	// that match the random datum hash (transaction or block hash) rolled earlier
	for _, priorityPeer := range priorityPeersAtHeight {
		for _, batchReceipt := range validBatchReceipts {
			if bytes.Equal(batchReceipt.GetReceipt().RecipientPublicKey, priorityPeer.GetInfo().GetPublicKey()) {
				unlinkedBatchReceiptList = append(unlinkedBatchReceiptList, batchReceipt)
			}
		}
	}

	for idx, batchReceipt := range unlinkedBatchReceiptList {
		// remove RMRLinked from the batch receipt,
		// since this is an unlinked published receipt and the way we link published receipts is different from
		unlinkedReceiptToPublish, err := rs.ReceiptUtil.GeneratePublishedReceipt(
			batchReceipt.Receipt,
			uint32(idx),
			nil,
			0,
			rs.QueryExecutor,
			query.NewMerkleTreeQuery(),
		)
		if err != nil {
			rs.Logger.Error(err)
			continue
		}
		unlinkedPublishedReceipts = append(unlinkedPublishedReceipts, unlinkedReceiptToPublish)
	}
	return unlinkedPublishedReceipts, nil
}

// ValidateUnlinkedReceipts validate receipt published in a block downloaded from a peer (part of block validation process)
func (rs *ReceiptService) ValidateUnlinkedReceipts(
	receiptsToValidate []*model.PublishedReceipt,
	blockToValidate *model.Block,
) (validReceipts []*model.PublishedReceipt, err error) {
	var (
		lookBackBlock         *model.Block
		blockToValidateHeight = blockToValidate.Height
		lookBackHeight        = blockToValidateHeight - constant.BatchReceiptLookBackHeight
	)

	// possible no connected node || lastblock height is too low to select receipts
	if blockToValidateHeight < constant.BatchReceiptLookBackHeight {
		return nil, nil
	}

	// get the reference block's and transactions to select batch receipts from
	lookBackBlock, err = util.GetBlockByHeight(
		lookBackHeight,
		rs.QueryExecutor,
		rs.BlockQuery,
	)
	if err != nil {
		return validReceipts, err
	}
	// get block transactions and published receipts
	err = util.PopulateBlockData(lookBackBlock, rs.QueryExecutor, rs.TransactionQuery, rs.PublishedReceiptQuery)
	if err != nil {
		return nil, err
	}

	var (
		rndDatumType uint32
		hashList     [][]byte
		rndDatumHash []byte
	) // list of hashes of all data secured by the block
	hashList, err = rs.ReceiptUtil.BuildBlockDatumHashes(lookBackBlock, rs.QueryExecutor, rs.TransactionQuery)
	if err != nil {
		return validReceipts, err
	}

	// roll a random number based on new block seed to select specific data from the looked up
	rndDatumHash, rndDatumType, err = rs.ReceiptUtil.GetRandomDatumHash(hashList, blockToValidate.BlockSeed)
	if err != nil {
		return validReceipts, err
	}

	// the node looks up its position in scramble nodes at the look back height,
	// and computes its set of assigned receivers at that time (gets its priority peers at that height).
	priorityPeersAtHeight, err := rs.getPriorityPeersAtHeight(blockToValidate.BlocksmithPublicKey, lookBackHeight)
	if err != nil {
		return validReceipts, err
	}

	// look through each unlinked receipt published in the new block.
	for _, publishedReceipt := range receiptsToValidate {
		// validate that each one has the correct "data hash" we rolled,
		if !bytes.Equal(publishedReceipt.Receipt.DatumHash, rndDatumHash) ||
			publishedReceipt.Receipt.DatumType != rndDatumType {
			return nil, errors.New("InvalidReceiptDatumHash")
		}
		// verify that each receipt is from one of the 5 assigned receivers to that block creator at that time (height)
		// TODO: verify that is the recipient and not the sender
		if _, ok := priorityPeersAtHeight[hex.EncodeToString(publishedReceipt.Receipt.RecipientPublicKey)]; !ok {
			return nil, errors.New("InvalidReceiptReceiver")
		}
		validReceipts = append(validReceipts, publishedReceipt)
	}
	if len(validReceipts) != len(receiptsToValidate) {
		return nil, errors.New("LinkedReceiptFinalValidationError")
	}
	return validReceipts, nil
}

// SelectLinkedReceipts select receipts received from node's priority peers at a given block height (from either a transaction or block
// broadcast to them by current node
func (rs *ReceiptService) SelectLinkedReceipts(
	numberOfUnlinkedReceipts, numberOfReceipt, blockHeight uint32,
	blockSeed []byte,
) ([]*model.PublishedReceipt, error) {

	var (
		linkedReceipts []*model.PublishedReceipt
		referenceBlock *model.Block
		nodePublicKey  = rs.NodeConfiguration.GetNodePublicKey()
		// maxLookBackwardSteps max n. of times this node should look backwards trying to link receipts when he was one of the scramble nodes
		// note: numberOfReceipts = number of max priority peers the node has
		maxLookBackwardSteps           = numberOfReceipt
		rmrLinkedIndex, publishedIndex uint32
	)

	// possible no connected node || lastblock height is too low to select receipts
	if numberOfReceipt == 0 || blockHeight < 2*constant.BatchReceiptLookBackHeight {
		return nil, nil
	}

	// set first index for linked published receipts one unit more than the latest unlinked receipt index
	if numberOfUnlinkedReceipts > 0 {
		publishedIndex = rmrLinkedIndex + numberOfUnlinkedReceipts + 1
	}
	// loop backwards searching for blocks where current node was one of the block creators (when was in scramble node list)
	// TODO: add a limit to the number of max blocks that can be scanned
	for lookBackHeightInt := int32(blockHeight - 1); lookBackHeightInt >= 0; lookBackHeightInt-- {
		if lookBackHeightInt < constant.BatchReceiptLookBackHeight {
			break
		}
		var (
			refPublishedReceipt     = query.NewPublishedReceipt()
			lookBackHeight          = uint32(lookBackHeightInt)
			unlinkedReceiptLookback = lookBackHeight - constant.BatchReceiptLookBackHeight
		)

		if maxLookBackwardSteps == 0 {
			break
		}

		// get the block creator's pub key
		lookBackBlock, err := util.GetBlockByHeight(
			lookBackHeight,
			rs.QueryExecutor,
			rs.BlockQuery,
		)
		if err != nil {
			return nil, err
		}
		priorityPeersAtHeight, err := rs.getPriorityPeersAtHeight(lookBackBlock.BlocksmithPublicKey, unlinkedReceiptLookback)
		if err != nil || len(priorityPeersAtHeight) == 0 {
			if err != nil {
				rs.Logger.Error(err)
			} else {
				rs.Logger.Debug("no priority peers for node at block height: ", lookBackHeight)
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 10)
			rs.Logger.Error("============== 10")
			continue
		}
		if _, ok := priorityPeersAtHeight[hex.EncodeToString(nodePublicKey)]; !ok {
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 20)
			rs.Logger.Error("============== 20")
			continue
		}

		// we found one block where the node was one of the scramble nodes
		maxLookBackwardSteps--
		monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 30)
		rs.Logger.Error("============== 30")

		// get the unlinked published receipt that has current node as recipient, at reference block height
		batchReceiptsQ, rootArgs := rs.PublishedReceiptQuery.GetUnlinkedPublishedReceiptByBlockHeightAndReceiver(lookBackHeight, nodePublicKey)
		row, err := rs.QueryExecutor.ExecuteSelectRow(batchReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		err = rs.PublishedReceiptQuery.Scan(refPublishedReceipt, row)
		if err != nil {
			if err == sql.ErrNoRows {
				// there are no published receipts to link a batch to, fail this block and continue
				monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 40)
				rs.Logger.Error("============== 40")
				continue
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 50)
			rs.Logger.Error("============== 50")
			return nil, err
		}

		// get the block to select relative transaction hashes and previous block hash (similar to what we do for unlinked receipts)
		referenceBlock, err = util.GetBlockByHeight(
			refPublishedReceipt.GetReceipt().ReferenceBlockHeight,
			rs.QueryExecutor,
			rs.BlockQuery,
		)
		// this should not happen, since we already verify that this block exists when validating the receipt
		if err != nil {
			return nil, err
		}
		// get block transactions and published receipts
		err = util.PopulateBlockData(referenceBlock, rs.QueryExecutor, rs.TransactionQuery, rs.PublishedReceiptQuery)
		if err != nil {
			return nil, err
		}

		var (
			rndDatumType uint32
			hashList     [][]byte
			rndDatumHash []byte
		) // list of hashes of all data secured by the block
		hashList, err = rs.ReceiptUtil.BuildBlockDatumHashes(referenceBlock, rs.QueryExecutor, rs.TransactionQuery)
		if err != nil {
			return nil, err
		}

		// roll a random number based on new block seed to select specific data from the looked up
		rndDatumHash, rndDatumType, err = rs.ReceiptUtil.GetRandomDatumHash(hashList, blockSeed)
		if err != nil {
			return nil, err
		}

		var (
			priorityPeersArray []*model.Peer
		)
		// the node looks up its position in scramble nodes at the referenceBlock height,
		// and computes its set of assigned receivers at that time (gets its priority peers at that height).
		nodePubKey := rs.NodeConfiguration.GetNodePublicKey()
		priorityPeersAtHeight, err = rs.getPriorityPeersAtHeight(nodePubKey, referenceBlock.Height)
		if err != nil || len(priorityPeersAtHeight) == 0 {
			if err != nil {
				rs.Logger.Error(err)
			} else {
				rs.Logger.Debug("no priority peers for node at block height: ", referenceBlock.Height)
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 60)
			rs.Logger.Error("============== 60")
			continue
		}
		for _, pp := range priorityPeersAtHeight {
			priorityPeersArray = append(priorityPeersArray, pp)
		}
		priorityPeersArray = sortPeers(priorityPeersArray)

		// Use the new block seed (from the block I am creating) to "roll" a random number to select one of my N assigned receivers.
		// This determines the “receiver” I must produce a receipt for.
		var (
			rng        = crypto.NewRandomNumberGenerator()
			rndPeerIdx int
		)
		err = rng.Reset(constant.ReceiptSelectionSeedPrefix, blockSeed)
		if err != nil {
			return nil, err
		}
		rndSeedInt := rng.Next()
		rndSelSeed := rand.NewSource(rndSeedInt)
		rnd := rand.New(rndSelSeed)
		if len(priorityPeersAtHeight) > 1 {
			rndPeerIdx = rnd.Intn(len(priorityPeersAtHeight) - 1)
		}
		rndPeer := priorityPeersArray[rndPeerIdx]

		if rndPeer.GetInfo() == nil {
			return nil, errors.New("priorityPeerInfoEmpty")
		}

		var (
			batchReceiptToLink = query.NewBatchReceipt()
		)
		// 2. Query my receipts in the batch, to find one that matches the data hash we rolled,
		// and also matches the receiver that we rolled. If we *do not* have a receipt in the batch that matches these 2,
		// then we fail this block (cannot link a receipt to it), and continue iterating backwards.
		batchReceiptsQ, rootArgs = rs.NodeReceiptQuery.GetReceiptByRecipientAndDatumHash(
			rndDatumHash,
			rndDatumType,
			rndPeer.GetInfo().PublicKey,
		)
		row, err = rs.QueryExecutor.ExecuteSelectRow(batchReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		err = rs.NodeReceiptQuery.Scan(batchReceiptToLink, row)
		if err != nil {
			if err != sql.ErrNoRows {
				rs.Logger.Error(err)
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 70)
			rs.Logger.Error("============== 70")
			continue
		}
		if err = rs.ValidateReceipt(batchReceiptToLink.Receipt, true); err != nil {
			// batch receipt is invalid. fail the block and continue
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 80)
			rs.Logger.Error("============== 80")
			continue
		}

		// 3. If we DO have a matching receipt (which we should, if we’ve been doing all of our work on the network),
		// this becomes one of our N "linked receipts", which we can publish in our new block.
		publishedReceipt, err := rs.ReceiptUtil.GeneratePublishedReceipt(
			batchReceiptToLink.Receipt,
			publishedIndex,
			refPublishedReceipt.Receipt.RMR,
			batchReceiptToLink.RMRBatchIndex,
			rs.QueryExecutor,
			query.NewMerkleTreeQuery(),
		)
		publishedIndex++
		rmrLinkedIndex++
		if err != nil {
			rs.Logger.Error(err)
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 90)
			rs.Logger.Error("============== 90")
			continue
		}
		monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 100)
		rs.Logger.Error("============== 100")
		linkedReceipts = append(linkedReceipts, publishedReceipt)
	}
	monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 110)
	rs.Logger.Error("============== 110")
	return linkedReceipts, nil
}

func (rs *ReceiptService) ValidateLinkedReceipts(
	receiptsToValidate []*model.PublishedReceipt,
	blockToValidate *model.Block,
	maxLookBackwardSteps int32,
) (validReceipts []*model.PublishedReceipt, err error) {
	var (
		linkedReceipts []*model.PublishedReceipt
	)

	if receiptsToValidate == nil || len(receiptsToValidate) == 0 {
		return receiptsToValidate, nil
	}

	// loop backwards searching for blocks where current block creators was in priority peers of another block creator
	for lookBackHeightInt := int32(blockToValidate.Height - 1); lookBackHeightInt >= 0; lookBackHeightInt-- {
		if lookBackHeightInt < constant.BatchReceiptLookBackHeight {
			break
		}
		var (
			lookBackHeight          = uint32(lookBackHeightInt)
			unlinkedReceiptLookback = lookBackHeight - constant.BatchReceiptLookBackHeight
			lookBackBlockReceipt    *model.PublishedReceipt
			lookBackBlock, err      = util.GetBlockByHeight(lookBackHeight, rs.QueryExecutor, rs.BlockQuery)
			receiptToValidate       *model.PublishedReceipt
		)
		// this should not happen, since we already verify that this block exists when validating the receipt
		if err != nil {
			return nil, err
		}

		if maxLookBackwardSteps == 0 {
			break
		}

		// check if the new block's creator is one of Priority Peers at look back height and if not skip this block and continue
		// if not returns and empty peer array
		priorityPeersAtHeight, err := rs.getPriorityPeersAtHeight(lookBackBlock.BlocksmithPublicKey, unlinkedReceiptLookback)
		if err != nil || len(priorityPeersAtHeight) == 0 {
			if err != nil {
				rs.Logger.Error(err)
			} else {
				rs.Logger.Debug("no priority peers for node at block height: ", lookBackHeight)
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 200)
			rs.Logger.Error("============== 200")
			return nil, errors.New("CannotValidatePriorityPeersAtHeight")
		}

		if _, ok := priorityPeersAtHeight[hex.EncodeToString(blockToValidate.BlocksmithPublicKey)]; !ok {
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 200)
			rs.Logger.Error("============== 210")
			continue
		}

		// get block transactions and published receipts
		err = util.PopulateBlockData(lookBackBlock, rs.QueryExecutor, rs.TransactionQuery, rs.PublishedReceiptQuery)
		if err != nil {
			return nil, err
		}

		// we found one block where the new block's creator was one of the priority peers to that block's creator :)
		// if we fail validation from now on, we miss one of the receipts that are supposed to be linked to the new block
		maxLookBackwardSteps--

		// loop through the unlinked published receipt from ref block and see if any of them have receipt.receiver = new_block.creator
		var (
			receiptFound = false
		)
		for _, lookBackBlockReceipt = range lookBackBlock.GetPublishedReceipts() {
			// did the new block creator have an unlinked receipt in ref block?
			if lookBackBlockReceipt.RMRLinked == nil {
				if bytes.Equal(lookBackBlockReceipt.RMRLinked, []byte{}) && bytes.Equal(lookBackBlockReceipt.Receipt.RecipientPublicKey, blockToValidate.BlocksmithPublicKey) {
					receiptFound = true
					break
				}
			}
		}
		// new block's creator didn't produce an unlinked receipt relative to the linked receipt we are validating,
		// when it was supposed to do it, so we skip this block
		if !receiptFound {
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 220)
			rs.Logger.Error("============== 220")
			continue
		}

		// Lookup the reference block
		var (
			referenceBlockHeight = lookBackBlockReceipt.Receipt.ReferenceBlockHeight
		)
		// get the block to select relative transaction hashes and previous block hash (similar to what we do for unlinked receipts)
		referenceBlock, err := util.GetBlockByHeight(
			referenceBlockHeight,
			rs.QueryExecutor,
			rs.BlockQuery,
		)
		// this should not happen, since we already verify that this block exists when validating the receipt
		if err != nil {
			return nil, err
		}
		// get block transactions and published receipts
		err = util.PopulateBlockData(referenceBlock, rs.QueryExecutor, rs.TransactionQuery, rs.PublishedReceiptQuery)
		if err != nil {
			return nil, err
		}

		var (
			rndDatumType uint32
			hashList     [][]byte
			rndDatumHash []byte
		) // list of hashes of all data secured by the block
		hashList, err = rs.ReceiptUtil.BuildBlockDatumHashes(referenceBlock, rs.QueryExecutor, rs.TransactionQuery)
		if err != nil {
			return nil, err
		}

		// roll a random number based on new block seed to select specific data from the looked up
		rndDatumHash, rndDatumType, err = rs.ReceiptUtil.GetRandomDatumHash(hashList, blockToValidate.BlocksmithPublicKey)
		if err != nil {
			return nil, err
		}

		var (
			priorityPeersArray []*model.Peer
		)
		// the node looks up its position in scramble nodes at the referenceBlock height,
		// and computes its set of assigned receivers at that time (gets its priority peers at that height).
		priorityPeersAtHeight, err = rs.getPriorityPeersAtHeight(blockToValidate.BlocksmithPublicKey, referenceBlock.Height)
		if err != nil || len(priorityPeersAtHeight) == 0 {
			if err != nil {
				rs.Logger.Error(err)
			} else {
				rs.Logger.Debug("no priority peers for node at block height: ", referenceBlock.Height)
			}
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 230)
			rs.Logger.Error("============== 230")
			return nil, errors.New("CannotValidatePriorityPeersAtHeight")
		}
		for _, pp := range priorityPeersAtHeight {
			priorityPeersArray = append(priorityPeersArray, pp)
		}
		priorityPeersArray = sortPeers(priorityPeersArray)

		// Use the new block seed (from the block I am creating) to "roll" a random number to select one of my N assigned receivers.
		// This determines the “receiver” I must produce a receipt for.
		var (
			rng        = crypto.NewRandomNumberGenerator()
			rndPeerIdx int
		)
		err = rng.Reset(constant.ReceiptSelectionSeedPrefix, blockToValidate.BlockSeed)
		if err != nil {
			return nil, err
		}
		rndSeedInt := rng.Next()
		rndSelSeed := rand.NewSource(rndSeedInt)
		rnd := rand.New(rndSelSeed)
		if len(priorityPeersAtHeight) > 1 {
			rndPeerIdx = rnd.Intn(len(priorityPeersAtHeight) - 1)
		}
		rndPeer := priorityPeersArray[rndPeerIdx]

		if rndPeer.GetInfo() == nil {
			return nil, errors.New("priorityPeerInfoEmpty")
		}

		// find the relative linked receipt (same datum hash) that we have to validate
		receiptFound = false
		for _, receiptToValidate = range receiptsToValidate {
			if bytes.Equal(receiptToValidate.RMRLinked, lookBackBlockReceipt.Receipt.RMR) {
				receiptFound = true
				break
			}
		}
		// TODO: make sure this doesn't create problems with forks. in case yes, we probably should fail the block
		if !receiptFound {
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 240)
			rs.Logger.Error("============== 240")
			continue
		}

		// intermediate hashes validation
		var (
			merkle        = util.NewMerkleRoot()
			hashedReceipt = sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(receiptToValidate.Receipt))
		)
		intermediateHashes, err := merkle.RestoreIntermediateHashes(receiptToValidate.IntermediateHashes)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		RMRToVerify, err := merkle.GetMerkleRootFromIntermediateHashes(hashedReceipt[:], receiptToValidate.RMRLinkedIndex, intermediateHashes)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		if !bytes.Equal(RMRToVerify, lookBackBlockReceipt.Receipt.RMR) {
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 250)
			rs.Logger.Error("============== 250")
			return nil, errors.New("InvalidLinkedRMR")
		}

		// does the newly published receipt match this data hash, and match this receiver?
		if !(bytes.Equal(receiptToValidate.Receipt.DatumHash, rndDatumHash) &&
			receiptToValidate.Receipt.DatumType == rndDatumType &&
			bytes.Equal(rndPeer.Info.PublicKey, receiptToValidate.Receipt.RecipientPublicKey)) {
			// then new block's linked receipt is valid!
			monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 260)
			rs.Logger.Error("============== 260")
			return nil, errors.New("LinkedReceiptsValidationError")
		}
		monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 270)
		rs.Logger.Error("============== 270")
		linkedReceipts = append(linkedReceipts, receiptToValidate)
	}
	if len(linkedReceipts) != len(receiptsToValidate) {
		monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 270)
		rs.Logger.Error("============== 280")
		return nil, errors.New("LinkedReceiptFinalValidationError")
	}
	monitoring.IncrementMainchainDownloadCycleDebugger(&chaintype.MainChain{}, 270)
	rs.Logger.Error("============== 290")
	return linkedReceipts, nil
}

// SelectReceipts select list of (linked and unlinked) receipts to be included in a block
func (rs *ReceiptService) SelectReceipts(
	numberOfReceipt, blockHeight uint32,
	blockSeed []byte,
) ([][]*model.PublishedReceipt, error) {
	var (
		err                              error
		unlinkedReceipts, linkedReceipts []*model.PublishedReceipt
		selectedReceipts                 = make([][]*model.PublishedReceipt, 2)
	)

	// select unlinked receipts
	unlinkedReceipts, err = rs.SelectUnlinkedReceipts(
		numberOfReceipt,
		blockHeight,
		blockSeed,
	)
	if err != nil {
		rs.Logger.Error(err)
	}
	selectedReceipts[0] = unlinkedReceipts

	// select linked receipts
	linkedReceipts, err = rs.SelectLinkedReceipts(
		uint32(len(unlinkedReceipts)),
		numberOfReceipt,
		blockHeight,
		blockSeed,
	)
	if err != nil {
		rs.Logger.Error(err)
	}
	selectedReceipts[1] = linkedReceipts

	return selectedReceipts, nil
}

// FlattenSelectedReceipts make a flat array of unlinked and linked receipts
func (rs *ReceiptService) FlattenSelectedReceipts(
	selectedReceiptsMatrix [][]*model.PublishedReceipt,
) []*model.PublishedReceipt {
	var selectedReceipts = make([]*model.PublishedReceipt, 0)
	for _, selectedReceiptArray := range selectedReceiptsMatrix {
		if len(selectedReceiptArray) > 0 {
			selectedReceipts = append(selectedReceipts, selectedReceiptArray...)
		}
	}
	return selectedReceipts
}

// GenerateReceiptsMerkleRoot generate merkle root of some batch receipts and also remove from cache
// generating will do when number of collected receipts(batch receipts) already <= the number of required
func (rs *ReceiptService) GenerateReceiptsMerkleRoot(block *model.Block) error {
	var (
		hashedReceipts              []*bytes.Buffer
		queries                     [][]interface{}
		batchReceipt                *model.BatchReceipt
		err                         error
		rootMerkle, treeMerkle      []byte
		blockAtHeight               *storage.BlockCacheObject
		isDbTransactionHighPriority = false
		receiptsCached              = make(map[string][]model.Receipt)
		receiptsToProcess           = make([]model.Receipt, 0)
		receiptsToSave              = make([]model.Receipt, 0)
	)

	blockAtHeight, err = util.GetBlockByHeightUseBlocksCache(
		block.Height,
		rs.QueryExecutor,
		rs.BlockQuery,
		rs.MainBlocksStorage,
	)
	if err != nil {
		return err
	}
	// Since this function runs with a delay after block has been pushed, double check that the block is still in db (eg.
	// if was in a fork it could have been popped off)
	if blockAtHeight.ID != block.ID {
		return errors.New("BlockNotInDb")
	}

	var datumHashes []string
	for _, tx := range block.GetTransactions() {
		hashString := hex.EncodeToString(tx.GetTransactionHash())
		if hashString != "" {
			datumHashes = append(datumHashes, hashString)
		}
	}
	blockHash := hex.EncodeToString(block.GetPreviousBlockHash())
	if blockHash != "" {
		datumHashes = append(datumHashes, blockHash)
	}

	if err := rs.BatchReceiptCacheStorage.GetItems(datumHashes, receiptsCached); err != nil {
		return err
	}

	receiptsToProcess = FlattenReceiptGroups(receiptsCached)

	// If no receipts in cache no need to return errors. just log a message
	if len(receiptsToProcess) == 0 {
		rs.Logger.Info("No Receipts for block height: ", block.Height)
		return nil
	}
	receiptsToProcess = SortReceipts(receiptsToProcess)

	err = rs.QueryExecutor.BeginTx(isDbTransactionHighPriority, monitoring.GenerateReceiptsMerkleRootOwnerProcess)
	if err != nil {
		return err
	}
	err = func() error {
		for _, receiptToProcess := range receiptsToProcess {
			b := receiptToProcess
			err = rs.ValidateReceipt(&b, false)
			if err == nil {
				receiptsToSave = append(receiptsToSave, b)
				hashedReceipt := sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(&b))
				hashedReceipts = append(hashedReceipts, bytes.NewBuffer(hashedReceipt[:]))
			}
		}
		_, err = rs.MerkleRootUtil.GenerateMerkleRoot(hashedReceipts)
		if err != nil {
			return err
		}
		rootMerkle, treeMerkle = rs.MerkleRootUtil.ToBytes()

		queries = make([][]interface{}, len(hashedReceipts)+1)
		for k, receipt := range receiptsToSave {
			b := receipt
			batchReceipt = &model.BatchReceipt{
				Receipt:       &b,
				RMRBatch:      rootMerkle,
				RMRBatchIndex: uint32(k),
			}
			insertNodeReceiptQ, insertNodeReceiptArgs := rs.NodeReceiptQuery.InsertReceipt(batchReceipt)
			queries[k] = append([]interface{}{insertNodeReceiptQ}, insertNodeReceiptArgs...)
		}
		insertMerkleTreeQ, insertMerkleTreeArgs := rs.MerkleTreeQuery.InsertMerkleTree(
			rootMerkle,
			treeMerkle,
			time.Now().Unix(),
			block.GetHeight(),
		)
		queries[len(queries)-1] = append([]interface{}{insertMerkleTreeQ}, insertMerkleTreeArgs...)

		err = rs.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		if rollbackErr := rs.QueryExecutor.RollbackTx(isDbTransactionHighPriority); rollbackErr != nil {
			return blocker.NewBlocker(blocker.DBErr, fmt.Sprintf("err: %v - rollbackErr: %v", err, rollbackErr))
		}
		return err
	}

	err = rs.QueryExecutor.CommitTx(isDbTransactionHighPriority)
	if err != nil {
		return err
	}
	// update local cache
	rs.LastMerkleRoot = rootMerkle
	// overwrite receipt cache with remaining receipts to be processed
	return rs.BatchReceiptCacheStorage.RemoveItems(datumHashes)
}

// CheckDuplication check existing batch receipt in cache storage
func (rs *ReceiptService) CheckDuplication(publicKey, datumHash []byte) (err error) {
	var (
		receiptKey []byte
		cType      chaintype.ChainType
	)
	if len(publicKey) == 0 && len(datumHash) == 0 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"EmptyParams",
		)
	}
	receiptKey, err = rs.ReceiptUtil.GetReceiptKey(datumHash, publicKey)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			err.Error(),
		)
	}

	err = rs.ReceiptReminderStorage.GetItem(hex.EncodeToString(receiptKey), &cType)
	if err != nil {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailedGetReceiptKey",
		)
	}
	if cType != nil {
		return blocker.NewBlocker(blocker.DuplicateReceiptErr, "ReceiptExistsOnReminder")
	}
	return nil
}

func (rs *ReceiptService) ValidateReceipt(
	receipt *model.Receipt,
	validateRefBlock bool,
) error {
	// get or build scramble nodes at height
	scrambleNode, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(receipt.ReferenceBlockHeight)
	if err != nil {
		return err
	}
	return rs.ReceiptUtil.ValidateReceiptHelper(
		receipt,
		validateRefBlock,
		rs.QueryExecutor,
		rs.BlockQuery,
		rs.MainBlocksStorage,
		rs.Signature,
		scrambleNode,
	)
}

// GetPublishedReceiptsByHeight that handling database connection to get published receipts by height
func (rs *ReceiptService) GetPublishedReceiptsByHeight(blockHeight uint32) ([]*model.PublishedReceipt, error) {
	var (
		publishedReceipts []*model.PublishedReceipt
		rows              *sql.Rows
		err               error
	)

	qStr, qArgs := rs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeight(blockHeight)
	rows, err = rs.QueryExecutor.ExecuteSelect(qStr, false, qArgs...)
	if err != nil {
		return publishedReceipts, err
	}
	defer rows.Close()

	publishedReceipts, err = rs.PublishedReceiptQuery.BuildModel(publishedReceipts, rows)
	if err != nil {
		return publishedReceipts, err
	}
	return publishedReceipts, nil
}

// GenerateReceiptWithReminder generate batch receipt at last block and store into batch receipt storage
func (rs *ReceiptService) GenerateReceiptWithReminder(
	ct chaintype.ChainType,
	receivedDatumHash []byte,
	lastBlock *storage.BlockCacheObject,
	senderPublicKey []byte,
	nodeSecretPhrase string,
	datumType uint32,
) (*model.Receipt, error) {
	var (
		rmrLinked     = rs.LastMerkleRoot
		receipt       *model.Receipt
		err           error
		nodePublicKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nodeSecretPhrase)
	)

	// generate receipt
	receipt, err = rs.ReceiptUtil.GenerateReceipt(
		ct,
		lastBlock,
		senderPublicKey,
		nodePublicKey,
		receivedDatumHash,
		rmrLinked,
		datumType,
	)
	if err != nil {
		return nil, err
	}
	receipt.RecipientSignature = rs.Signature.SignByNode(
		rs.ReceiptUtil.GetUnsignedReceiptBytes(receipt),
		nodeSecretPhrase,
	)

	receiptKey, err := rs.ReceiptUtil.GetReceiptKey(receipt.GetDatumHash(), senderPublicKey)
	if err != nil {
		return receipt, err
	}
	err = rs.ReceiptReminderStorage.SetItem(hex.EncodeToString(receiptKey), ct)
	if err != nil {
		return receipt, err
	}
	return receipt, err
}

func (rs *ReceiptService) StoreReceipt(receipt *model.Receipt, senderPublicKey []byte, chaintype chaintype.ChainType) (err error) {
	b := *receipt
	err = rs.BatchReceiptCacheStorage.SetItem(hex.EncodeToString(receipt.DatumHash), b)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReceiptService) ClearCache() {
	_ = rs.BatchReceiptCacheStorage.ClearCache()
	_ = rs.ReceiptReminderStorage.ClearCache()
}

func FlattenReceiptGroups(receiptGroups map[string][]model.Receipt) []model.Receipt {
	var receipts []model.Receipt
	for _, receiptGroup := range receiptGroups {
		receipts = append(receipts, receiptGroup...)
	}
	return receipts
}

func SortReceipts(receipts []model.Receipt) []model.Receipt {
	sort.SliceStable(receipts, func(i, j int) bool {
		// sort by signature bytes
		return bytes.Compare(receipts[i].GetRecipientSignature(), receipts[j].GetRecipientSignature()) < 0
	})
	return receipts
}

func sortPeers(receipts []*model.Peer) []*model.Peer {
	sort.SliceStable(receipts, func(i, j int) bool {
		// sort by signature bytes
		return bytes.Compare(receipts[i].Info.GetPublicKey(), receipts[j].Info.GetPublicKey()) < 0
	})
	return receipts
}
