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
	"encoding/hex"
	"github.com/zoobc/zoobc-core/common/blocker"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"

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
		ValidateReceiptSenderRecipient(
			receipt *model.Receipt,
			scrambledNode *model.ScrambledNodes,
		) error
	}

	ReceiptUtil struct{}
)

func NewReceiptUtil() *ReceiptUtil {
	return &ReceiptUtil{}
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
