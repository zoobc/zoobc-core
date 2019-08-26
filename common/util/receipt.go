package util

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/model"
)

// GenerateReceipt generate receipt object that act as proof of receipt on data. Data received can be
// block, transaction, etc.
// generated receipt will not be signed yet (RecipientSignature = nil), will need to be signed using SignReceipt method.
// todo: andy-shi88: receipt merkle root value is not assigned yet
func GenerateReceipt(
	referenceBlock *model.Block,
	senderAddress, recipientAddress string,
	datumHash []byte,
	datumType uint32,
) (*model.Receipt, error) {
	refBlockHash, _ := GetBlockHash(referenceBlock)
	return &model.Receipt{
		SenderAccountAddress:    senderAddress,
		RecipientAccountAddress: recipientAddress,
		DatumType:               datumType,
		DatumHash:               datumHash,
		ReferenceBlockHeight:    referenceBlock.Height,
		ReferenceBlockHash:      refBlockHash,
		ReceiptMerkleRoot:       nil,
	}, nil
}

func GetUnsignedReceiptBytes(
	receipt *model.Receipt,
) []byte {
	buffer := bytes.NewBuffer([]byte{})
	recipientPublicKey, _ := GetPublicKeyFromAddress(receipt.RecipientAccountAddress)
	senderPublicKey, _ := GetPublicKeyFromAddress(receipt.SenderAccountAddress)
	buffer.Write(senderPublicKey)
	buffer.Write(recipientPublicKey)
	buffer.Write(ConvertUint32ToBytes(receipt.ReferenceBlockHeight))
	buffer.Write(receipt.ReferenceBlockHash)
	buffer.Write(ConvertUint32ToBytes(receipt.DatumType))
	buffer.Write(receipt.DatumHash)
	buffer.Write(receipt.ReceiptMerkleRoot)
	return buffer.Bytes()
}
