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
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/observer"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"math/rand"
	"sort"
	"time"

	"golang.org/x/crypto/ed25519"

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
				// wait for node to collect all receipt in case of heavy network load
				time.Sleep(constant.BatchReceiptWaitingTime)
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
func (rs *ReceiptService) SelectReceiptsOld(
	blockTimestamp int64,
	numberOfReceipt, lastBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var (
		linkedReceiptList = make(map[string][]*model.BatchReceipt)
		// this variable is to store picked receipt recipient to avoid duplicates
		pickedRecipients  = make(map[string]bool)
		lowerBlockHeight  uint32
		linkedReceiptTree = make(map[string][]byte)
		err               error
	)

	if numberOfReceipt < 1 { // possible no connected node
		return []*model.PublishedReceipt{}, nil
	}
	// get the last merkle tree we have build so far
	if lastBlockHeight > constant.MinRollbackBlocks {
		lowerBlockHeight = lastBlockHeight - constant.MinRollbackBlocks
	}

	linkedReceiptTree, err = func() (map[string][]byte, error) {
		treeQ := rs.MerkleTreeQuery.SelectMerkleTreeForPublishedReceipts(
			lowerBlockHeight,
			lastBlockHeight,
			numberOfReceipt*constant.ReceiptBatchPickMultiplier)
		linkedTreeRows, err := rs.QueryExecutor.ExecuteSelect(treeQ, false)
		if err != nil {
			return linkedReceiptTree, err
		}
		defer linkedTreeRows.Close()

		return rs.MerkleTreeQuery.BuildTree(linkedTreeRows)
	}()
	if err != nil {
		return nil, err
	}
	for linkedRoot := range linkedReceiptTree {
		var nodeReceipts []*model.BatchReceipt

		nodeReceipts, err = func() ([]*model.BatchReceipt, error) {
			nodeReceiptsQ, rootArgs := rs.NodeReceiptQuery.GetReceiptsByRootInRange(lowerBlockHeight, lastBlockHeight, []byte(linkedRoot))
			rows, err := rs.QueryExecutor.ExecuteSelect(nodeReceiptsQ, false, rootArgs...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			return rs.NodeReceiptQuery.BuildModel(nodeReceipts, rows)
		}()
		if err != nil {
			return nil, err
		}
		for _, rc := range nodeReceipts {
			if !pickedRecipients[string(rc.GetReceipt().GetRecipientPublicKey())] {
				pickedRecipients[string(rc.GetReceipt().GetRecipientPublicKey())] = true
				linkedReceiptList[linkedRoot] = append(linkedReceiptList[linkedRoot], rc)
			}
		}
	}
	// limit the selected portion to `numberOfReceipt` receipts
	// filter the selected receipts on second phase
	var (
		results []*model.PublishedReceipt
	)
	for rcRoot, rcReceipt := range linkedReceiptList {
		merkle := util.MerkleRoot{}
		merkle.HashTree = merkle.FromBytes(linkedReceiptTree[rcRoot], []byte(rcRoot))
		for _, rc := range rcReceipt {
			if len(results) >= int(numberOfReceipt) {
				break
			}
			err = rs.ValidateReceipt(rc.Receipt)
			if err != nil {
				// skip invalid receipt
				continue
			}
			var intermediateHashes [][]byte
			rcByte := rs.ReceiptUtil.GetSignedReceiptBytes(rc.Receipt)
			rcHash := sha3.Sum256(rcByte)

			intermediateHashesBuffer := merkle.GetIntermediateHashes(
				bytes.NewBuffer(rcHash[:]),
				int32(rc.RMRIndex),
			)
			for _, buf := range intermediateHashesBuffer {
				intermediateHashes = append(intermediateHashes, buf.Bytes())
			}
			results = append(
				results,
				&model.PublishedReceipt{
					Receipt:            rc.GetReceipt(),
					IntermediateHashes: merkle.FlattenIntermediateHashes(intermediateHashes),
					ReceiptIndex:       rc.RMRIndex,
				},
			)
		}
	}
	// select non-linked receipt
	if len(results) < int(numberOfReceipt) {
		rmrLinkedReceipts, err := rs.pickReceipts(
			numberOfReceipt, results, pickedRecipients, lowerBlockHeight, lastBlockHeight)
		if err != nil {
			return nil, err
		}
		results = rmrLinkedReceipts
	}

	return results, nil
}

// SelectUnlinkedReceipts select receipts received from node's priority peers at a given block height (from either a transaction or block
// broadcast to them by current node
func (rs *ReceiptService) SelectUnlinkedReceipts(
	numberOfReceipt, blockHeight uint32,
	previousBlockHash, blockSeed []byte,
	secretPhrase string,
) ([]*model.BatchReceipt, error) {
	var (
		err                       error
		batchMerkleRoot           []byte
		batchReceipts             []*model.BatchReceipt
		lookBackBlock             model.Block
		lookBackBlockTransactions []*model.Transaction
		unlinkedReceiptList       []*model.BatchReceipt
		lookBackHeight            = blockHeight - constant.BatchReceiptLookBackHeight
		emptyReceipts             = make([]*model.BatchReceipt, 0)
	)

	// possible no connected node || lastblock height is too low to select receipts
	if numberOfReceipt == 0 || blockHeight < constant.BatchReceiptLookBackHeight {
		return emptyReceipts, nil
	}

	// get the reference block's transactions to select batch receipts from
	blRow, err := rs.QueryExecutor.ExecuteSelectRow(rs.BlockQuery.GetBlockByHeight(lookBackHeight), false)
	if err != nil {
		return nil, err
	}
	if err = rs.BlockQuery.Scan(&lookBackBlock, blRow); err != nil {
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
	mrRow, err := rs.QueryExecutor.ExecuteSelectRow(rs.MerkleTreeQuery.SelectMerkleTreeAtHeight(lookBackHeight), false)
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
	if len(batchReceipts) == 0 {
		return nil, nil
	}

	// the node looks up its position in scramble nodes at the look back height,
	// and computes its set of assigned receivers at that time (gets its priority peers at that height).
	priorityPeersAtHeight, err := func() (map[string]*model.Peer, error) {
		nodePubKey := signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase)
		scrambleNodes, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(lookBackHeight)
		if scrambleNodes.NodePublicKeyToIDMap == nil {
			return make(map[string]*model.Peer, 0), nil
		}
		scrambleNodeID, ok := scrambleNodes.NodePublicKeyToIDMap[hex.EncodeToString(nodePubKey)]
		if !ok {
			// return empty priority peers list if current node is not found in scramble nodes at look back height
			return make(map[string]*model.Peer, 0), nil
		}
		peers, err := p2pUtil.GetPriorityPeersByNodeID(
			scrambleNodeID,
			scrambleNodes,
		)
		if err != nil {
			return make(map[string]*model.Peer, 0), nil
		}
		return peers, nil
	}()
	if len(priorityPeersAtHeight) == 0 {
		// TODO: make sure we want to return empty receipts in case we can't find priority peers for current node at look back height
		// eg. if we return error instead, the block won't be generated/broadcast and this blocksmith would skip this round
		return emptyReceipts, nil
	}

	// select all batch (node) receipts, collected at look back height, from one of the priority peers (at that height),
	// that match the random datum hash (transaction or block hash) rolled earlier
	for _, priorityPeer := range priorityPeersAtHeight {
		for _, batchReceipt := range batchReceipts {
			if bytes.Equal(batchReceipt.GetReceipt().RecipientPublicKey, priorityPeer.GetInfo().GetPublicKey()) {
				unlinkedReceiptList = append(unlinkedReceiptList, batchReceipt)
			}
		}
	}
	return unlinkedReceiptList, nil
}

// SelectUnlinkedReceipts select receipts received from node's priority peers at a given block height (from either a transaction or block
// broadcast to them by current node
func (rs *ReceiptService) SelectLinkedReceipts(
	numberOfReceipt, blockHeight uint32,
	previousBlockHash, blockSeed []byte,
	secretPhrase string,
) ([]*model.BatchReceipt, error) {

	var (
		linkedReceipts []*model.BatchReceipt
		emptyReceipts  = make([]*model.BatchReceipt, 0)
		nodePublicKey  = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase)
		// maxLookBackwardSteps max n. of times this node should look backwards trying to link receipts when he was one of the scramble nodes
		// note: numberOfReceipts = number of max priority peers the node has
		maxLookBackwardSteps = numberOfReceipt
	)

	// loop backwards searching for blocks where current node was one of the block creators (when was in scramble node list)
	for refHeight := blockHeight; refHeight >= 0; refHeight-- {
		if maxLookBackwardSteps <= 0 {
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
		var (
			batchReceipts     []*model.BatchReceipt
			publishedReceipts []*model.PublishedReceipt
		)
		// get all batch receipts for selected merkle root and datum_hash ordered by recipient_public_key, reference_block_height
		publishedReceipts, err = func() ([]*model.PublishedReceipt, error) {
			var receipts []*model.PublishedReceipt
			batchReceiptsQ, rootArgs := rs.PublishedReceiptQuery.GetPublishedReceiptByBlockHeightWithMerkleRoot(refHeight)
			rows, err := rs.QueryExecutor.ExecuteSelect(batchReceiptsQ, false, rootArgs...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			return rs.PublishedReceiptQuery.BuildModel(receipts, rows)
		}()
		if err != nil {
			return nil, err
		}
		if len(publishedReceipts) == 0 {
			continue
		}
		// If there is a receipt where I was the receiver: take the 'reference height', 'reference block hash',
		// and 'receipt merkle root' from that receipt. Use any of them to look up my batch for that block.
		// If we do not have a corresponding batch for that height/merkle root, we fail this block (cannot link a receipt to it.)
		var batchFound = false
		for _, publishedReceipt := range publishedReceipts {
			// check if this published receipt comes from this node
			if bytes.Equal(publishedReceipt.GetReceipt().RecipientPublicKey, nodePublicKey) {
				// check if we have a batch with merkle root that matches the ones in current published receipt
				batchReceipts, err = func() ([]*model.BatchReceipt, error) {
					var receipts []*model.BatchReceipt
					batchReceiptsQ, rootArgs := rs.NodeReceiptQuery.GetReceiptsByRoot(publishedReceipt.GetReceipt().RMRLinked)
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
				// no receipts to link to this block
				if len(batchReceipts) > 0 {
					batchFound = true
				}
			}
		}

		// cannot link any receipt for this block
		if !batchFound {
			continue
		}

		//STEF continue from here!

		return linkedReceipts, nil
	}

}

// SelectReceipts select list of (linked and unlinked) receipts to be included in a block
func (rs *ReceiptService) SelectReceipts(
	numberOfReceipt, blockHeight uint32,
	previousBlockHash, blockSeed []byte,
	secretPhrase string,
) ([]*model.PublishedReceipt, error) {
	var (
		err                 error
		unlinkedReceiptList []*model.BatchReceipt
	)

	// select unlinked receipts
	unlinkedReceiptList, err := rs.SelectUnlinkedReceipts(
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
}

func (rs *ReceiptService) pickReceipts(
	numberOfReceipt uint32,
	pickedReceipts []*model.PublishedReceipt,
	pickedRecipients map[string]bool,
	lowerBlockHeight, upperBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var receipts []*model.BatchReceipt
	receipts, err := func() ([]*model.BatchReceipt, error) {
		receiptsQ := rs.NodeReceiptQuery.GetReceiptsWithUniqueRecipient(
			numberOfReceipt*constant.ReceiptBatchPickMultiplier, lowerBlockHeight, upperBlockHeight)
		rows, err := rs.QueryExecutor.ExecuteSelect(receiptsQ, false)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return rs.NodeReceiptQuery.BuildModel(receipts, rows)
	}()
	if err != nil {
		return nil, err
	}
	for _, rc := range receipts {
		if len(pickedReceipts) >= int(numberOfReceipt) {
			break
		}
		errValid := rs.ValidateReceipt(rc.GetReceipt())
		if errValid != nil {
			// skipped invalid receipt
			continue
		}
		if !pickedRecipients[string(rc.GetReceipt().RecipientPublicKey)] {
			pickedReceipts = append(pickedReceipts, &model.PublishedReceipt{
				Receipt:            rc.GetReceipt(),
				IntermediateHashes: nil,
				ReceiptIndex:       rc.RMRIndex,
			})
			pickedRecipients[string(rc.Receipt.RecipientPublicKey)] = true
		}
	}
	return pickedReceipts, nil
}

// GenerateReceiptsMerkleRoot generate merkle root of some batch receipts and also remove from cache
// generating will do when number of collected receipts(batch receipts) already <= the number of required
func (rs *ReceiptService) GenerateReceiptsMerkleRoot(block *model.Block) error {
	var (
		receiptsCached, receipts    []model.Receipt
		hashedReceipts              []*bytes.Buffer
		merkleRoot                  util.MerkleRoot
		queries                     [][]interface{}
		batchReceipt                *model.BatchReceipt
		err                         error
		rootMerkle, treeMerkle      []byte
		blockAtHeight               *storage.BlockCacheObject
		isDbTransactionHighPriority = false
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

	if err := rs.BatchReceiptCacheStorage.GetAllItems(&receiptsCached); err != nil {
		return err
	}
	// If no receipts in cache no need to return errors. just log a message
	if len(receiptsCached) == 0 {
		rs.Logger.Info("No Receipts for block height: ", block.Height)
		return nil
	}
	// Need to sorting before do next
	sort.SliceStable(receiptsCached, func(i, j int) bool {
		return receiptsCached[i].ReferenceBlockHeight < receiptsCached[j].ReferenceBlockHeight
	})

	var (
		receiptsToProcess = make([]model.Receipt, 0)
		remainingReceipts = make([]model.Receipt, 0)
		// note that expirationHeight cannot be negative because is a uint32
		expirationHeight = block.Height - constant.ReceiptPoolMaxLife
	)
	// Extract from receipt pool only the ones that reference current block
	for _, receipt := range receiptsCached {
		if receipt.ReferenceBlockHeight == block.Height && bytes.Equal(receipt.ReferenceBlockHash, block.BlockHash) {
			receiptsToProcess = append(receiptsToProcess, receipt)
		} else if receipt.ReferenceBlockHeight > constant.ReceiptPoolMaxLife && receipt.ReferenceBlockHeight < expirationHeight {
			continue
		} else {
			remainingReceipts = append(remainingReceipts, receipt)
		}
	}

	err = rs.QueryExecutor.BeginTx(isDbTransactionHighPriority, monitoring.GenerateReceiptsMerkleRootOwnerProcess)
	if err != nil {
		return err
	}
	err = func() error {
		for _, receiptToProcess := range receiptsToProcess {
			b := receiptToProcess
			err = rs.ValidateReceipt(&b)
			if err == nil {
				receipts = append(receipts, b)
				hashedReceipt := sha3.Sum256(rs.ReceiptUtil.GetSignedReceiptBytes(&b))
				hashedReceipts = append(hashedReceipts, bytes.NewBuffer(hashedReceipt[:]))
			}
		}
		_, err = merkleRoot.GenerateMerkleRoot(hashedReceipts)
		if err != nil {
			return err
		}
		rootMerkle, treeMerkle = merkleRoot.ToBytes()

		queries = make([][]interface{}, len(hashedReceipts)+1)
		for k, receipt := range receipts {
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
	return rs.BatchReceiptCacheStorage.SetItems(remainingReceipts)
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
) error {
	var (
		blockAtHeight *storage.BlockCacheObject
		err           error
	)
	if len(receipt.GetRecipientPublicKey()) != ed25519.PublicKeySize {
		return blocker.NewBlocker(blocker.ValidationErr,
			"[SendBlockTransactions:MaliciousReceipt] - %d is %s",
			len(receipt.GetRecipientPublicKey()),
			"InvalidReceiptRecipientPublicKeySize",
		)
	}
	if len(receipt.GetRecipientSignature()) != ed25519.SignatureSize {
		return blocker.NewBlocker(blocker.ValidationErr,
			"[SendBlockTransactions:MaliciousReceipt] - %d is %s",
			len(receipt.GetRecipientPublicKey()),
			"InvalidReceiptSignatureSize",
		)
	}

	unsignedBytes := rs.ReceiptUtil.GetUnsignedReceiptBytes(receipt)
	if !rs.Signature.VerifyNodeSignature(
		unsignedBytes,
		receipt.RecipientSignature,
		receipt.RecipientPublicKey,
	) {
		// rollback
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidReceiptSignature",
		)
	}
	blockAtHeight, err = util.GetBlockByHeightUseBlocksCache(
		receipt.ReferenceBlockHeight,
		rs.QueryExecutor,
		rs.BlockQuery,
		rs.MainBlocksStorage,
	)
	if err != nil {
		return err
	}
	// check block hash
	if !bytes.Equal(blockAtHeight.BlockHash, receipt.ReferenceBlockHash) {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidReceiptBlockHash")
	}
	// get or build scramble nodes at height
	scrambleNode, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(receipt.ReferenceBlockHeight)
	if err != nil {
		return err
	}
	err = rs.ReceiptUtil.ValidateReceiptSenderRecipient(receipt, scrambleNode)
	if err != nil {
		return err
	}
	return nil
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
	err = rs.BatchReceiptCacheStorage.SetItem(nil, b)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReceiptService) ClearCache() {
	_ = rs.BatchReceiptCacheStorage.ClearCache()
	_ = rs.ReceiptReminderStorage.ClearCache()
}
