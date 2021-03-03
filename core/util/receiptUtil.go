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
package util

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/query"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"math/rand"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptUtilInterface interface {
		GetNumberOfMaxReceipts(numberOfSortedBlocksmiths int) uint32
		GenerateReceipt(
			ct chaintype.ChainType,
			referenceBlock *storage.BlockCacheObject,
			senderPublicKey, recipientPublicKey, datumHash, rmrLinked []byte,
			datumType uint32,
		) (*model.Receipt, error)
		GetUnsignedReceiptBytes(receipt *model.Receipt) []byte
		GetSignedReceiptBytes(receipt *model.Receipt) []byte
		GetReceiptKey(
			dataHash, senderPublicKey []byte,
		) ([]byte, error)
		ValidateReceiptHelper(
			receipt *model.Receipt,
			validateRefBlock bool,
			executor query.ExecutorInterface,
			blockQuery query.BlockQueryInterface,
			mainBlockStorage storage.CacheStackStorageInterface,
			signature crypto.SignatureInterface,
			scrambleNodesAtHeight *model.ScrambledNodes,
		) error
		ValidateReceiptSenderRecipient(
			receipt *model.Receipt,
			scrambledNode *model.ScrambledNodes,
		) error
		GetPriorityPeersAtHeight(
			nodePubKey []byte,
			scrambleNodes *model.ScrambledNodes,
		) (map[string]*model.Peer, error)
		GeneratePublishedReceipt(
			batchReceipt *model.BatchReceipt,
		) (*model.PublishedReceipt, error)
		IsPublishedReceiptEqual(a, b *model.PublishedReceipt) error
		BuildBlockDatumHashes(
			block *model.Block,
			executor query.ExecutorInterface,
			transactionQuery query.TransactionQueryInterface,
		) ([][]byte, error)
		GetRandomDatumHash(hashList [][]byte, blockSeed []byte) (rndDatumHash []byte, rndDatumType uint32, err error)
	}

	ReceiptUtil struct{}
)

func NewReceiptUtil() *ReceiptUtil {
	return &ReceiptUtil{}
}

// buildBlockDatumHashes build an array containing all hashes of data secured by the block (transactions and previous block)
func (ru *ReceiptUtil) BuildBlockDatumHashes(
	block *model.Block,
	executor query.ExecutorInterface,
	transactionQuery query.TransactionQueryInterface,
) ([][]byte, error) {
	var (
		// hashList list of prev previousBlock hash + all previousBlock tx hashes (it should match all batch receipts the nodes already have)
		hashList = [][]byte{
			block.PreviousBlockHash,
		}
	)
	blockTransactions, err := func() ([]*model.Transaction, error) {
		var transactions []*model.Transaction
		qry, args := transactionQuery.GetTransactionsByBlockID(block.ID)
		rows, err := executor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return transactionQuery.BuildModel(transactions, rows)
	}()
	if err != nil {
		return nil, err
	}
	// add transaction hashes of current block
	for _, tx := range blockTransactions {
		hashList = append(hashList, tx.TransactionHash)
	}
	return hashList, nil
}

func (ru *ReceiptUtil) GetRandomDatumHash(hashList [][]byte, blockSeed []byte) (rndDatumHash []byte, rndDatumType uint32, err error) {
	// roll a random number based on lastBlock's previousBlock seed to select specific data from the looked up previousBlock
	// (lastblock height - BatchReceiptLookBackHeight)
	var (
		rng             = crypto.NewRandomNumberGenerator()
		rndSelectionIdx int
	)

	// instantiate a pseudo random number generator using block seed (new block) as random seed
	err = rng.Reset(constant.BlocksmithSelectionSeedPrefix, blockSeed)
	if err != nil {
		return nil, 0, err
	}
	rndSeedInt := rng.Next()
	rndSelSeed := rand.NewSource(rndSeedInt)
	rnd := rand.New(rndSelSeed)
	// rndSelectionIdx pseudo-random number in hashList array indexes
	if len(hashList) > 1 {
		rndSelectionIdx = rnd.Intn(len(hashList) - 1)
	}
	// rndDatumHash hash to be used to find relative batch (node) receipts later on
	if rndSelectionIdx > len(hashList)-1 {
		return nil, 0, errors.New("BatchReceiptIndexOutOfRange")
	}
	rndDatumHash = hashList[rndSelectionIdx]
	// first element of hashList array is always an hash relative to a block
	if rndSelectionIdx == 0 {
		rndDatumType = constant.ReceiptDatumTypeBlock
	} else {
		rndDatumType = constant.ReceiptDatumTypeTransaction
	}
	return rndDatumHash, rndDatumType, nil
}

