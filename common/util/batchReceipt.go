package util

import (
	"bytes"

	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

// GenerateReceipt generate receipt object that act as proof of receipt on data. Data received can be
// block, transaction, etc.
// generated receipt will not be signed yet (RecipientSignature = nil), will need to be signed using SignReceipt method.
func GenerateBatchReceipt(
	ct chaintype.ChainType,
	referenceBlock *model.Block,
	senderPublicKey, recipientPublicKey, datumHash, rmrLinked []byte,
	datumType uint32,
) (*model.BatchReceipt, error) {
	refBlockHash, err := GetBlockHash(referenceBlock, ct)
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
	buffer.Write(receipt.RMRLinked)
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
	buffer.Write(receipt.RMRLinked)
	buffer.Write(receipt.RecipientSignature)
	return buffer.Bytes()
}

func GetReceiptKey(
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
