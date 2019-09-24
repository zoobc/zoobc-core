package util

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/model"
)

// GenerateReceipt generate receipt object that act as proof of receipt on data. Data received can be
// block, transaction, etc.
// generated receipt will not be signed yet (RecipientSignature = nil), will need to be signed using SignReceipt method.
func GenerateBatchReceipt(
	referenceBlock *model.Block,
	senderPublicKey, recipientPublicKey, datumHash []byte,
	datumType uint32,
) (*model.BatchReceipt, error) {
	refBlockHash, _ := GetBlockHash(referenceBlock)
	return &model.BatchReceipt{
		SenderPublicKey:      senderPublicKey,
		RecipientPublicKey:   recipientPublicKey,
		DatumType:            datumType,
		DatumHash:            datumHash,
		ReferenceBlockHeight: referenceBlock.Height,
		ReferenceBlockHash:   refBlockHash,
	}, nil
}

// GetUnsignedReceiptBytes Client task while doing validation signature
func GetUnsignedBatchReceiptBytes(
	receipt *model.BatchReceipt,
) []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(receipt.SenderPublicKey)
	buffer.Write(receipt.RecipientPublicKey)
	buffer.Write(ConvertUint32ToBytes(receipt.ReferenceBlockHeight))
	buffer.Write(receipt.ReferenceBlockHash)
	buffer.Write(ConvertUint32ToBytes(receipt.DatumType))
	buffer.Write(receipt.DatumHash)
	buffer.Write(receipt.ReceiptMerkleRoot)
	return buffer.Bytes()
}

// GetSignedReceiptBytes Client task before store into database batch_receipt
func GetSignedBatchReceiptBytes(receipt *model.BatchReceipt) []byte {

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(receipt.SenderPublicKey)
	buffer.Write(receipt.RecipientPublicKey)
	buffer.Write(ConvertUint32ToBytes(receipt.ReferenceBlockHeight))
	buffer.Write(receipt.ReferenceBlockHash)
	buffer.Write(ConvertUint32ToBytes(receipt.DatumType))
	buffer.Write(receipt.DatumHash)
	buffer.Write(receipt.ReceiptMerkleRoot)
	buffer.Write(receipt.RecipientSignature)
	return buffer.Bytes()
}