func (ru *ReceiptUtil) GeneratePublishedReceipt(
	batchReceipt *model.BatchReceipt,
) (*model.PublishedReceipt, error) {
	var (
		intermediateHashes [][]byte
		merkle             = util.MerkleRoot{}
		receipt            = batchReceipt.GetReceipt()
	)

	rcByte := ru.GetSignedReceiptBytes(receipt)
	rcHash := sha3.Sum256(rcByte)

	intermediateHashesBuffer := merkle.GetIntermediateHashes(
		bytes.NewBuffer(rcHash[:]),
		int32(batchReceipt.RMRIndex),
	)
	for _, buf := range intermediateHashesBuffer {
		intermediateHashes = append(intermediateHashes, buf.Bytes())
	}
	return &model.PublishedReceipt{
		Receipt:            receipt,
		IntermediateHashes: merkle.FlattenIntermediateHashes(intermediateHashes),
		ReceiptIndex:       batchReceipt.RMRIndex,
		RMR:                batchReceipt.RMR,
	}, nil
}

// GetPriorityPeersAtHeight get priority peers map with peer public key (hex) as map key
func (ru *ReceiptUtil) GetPriorityPeersAtHeight(
	nodePubKey []byte,
	scrambleNodes *model.ScrambledNodes,
) (map[string]*model.Peer, error) {
	var peersByPubKeyMap = make(map[string]*model.Peer, 0)
	scrambleNodeID, ok := scrambleNodes.NodePublicKeyToIDMap[hex.EncodeToString(nodePubKey)]
	if !ok {
		// return empty priority peers list if current node is not found in scramble nodes at look back height
		return make(map[string]*model.Peer), nil
	}
	peers, err := p2pUtil.GetPriorityPeersByNodeID(
		scrambleNodeID,
		scrambleNodes,
	)
	if err != nil {
		return make(map[string]*model.Peer), nil
	}
	for _, peer := range peers {
		peersByPubKeyMap[hex.EncodeToString(peer.Info.PublicKey)] = peer
	}
	return peers, nil
}

