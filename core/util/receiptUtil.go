package util

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	ReceiptUtilInterface interface {
		GetNumberOfMaxReceipts(numberOfSortedBlocksmiths int) uint32

		GenerateBatchReceipt(
			ct chaintype.ChainType,
			referenceBlock *model.Block,
			senderPublicKey, recipientPublicKey, datumHash, rmrLinked []byte,
			datumType uint32,
		) (*model.BatchReceipt, error)

		GetUnsignedBatchReceiptBytes(
			receipt *model.BatchReceipt,
		) []byte

		GetSignedBatchReceiptBytes(receipt *model.BatchReceipt) []byte

		GetReceiptKey(
			dataHash, senderPublicKey []byte,
		) ([]byte, error)
	}

	ReceiptUtil struct{}
)

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

// GenerateBatchReceipt generate receipt object that act as proof of receipt on data. Data received can be
// block, transaction, etc.
// generated receipt will not be signed yet (RecipientSignature = nil), will need to be signed using SignReceipt method.
func (ru *ReceiptUtil) GenerateBatchReceipt(
	ct chaintype.ChainType,
	referenceBlock *model.Block,
	senderPublicKey, recipientPublicKey, datumHash, rmrLinked []byte,
	datumType uint32,
) (*model.BatchReceipt, error) {
	refBlockHash, err := util.GetBlockHash(referenceBlock, ct)
	if err != nil {
		return nil, err
	}
	return &model.BatchReceipt{
		SenderPublicKey:      senderPublicKey,
		RecipientPublicKey:   recipientPublicKey,
		DatumType:            datumType,
		DatumHash:            datumHash,
		ReferenceBlockHeight: referenceBlock.Height,
		ReferenceBlockHash:   refBlockHash,
		RMRLinked:            rmrLinked,
	}, nil
}

// GetUnsignedBatchReceiptBytes Client task while doing validation signature
func (ru *ReceiptUtil) GetUnsignedBatchReceiptBytes(
	receipt *model.BatchReceipt,
) []byte {

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

// GetSignedBatchReceiptBytes Client task before store into database batch_receipt
func (ru *ReceiptUtil) GetSignedBatchReceiptBytes(receipt *model.BatchReceipt) []byte {
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
