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
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptServiceInterface interface {
		Initialize() error
		SelectReceipts(
			blockTimestamp int64,
			numberOfReceipt uint32,
			lastBlockHeight uint32,
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
				rs.GenerateReceiptsMerkleRoot(b)
			}
		},
	}
}

func NewReceiptService(
	nodeReceiptQuery query.BatchReceiptQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	nodeRegistrationQuery query.NodeRegistrationQueryInterface,
	blockQuery query.BlockQueryInterface,
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
func (rs *ReceiptService) SelectReceipts(
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
		treeQ := rs.MerkleTreeQuery.SelectMerkleTree(
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
			nodeReceiptsQ, rootArgs := rs.NodeReceiptQuery.GetReceiptByRoot(lowerBlockHeight, lastBlockHeight, []byte(linkedRoot))
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

func (rs *ReceiptService) pickReceipts(
	numberOfReceipt uint32,
	pickedReceipts []*model.PublishedReceipt,
	pickedRecipients map[string]bool,
	lowerBlockHeight, upperBlockHeight uint32,
) ([]*model.PublishedReceipt, error) {
	var receipts []*model.BatchReceipt
	receipts, err := func() ([]*model.BatchReceipt, error) {
		// STEF check this logic: seems doesn't select anything even when there are published receipts to be picked up
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

// TODO: add a scheduled job to prune old receipts after n (MinRollbackSafe?) blocks
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

	if err = rs.BatchReceiptCacheStorage.GetAllItems(&receiptsCached); err != nil {
		return err
	}
	// If no receipts in cache no need to log errors
	if len(receiptsCached) == 0 {
		// TODO: lower this to level 'Info' to not pollute the logs
		rs.Logger.Error("No Receipts for block height: ", block.Height)
		return nil
	}
	// Need to sorting before do next
	sort.SliceStable(receiptsCached, func(i, j int) bool {
		return receiptsCached[i].ReferenceBlockHeight < receiptsCached[j].ReferenceBlockHeight
	})

	var (
		receiptsToProcess = make([]model.Receipt, 0)
		remainingReceipts = make([]model.Receipt, 0)
	)
	// Extract from receipt pool only the ones that reference current block
	for _, receipt := range receiptsCached {
		if receipt.ReferenceBlockHeight == block.Height && bytes.Equal(receipt.ReferenceBlockHash, block.BlockHash) {
			receiptsToProcess = append(receiptsToProcess, receipt)
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
	err = rs.validateReceiptSenderRecipient(receipt)
	if err != nil {
		return err
	}
	return nil
}

func (rs *ReceiptService) validateReceiptSenderRecipient(
	receipt *model.Receipt,
) error {
	var (
		err   error
		peers map[string]*model.Peer
	)
	// get or build scrambled nodes at height
	scrambledNode, err := rs.ScrambleNodeService.GetScrambleNodesByHeight(receipt.ReferenceBlockHeight)
	if err != nil {
		return err
	}
	// get sender address at height
	senderNodeID, ok := scrambledNode.NodePublicKeyToIDMap[hex.EncodeToString(receipt.GetSenderPublicKey())]
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptSenderNotInScrambleList")
	}
	// get recipient address at height
	recipientNodeID, ok := scrambledNode.NodePublicKeyToIDMap[hex.EncodeToString(receipt.GetRecipientPublicKey())]
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "ReceiptRecipientNotInScrambleList")
	}
	if peers, err = p2pUtil.GetPriorityPeersByNodeID(
		senderNodeID,
		scrambledNode,
	); err != nil {
		return err
	}

	// check if recipient is in sender.Peers list
	for _, peer := range peers {
		if peer.GetInfo().ID == recipientNodeID {
			// valid recipient and sender
			return nil
		}
	}
	return blocker.NewBlocker(blocker.ValidationErr, "ReceiptRecipientNotInPriorityList")
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