// ValidateReceiptHelper helper function for better code testability
func (ru *ReceiptUtil) ValidateReceiptHelper(
	receipt *model.Receipt,
	validateRefBlock bool,
	executor query.ExecutorInterface,
	blockQuery query.BlockQueryInterface,
	mainBlockStorage storage.CacheStackStorageInterface,
	signature crypto.SignatureInterface,
	scrambleNodesAtHeight *model.ScrambledNodes,
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

	unsignedBytes := ru.GetUnsignedReceiptBytes(receipt)
	if !signature.VerifyNodeSignature(
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

	// validate reference block hash only if necessary
	// Eg. when collecting batch receipts from peers, we don't want to check if the receipt come from a fork (or we are in a temporary fork,
	// thus we would not collect a good receipt).
	if validateRefBlock {
		blockAtHeight, err = util.GetBlockByHeightUseBlocksCache(
			receipt.ReferenceBlockHeight,
			executor,
			blockQuery,
			mainBlockStorage,
		)
		if err != nil {
			return err
		}
		// check block hash
		if !bytes.Equal(blockAtHeight.BlockHash, receipt.ReferenceBlockHash) {
			return blocker.NewBlocker(blocker.ValidationErr, "InvalidReceiptBlockHash")
		}
	}

	err = ru.ValidateReceiptSenderRecipient(receipt, scrambleNodesAtHeight)
	if err != nil {
		return err
	}
	return nil
}

func (ru *ReceiptUtil) ValidateReceiptSenderRecipient(
	receipt *model.Receipt,
	scrambledNode *model.ScrambledNodes,
) error {
	var (
		err   error
		peers map[string]*model.Peer
	)
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

func (ru *ReceiptUtil) GetNumberOfMaxReceipts(numberOfSortedBlocksmiths int) uint32 {
	if numberOfSortedBlocksmiths < 1 {
		return 0 // avoid overflow
	}
	if (numberOfSortedBlocksmiths - 1) < constant.PriorityStrategyMaxPriorityPeers {
		// return all blocksmiths excepth the node itself
		return uint32(numberOfSortedBlocksmiths - 1)
	}
	return constant.PriorityStrategyMaxPriorityPeers
}

// GenerateReceipt generate receipt object that act as proof of receipt on data. Data received can be
// block, transaction, etc.
// generated receipt will not be signed yet (RecipientSignature = nil), will need to be signed using SignReceipt method.
func (ru *ReceiptUtil) GenerateReceipt(
	ct chaintype.ChainType,
	referenceBlock *storage.BlockCacheObject,
	senderPublicKey, recipientPublicKey, datumHash, rmrLinked []byte,
	datumType uint32,
) (*model.Receipt, error) {

	return &model.Receipt{
		SenderPublicKey:      senderPublicKey,
		RecipientPublicKey:   recipientPublicKey,
		DatumType:            datumType,
		DatumHash:            datumHash,
		ReferenceBlockHeight: referenceBlock.Height,
		ReferenceBlockHash:   referenceBlock.BlockHash,
		RMRLinked:            rmrLinked,
	}, nil
}

// GetUnsignedReceiptBytes Client task while doing validation signature
func (ru *ReceiptUtil) GetUnsignedReceiptBytes(receipt *model.Receipt) []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(receipt.SenderPublicKey)
	buffer.Write(receipt.RecipientPublicKey)
	buffer.Write(util.ConvertUint32ToBytes(receipt.ReferenceBlockHeight))
	buffer.Write(receipt.ReferenceBlockHash)
	buffer.Write(util.ConvertUint32ToBytes(receipt.DatumType))
	buffer.Write(receipt.DatumHash)
	buffer.Write(receipt.RMRLinked)
	return buffer.Bytes()
}

// GetSignedReceiptBytes Client task before store into database batch_receipt
func (ru *ReceiptUtil) GetSignedReceiptBytes(receipt *model.Receipt) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(receipt.SenderPublicKey)
	buffer.Write(receipt.RecipientPublicKey)
	buffer.Write(util.ConvertUint32ToBytes(receipt.ReferenceBlockHeight))
	buffer.Write(receipt.ReferenceBlockHash)
	buffer.Write(util.ConvertUint32ToBytes(receipt.DatumType))
	buffer.Write(receipt.DatumHash)
	buffer.Write(receipt.RMRLinked)
	buffer.Write(receipt.RecipientSignature)
	return buffer.Bytes()
}

func (ru *ReceiptUtil) GetReceiptKey(
	dataHash, senderPublicKey []byte,
) ([]byte, error) {
	digest := sha3.New256()
	_, err := digest.Write(dataHash)
	if err != nil {
		return nil, err
	}
	_, err = digest.Write(senderPublicKey)
	if err != nil {
		return nil, err
	}
	receiptKey := digest.Sum([]byte{})
	return receiptKey, nil
}

func (ru *ReceiptUtil) IsPublishedReceiptEqual(a, b *model.PublishedReceipt) error {
	if a.BlockHeight != b.BlockHeight {
		return errors.New("BlockHeight")
	}
	if a.Receipt.ReferenceBlockHeight != a.Receipt.ReferenceBlockHeight {
		return errors.New("ReferenceBlockHeight")
	}
	if !bytes.Equal(a.Receipt.ReferenceBlockHash, b.Receipt.ReferenceBlockHash) {
		return errors.New("ReferenceBlockHash")
	}
	if !bytes.Equal(a.Receipt.RecipientPublicKey, b.Receipt.RecipientPublicKey) {
		return errors.New("RecipientPubKey")
	}
	if !bytes.Equal(a.Receipt.SenderPublicKey, b.Receipt.SenderPublicKey) {
		return errors.New("SenderPubKey")
	}
	if !bytes.Equal(a.Receipt.RMRLinked, b.Receipt.RMRLinked) {
		return errors.New("RMRLinked")
	}

	return nil
}
