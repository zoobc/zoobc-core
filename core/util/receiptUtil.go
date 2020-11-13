package util

import (
	"bytes"

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
