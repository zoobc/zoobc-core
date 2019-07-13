package util

import (
	"bytes"
	"errors"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

// GetTransactionBytes translate transaction model to its byte representation
// provide sign = true to translate transaction with its signature, sign = false
// for without signature (used for verify signature)
func GetTransactionBytes(transaction *model.Transaction, sign bool) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(transaction.TransactionType)[:2])
	buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.Timestamp)))
	buffer.Write(util.ConvertUint32ToBytes(transaction.SenderAccountType))
	buffer.Write(util.ConvertStringToBytes(transaction.SenderAccountAddress))
	buffer.Write(util.ConvertUint32ToBytes(transaction.RecipientAccountType))
	buffer.Write(util.ConvertStringToBytes(transaction.RecipientAccountAddress))
	buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.Fee)))
	// transaction body length
	buffer.Write(util.ConvertUint32ToBytes(transaction.TransactionBodyLength))
	buffer.Write(transaction.TransactionBodyBytes)
	if sign {
		if transaction.Signature == nil {
			return nil, errors.New("TransactionSignatureNotExist")
		}
		buffer.Write(transaction.Signature)
	}
	return buffer.Bytes(), nil
}

// ParseTransactionBytes build transaction from transaction bytes
func ParseTransactionBytes(transactionBytes []byte, sign bool) (*model.Transaction, error) {
	buffer := bytes.NewBuffer(transactionBytes)
	transactionTypeBytes := buffer.Next(2)
	transactionType := util.ConvertBytesToUint32([]byte{transactionTypeBytes[0], transactionTypeBytes[1], 0, 0})
	timestamp := util.ConvertBytesToUint64(buffer.Next(8))
	senderAccountType := util.ConvertBytesToUint32(buffer.Next(4))

	senderAccountAddressLength := util.ConvertBytesToUint32(buffer.Next(4))
	senderAccountAddress := bytes.NewBuffer(buffer.Next(int(senderAccountAddressLength))).String()

	recipientAccountType := util.ConvertBytesToUint32(buffer.Next(4))

	recipientAccountAddressLength := util.ConvertBytesToUint32(buffer.Next(4))
	recipientAccountAddress := bytes.NewBuffer(buffer.Next(int(recipientAccountAddressLength))).String()

	fee := util.ConvertBytesToUint64(buffer.Next(8))
	transactionBodyLength := util.ConvertBytesToUint32(buffer.Next(4))
	transactionBodyBytes := buffer.Next(int(transactionBodyLength))
	var signature []byte
	if sign {
		signature = buffer.Next(64)
		if len(signature) < 64 { // signature is not there
			return nil, errors.New("TransactionSignatureNotExist")
		}
	}
	return &model.Transaction{
		TransactionType:         transactionType,
		Timestamp:               int64(timestamp),
		SenderAccountType:       senderAccountType,
		SenderAccountAddress:    senderAccountAddress,
		RecipientAccountType:    recipientAccountType,
		RecipientAccountAddress: recipientAccountAddress,
		Fee:                     int64(fee),
		TransactionBodyLength:   transactionBodyLength,
		TransactionBodyBytes:    transactionBodyBytes,
		Signature:               signature,
	}, nil
}
