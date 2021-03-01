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
			previousBlockHash, blockSeed []byte,
			secretPhrase string,
		) ([]*model.PublishedReceipt, error)
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

// SelectReceipts select list of receipts to be included in a block by prioritizing receipts that might
// increase the participation score of the node
// func (rs *ReceiptService) SelectReceiptsOld(
// 	blockTimestamp int64,
// 	numberOfReceipt, lastBlockHeight uint32,
// ) ([]*model.PublishedReceipt, error) {
// 	var (
// 		linkedReceiptList = make(map[string][]*model.BatchReceipt)
// 		// this variable is to store picked receipt recipient to avoid duplicates
// 		pickedRecipients  = make(map[string]bool)
// 		lowerBlockHeight  uint32
// 		linkedReceiptTree = make(map[string][]byte)
// 		err               error
// 	)
//
// 	if numberOfReceipt < 1 { // possible no connected node
// 		return []*model.PublishedReceipt{}, nil
// 	}
// 	// get the last merkle tree we have build so far
// 	if lastBlockHeight > constant.MinRollbackBlocks {
// 		lowerBlockHeight = lastBlockHeight - constant.MinRollbackBlocks
// 	}
//
// 	linkedReceiptTree, err = func() (map[string][]byte, error) {
// 		treeQ := rs.MerkleTreeQuery.SelectMerkleTreeForPublishedReceipts(
// 			lowerBlockHeight,
// 			lastBlockHeight,
// 			numberOfReceipt*constant.ReceiptBatchPickMultiplier)
// 		linkedTreeRows, err := rs.QueryExecutor.ExecuteSelect(treeQ, false)
// 		if err != nil {
// 			return linkedReceiptTree, err
// 		}
// 		defer linkedTreeRows.Close()
//
// 		return rs.MerkleTreeQuery.BuildTree(linkedTreeRows)
// 	}()
// 	if err != nil {
// 		return nil, err
// 	}
// 	for linkedRoot := range linkedReceiptTree {
// 		var nodeReceipts []*model.BatchReceipt
//
// 		nodeReceipts, err = func() ([]*model.BatchReceipt, error) {
// 			nodeReceiptsQ, rootArgs := rs.NodeReceiptQuery.GetReceiptsByRootInRange(lowerBlockHeight, lastBlockHeight, []byte(linkedRoot))
// 			rows, err := rs.QueryExecutor.ExecuteSelect(nodeReceiptsQ, false, rootArgs...)
// 			if err != nil {
// 				return nil, err
// 			}
// 			defer rows.Close()
// 			return rs.NodeReceiptQuery.BuildModel(nodeReceipts, rows)
// 		}()
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, rc := range nodeReceipts {
// 			if !pickedRecipients[string(rc.GetReceipt().GetRecipientPublicKey())] {
// 				pickedRecipients[string(rc.GetReceipt().GetRecipientPublicKey())] = true
// 				linkedReceiptList[linkedRoot] = append(linkedReceiptList[linkedRoot], rc)
// 			}
// 		}
// 	}
// 	// limit the selected portion to `numberOfReceipt` receipts
// 	// filter the selected receipts on second phase
// 	var (
// 		results []*model.PublishedReceipt
// 	)
// 	for rcRoot, rcReceipt := range linkedReceiptList {
// 		merkle := util.MerkleRoot{}
// 		merkle.HashTree = merkle.FromBytes(linkedReceiptTree[rcRoot], []byte(rcRoot))
// 		for _, rc := range rcReceipt {
// 			if len(results) >= int(numberOfReceipt) {
// 				break
// 			}
// 			err = rs.ValidateReceipt(rc.Receipt)
// 			if err != nil {
// 				// skip invalid receipt
// 				continue
// 			}
// 			var intermediateHashes [][]byte
// 			rcByte := rs.ReceiptUtil.GetSignedReceiptBytes(rc.Receipt)
// 			rcHash := sha3.Sum256(rcByte)
//
// 			intermediateHashesBuffer := merkle.GetIntermediateHashes(
// 				bytes.NewBuffer(rcHash[:]),
// 				int32(rc.RMRIndex),
// 			)
// 			for _, buf := range intermediateHashesBuffer {
// 				intermediateHashes = append(intermediateHashes, buf.Bytes())
// 			}
// 			results = append(
// 				results,
// 				&model.PublishedReceipt{
// 					Receipt:            rc.GetReceipt(),
// 					IntermediateHashes: merkle.FlattenIntermediateHashes(intermediateHashes),
// 					ReceiptIndex:       rc.RMRIndex,
// 				},
// 			)
// 		}
// 	}
// 	// select non-linked receipt
// 	if len(results) < int(numberOfReceipt) {
// 		rmrLinkedReceipts, err := rs.pickReceipts(
// 			numberOfReceipt, results, pickedRecipients, lowerBlockHeight, lastBlockHeight)
// 		if err != nil {
// 			return nil, err
// 		}
// 		results = rmrLinkedReceipts
// 	}
//
// 	return results, nil
// }

// SelectUnlinkedReceipts select receipts received from node's priority peers at a given block height (from either a transaction or block
// broadcast to them by current node
func (rs *ReceiptService) SelectUnlinkedReceipts(
	numberOfReceipt, blockHeight uint32,
	previousBlockHash, blockSeed []byte,
	secretPhrase string,
) ([]*model.PublishedReceipt, error) {
	var (
		err                               error
		qryStr                            string
		batchMerkleRoot                   []byte
		batchReceipts, validBatchReceipts []*model.BatchReceipt
		lookBackBlock                     *storage.BlockCacheObject
		lookBackBlockTransactions         []*model.Transaction
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
	lookBackBlock, err = util.GetBlockByHeightUseBlocksCache(
		lookBackHeight,
		rs.QueryExecutor,
		rs.BlockQuery,
		rs.MainBlocksStorage,
	)
	if err != nil {
		return nil, err
	}
	lookBackBlockTransactions, err = func() ([]*model.Transaction, error) {
		var transactions []*model.Transaction
		qry, args := rs.TransactionQuery.GetTransactionsByBlockID(lookBackBlock.ID)
		rows, err := rs.QueryExecutor.ExecuteSelect(qry, false, args)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return rs.TransactionQuery.BuildModel(transactions, rows)
	}()

	// get merkle root for this reference previousBlock (previousBlock at lookBackHeight)
	qryStr = rs.MerkleTreeQuery.SelectMerkleTreeAtHeight(lookBackHeight)
	mrRow, err := rs.QueryExecutor.ExecuteSelectRow(qryStr, false)
	if err != nil {
		return nil, err
	}
	batchMerkleRoot, err = rs.MerkleTreeQuery.ScanRoot(mrRow)
	if err != nil {
		if err != sql.ErrNoRows {
			return emptyReceipts, nil
		}
		return nil, err
	}

	// roll a random number based on lastBlock's previousBlock seed to select specific data from the looked up previousBlock
	// (lastblock height - BatchReceiptLookBackHeight)
	var (
		rng = crypto.NewRandomNumberGenerator()
		// hashList list of prev previousBlock hash + all previousBlock tx hashes (it should match all batch receipts the nodes already have)
		hashList = [][]byte{
			previousBlockHash,
		}
		rndDatumType uint32
	)
	// add transaction hashes of current block
	for _, tx := range lookBackBlockTransactions {
		hashList = append(hashList, tx.TransactionHash)
	}
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, blockSeed)
	if err != nil {
		return nil, err
	}
	rndSeedInt := rng.Next()
	rndSelSeed := rand.NewSource(rndSeedInt)
	rnd := rand.New(rndSelSeed)
	// rndSelectionIdx pseudo-random number in hashList array indexes
	rndSelectionIdx := rnd.Intn(len(hashList) - 1)
	// rndDatumHash hash to be used to find relative batch (node) receipts later on
	if rndSelectionIdx > len(hashList)-1 {
		return nil, errors.New("BatchReceiptIndexOutOfRange")
	}
	rndDatumHash := hashList[rndSelectionIdx]
	// first element of hashList array is always an hash relative to a block
	if rndSelectionIdx == 0 {
		rndDatumType = constant.ReceiptDatumTypeBlock
	} else {
		rndDatumType = constant.ReceiptDatumTypeTransaction
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
	priorityPeersAtHeight, err := func() (map[string]*model.Peer, error) {
		var scrambleNodes *model.ScrambledNodes
		scrambleNodes, err = rs.ScrambleNodeService.GetScrambleNodesByHeight(lookBackHeight)
		if err != nil {
			return nil, err
		}
		if scrambleNodes.NodePublicKeyToIDMap == nil {
			return nil, errors.New("NodePublicKeyToIDMapEmpty")
		}
		return rs.ReceiptUtil.GetPriorityPeersAtHeight(secretPhrase, scrambleNodes)
	}()
	if err != nil || len(priorityPeersAtHeight) == 0 {
		if err != nil {
			rs.Logger.Error(err)
		}
		// TODO: make sure we want to return empty receipts in case we can't find priority peers for current node at look back height
		// eg. if we return error instead, the block won't be generated/broadcast and this blocksmith would skip this round

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

	for _, unlinkedBatchReceipt := range unlinkedBatchReceiptList {
		unlinkedReceiptToPublish, err := rs.ReceiptUtil.GeneratePublishedReceipt(unlinkedBatchReceipt)
		if err != nil {
			rs.Logger.Error(err)
			continue
		}
		unlinkedPublishedReceipts = append(unlinkedPublishedReceipts, unlinkedReceiptToPublish)
	}
	return unlinkedPublishedReceipts, nil
}

// SelectUnlinkedReceipts select receipts received from node's priority peers at a given block height (from either a transaction or block
// broadcast to them by current node
func (rs *ReceiptService) SelectLinkedReceipts(
	numberOfReceipt, blockHeight uint32,
	blockSeed []byte,
	secretPhrase string,
) ([]*model.PublishedReceipt, error) {

	var (
		linkedReceipts []*model.PublishedReceipt
		referenceBlock *model.Block
		nodePublicKey  = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase)
		// maxLookBackwardSteps max n. of times this node should look backwards trying to link receipts when he was one of the scramble nodes
		// note: numberOfReceipts = number of max priority peers the node has
		maxLookBackwardSteps = numberOfReceipt
		emptyReceipts        = make([]*model.PublishedReceipt, 0)
	)

	// possible no connected node || lastblock height is too low to select receipts
	if numberOfReceipt == 0 || blockHeight < constant.BatchReceiptLookBackHeight {
		return emptyReceipts, nil
	}

	// loop backwards searching for blocks where current node was one of the block creators (when was in scramble node list)
	for refHeight := blockHeight; refHeight < 0; refHeight-- {
		var (
			refPublishedReceipt model.PublishedReceipt
		)

		if maxLookBackwardSteps == 0 {
			break
		}
		scrambleNodes, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(refHeight)
		if err != nil {
			return nil, err
		}
		// if not in scramble nodes try next block
		if _, ok := scrambleNodes.NodePublicKeyToIDMap[hex.EncodeToString(nodePublicKey)]; !ok {
			continue
		}
		// we found one block where the node was one of the scramble nodes
		maxLookBackwardSteps--
		// get the unlinked published receipt that has current node as recipient, at reference block height
		batchReceiptsQ, rootArgs := rs.PublishedReceiptQuery.GetUnlinkedPublishedReceiptByBlockHeightAndReceiver(refHeight, nodePublicKey)
		row, err := rs.QueryExecutor.ExecuteSelectRow(batchReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		err = rs.PublishedReceiptQuery.Scan(&refPublishedReceipt, row)
		if err != nil {
			if err != sql.ErrNoRows {
				// there are no published receipts to link a batch too, fail this block and continue
				continue
			}
			return nil, err
		}

		// take the 'reference height', 'reference block hash' and look up the batch receipts for that block.
		// If we do not have a corresponding batch for that height / hash, we fail this block (cannot link a receipt to it.) and continue
		var batchReceiptToLink model.BatchReceipt
		batchReceiptsQ, rootArgs = rs.NodeReceiptQuery.GetReceiptsByRefBlockHeightAndRefBlockHash(
			refPublishedReceipt.GetReceipt().ReferenceBlockHeight,
			refPublishedReceipt.GetReceipt().ReferenceBlockHash,
		)
		row, err = rs.QueryExecutor.ExecuteSelectRow(batchReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		err = rs.NodeReceiptQuery.Scan(&batchReceiptToLink, row)
		if err != nil {
			if err != sql.ErrNoRows {
				// there is no batch receipt that matches this published receipt ref height/hash. fail the block and continue
				continue
			}
			return nil, err
		}
		if err = rs.ValidateReceipt(batchReceiptToLink.Receipt, true); err != nil {
			// batch receipt is invalid. fail the block and continue
			continue
		}

		// get the block to select relative transaction hashes and previous block hash (similar to what we do for unlinked receipts)
		referenceBlock, err = util.GetBlockByHeight(
			batchReceiptToLink.GetReceipt().ReferenceBlockHeight,
			rs.QueryExecutor,
			rs.BlockQuery,
		)
		// this should not happen, since we already verify that this block exists when validating the receipt
		if err != nil {
			return nil, err
		}
		lookBackBlockTransactions, err := func() ([]*model.Transaction, error) {
			var transactions []*model.Transaction
			qry, args := rs.TransactionQuery.GetTransactionsByBlockID(referenceBlock.ID)
			rows, err := rs.QueryExecutor.ExecuteSelect(qry, false, args)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			return rs.TransactionQuery.BuildModel(transactions, rows)
		}()

		// roll a random number based on lastBlock's previousBlock seed to select specific data from the looked up previousBlock
		// (lastblock height - BatchReceiptLookBackHeight)
		var (
			rng = crypto.NewRandomNumberGenerator()
			// hashList list of prev previousBlock hash + all previousBlock tx hashes (it should match all batch receipts the nodes already have)
			hashList = [][]byte{
				referenceBlock.PreviousBlockHash,
			}
			referenceBlockHeight = referenceBlock.Height
			rndDatumType         uint32
		)
		// add transaction hashes of current block
		for _, tx := range lookBackBlockTransactions {
			hashList = append(hashList, tx.TransactionHash)
		}
		err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, blockSeed)
		if err != nil {
			return nil, err
		}
		rndSeedInt := rng.Next()
		rndSelSeed := rand.NewSource(rndSeedInt)
		rnd := rand.New(rndSelSeed)
		// rndSelectionIdx pseudo-random number in hashList array indexes
		rndSelectionIdx := rnd.Intn(len(hashList) - 1)
		// rndDatumHash hash to be used to find relative batch (node) receipts later on
		if rndSelectionIdx > len(hashList)-1 {
			return nil, errors.New("BatchReceiptIndexOutOfRange")
		}
		rndDatumHash := hashList[rndSelectionIdx]
		// first element of hashList array is always an hash relative to a block
		if rndSelectionIdx == 0 {
			rndDatumType = constant.ReceiptDatumTypeBlock
		} else {
			rndDatumType = constant.ReceiptDatumTypeTransaction
		}

		// the node looks up its position in scramble nodes at the referenceBlock height,
		// and computes its set of assigned receivers at that time (gets its priority peers at that height).
		priorityPeersAtHeight, err := func() ([]*model.Peer, error) {
			var (
				scrambleNodes      *model.ScrambledNodes
				priorityPeersArray []*model.Peer
				priorityPeersMap   map[string]*model.Peer
			)
			scrambleNodes, err = rs.ScrambleNodeService.GetScrambleNodesByHeight(referenceBlockHeight)
			if err != nil {
				return nil, err
			}
			if scrambleNodes.NodePublicKeyToIDMap == nil {
				return nil, errors.New("NodePublicKeyToIDMapEmpty")
			}
			priorityPeersMap, err = rs.ReceiptUtil.GetPriorityPeersAtHeight(secretPhrase, scrambleNodes)
			if err != nil {
				return nil, err
			}
			for _, pp := range priorityPeersMap {
				priorityPeersArray = append(priorityPeersArray, pp)
			}
			return priorityPeersArray, nil
		}()
		if err != nil || len(priorityPeersAtHeight) == 0 {
			if err != nil {
				rs.Logger.Error(err)
			}
			// TODO: lower this to debug once tested on alpha network to not pollute logs
			rs.Logger.Error("no priority peers for node at block height: ", referenceBlockHeight)
			continue
		}

		// Use the new block seed (from the block I am creating) to "roll" a random number to select one of my N assigned receivers.
		// This determines the “receiver” I must produce a receipt for.
		var (
			rndPeerIdx = rnd.Intn(len(priorityPeersAtHeight) - 1)
			rndPeer    = priorityPeersAtHeight[rndPeerIdx]
		)
		if rndPeer.GetInfo() == nil {
			return nil, errors.New("priorityPeerInfoEmpty")
		}

		// 2. Query my receipts in the batch, to find one that matches the data hash we rolled,
		// and also matches the receiver that we rolled. If we *do not* have a receipt in the batch that matches these 2,
		// then we fail this block (cannot link a receipt to it), and continue iterating backwards.
		var batchReceiptRmr model.BatchReceipt
		batchReceiptsQ, rootArgs = rs.NodeReceiptQuery.GetReceiptsByRecipientAndDatumHash(
			rndDatumHash,
			rndDatumType,
			rndPeer.GetInfo().PublicKey,
		)
		row, err = rs.QueryExecutor.ExecuteSelectRow(batchReceiptsQ, false, rootArgs...)
		if err != nil {
			return nil, err
		}
		err = rs.NodeReceiptQuery.Scan(&batchReceiptToLink, row)
		if err != nil {
			if err != sql.ErrNoRows {
				// there is no batch receipt that matches this ref hash and receiver. fail the block and continue
				continue
			}
			return nil, err
		}
		if err = rs.ValidateReceipt(batchReceiptToLink.Receipt, true); err != nil {
			// batch receipt is invalid. fail the block and continue
			continue
		}

		// 3. If we DO have a matching receipt (which we should, if we’ve been doing all of our work on the network),
		// this becomes one of our N "linked receipts", which we can publish in our new block.
		batchReceiptToPublish := &model.BatchReceipt{
			Receipt:  batchReceiptToLink.GetReceipt(),
			RMR:      batchReceiptRmr.RMR,
			RMRIndex: batchReceiptRmr.RMRIndex,
		}
		batchReceiptToPublish.Receipt.RMRLinked = batchReceiptToLink.RMR
		receiptToPublish, err := rs.ReceiptUtil.GeneratePublishedReceipt(batchReceiptToPublish)
		if err != nil {
			rs.Logger.Error(err)
			continue
		}
		linkedReceipts = append(linkedReceipts, receiptToPublish)
	}
	return linkedReceipts, nil
}

// SelectReceipts select list of (linked and unlinked) receipts to be included in a block
func (rs *ReceiptService) SelectReceipts(
	numberOfReceipt, blockHeight uint32,
	previousBlockHash, blockSeed []byte,
	secretPhrase string,
) ([]*model.PublishedReceipt, error) {
	var (
		err                              error
		unlinkedReceipts, linkedReceipts []*model.PublishedReceipt
	)

	// select unlinked receipts
	unlinkedReceipts, err = rs.SelectUnlinkedReceipts(
		numberOfReceipt,
		blockHeight,
		previousBlockHash,
		blockSeed,
		secretPhrase,
	)
	if err != nil {
		return nil, err
	}

	// select linked receipts
	linkedReceipts, err = rs.SelectLinkedReceipts(
		numberOfReceipt,
		blockHeight,
		blockSeed,
		secretPhrase,
	)
	if err != nil {
		return nil, err
	}

	return append(unlinkedReceipts, linkedReceipts...), nil
}

// TODO: delete this after successful test of new linked receipts logic
// func (rs *ReceiptService) pickReceipts(
// 	numberOfReceipt uint32,
// 	pickedReceipts []*model.PublishedReceipt,
// 	pickedRecipients map[string]bool,
// 	lowerBlockHeight, upperBlockHeight uint32,
// ) ([]*model.PublishedReceipt, error) {
// 	var receipts []*model.BatchReceipt
// 	receipts, err := func() ([]*model.BatchReceipt, error) {
// 		receiptsQ := rs.NodeReceiptQuery.GetReceiptsWithUniqueRecipient(
// 			numberOfReceipt*constant.ReceiptBatchPickMultiplier, lowerBlockHeight, upperBlockHeight)
// 		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer rows.Close()
// 		return rs.NodeReceiptQuery.BuildModel(receipts, rows)
// 	}()
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, rc := range receipts {
// 		if len(pickedReceipts) >= int(numberOfReceipt) {
// 			break
// 		}
// 		errValid := rs.ValidateReceipt(rc.GetReceipt(), true)
// 		if errValid != nil {
// 			// skipped invalid receipt
// 			continue
// 		}
// 		if !pickedRecipients[string(rc.GetReceipt().RecipientPublicKey)] {
// 			pickedReceipts = append(pickedReceipts, &model.PublishedReceipt{
// 				Receipt:            rc.GetReceipt(),
// 				IntermediateHashes: nil,
// 				ReceiptIndex:       rc.RMRIndex,
// 			})
// 			pickedRecipients[string(rc.Receipt.RecipientPublicKey)] = true
// 		}
// 	}
// 	return pickedReceipts, nil
// }

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
				Receipt:  &b,
				RMR:      rootMerkle,
				RMRIndex: uint32(k),
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
		true,
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
