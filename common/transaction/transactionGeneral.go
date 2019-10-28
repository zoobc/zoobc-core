package transaction

import (
	"bytes"
	"errors"
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

// GetTransactionBytes translate transaction model to its byte representation
// provide sign = true to translate transaction with its signature, sign = false
// for without signature (used for verify signature)
func GetTransactionBytes(transaction *model.Transaction, sign bool) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(transaction.TransactionType))
	buffer.Write(util.ConvertUint32ToBytes(transaction.Version)[:constant.TransactionVersion])
	buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.Timestamp)))
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(transaction.SenderAccountAddress)))))
	buffer.Write([]byte(transaction.SenderAccountAddress))
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(transaction.RecipientAccountAddress)))))
	if transaction.RecipientAccountAddress == "" {
		buffer.Write(make([]byte, constant.AccountAddress)) // if no recipient pad with 44 (zoobc address length)
	} else {
		buffer.Write([]byte(transaction.RecipientAccountAddress))
	}
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

	transactionTypeBytes, err := util.ReadTransactionBytes(buffer, int(constant.TransactionType))
	if err != nil {
		return nil, err
	}
	transactionType := util.ConvertBytesToUint32(transactionTypeBytes)
	transactionVersionByte, err := util.ReadTransactionBytes(buffer, int(constant.TransactionVersion))
	if err != nil {
		return nil, err
	}
	transactionVersion := uint32(transactionVersionByte[0])
	timestampBytes, err := util.ReadTransactionBytes(buffer, int(constant.Timestamp))
	if err != nil {
		return nil, err
	}
	timestamp := util.ConvertBytesToUint64(timestampBytes)
	senderAccountAddressLength, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	senderAccountAddress := ReadAccountAddress(util.ConvertBytesToUint32(senderAccountAddressLength), buffer)
	recipientAccountAddressLength, err := util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	recipientAccountAddress := ReadAccountAddress(util.ConvertBytesToUint32(recipientAccountAddressLength), buffer)
	feeBytes, err := util.ReadTransactionBytes(buffer, int(constant.Fee))
	if err != nil {
		return nil, err
	}
	fee := util.ConvertBytesToUint64(feeBytes)
	transactionBodyLengthBytes, err := util.ReadTransactionBytes(buffer, int(constant.TransactionBodyLength))
	if err != nil {
		return nil, err
	}
	transactionBodyLength := util.ConvertBytesToUint32(transactionBodyLengthBytes)
	transactionBodyBytes, err := util.ReadTransactionBytes(buffer, int(transactionBodyLength))
	if err != nil {
		return nil, err
	}
	var sig []byte
	if sign {
		var err error
		//TODO: implement below logic to allow multiple signature algorithm to work
		// first 4 bytes of signature are the signature type
		// signatureLengthBytes, err := ReadTransactionBytes(buffer, 2)
		// if err != nil {
		// 	return nil, err
		// }
		// signatureLength := int(ConvertBytesToUint32(signatureLengthBytes))
		sig, err = util.ReadTransactionBytes(buffer, int(constant.SignatureType+constant.AccountSignature))
		if err != nil {
			return nil, blocker.NewBlocker(
				blocker.ParserErr,
				"no transaction signature",
			)
		}
	}
	// compute and return tx hash and ID too
	transactionHash := sha3.Sum256(transactionBytes)
	txID, _ := GetTransactionID(transactionHash[:])
	tx := &model.Transaction{
		ID:                      txID,
		TransactionType:         transactionType,
		Version:                 transactionVersion,
		Timestamp:               int64(timestamp),
		SenderAccountAddress:    string(senderAccountAddress),
		RecipientAccountAddress: string(recipientAccountAddress),
		Fee:                     int64(fee),
		TransactionBodyLength:   transactionBodyLength,
		TransactionBodyBytes:    transactionBodyBytes,
		TransactionHash:         transactionHash[:],
		Signature:               sig,
	}
	return tx, nil
}

// ReadAccountAddress to read the sender or recipient address from transaction bytes
// depend on their account types.
func ReadAccountAddress(accountType uint32, transactionBuffer *bytes.Buffer) []byte {
	switch accountType {
	case 0:
		return transactionBuffer.Next(int(constant.AccountAddress)) // zoobc account address length
	default:
		return transactionBuffer.Next(int(constant.AccountAddress)) // default to zoobc account address
	}
}

// GetTransactionID calculate and returns a transaction ID given a transaction model
func GetTransactionID(transactionHash []byte) (int64, error) {
	if len(transactionHash) == 0 {
		return -1, errors.New("InvalidTransactionHash")
	}
	ID := int64(util.ConvertBytesToUint64(transactionHash))
	return ID, nil
}

// ValidateTransaction take in transaction object and execute basic validation
func ValidateTransaction(
	tx *model.Transaction,
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	verifySignature bool,
) error {
	if tx.Fee <= 0 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxFeeZero",
		)
	}
	if tx.SenderAccountAddress == "" {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxSenderEmpty",
		)
	}
	// check if transaction is coming from future / comparison in second
	if tx.Timestamp > time.Now().UnixNano()/1e9 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxComeFromFuture",
		)
	}

	// validate sender account
	qry, args := accountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAccountAddress)
	rows, err := queryExecutor.ExecuteSelect(qry, false, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	res := accountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
	if len(res) == 0 {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxSenderNotFound",
		)
	}
	senderAccountBalance := res[0]
	if senderAccountBalance.SpendableBalance < tx.Fee {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxAccountBalanceNotEnough",
		)
	}

	// formally validate transaction body
	if len(tx.TransactionBodyBytes) != int(tx.TransactionBodyLength) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"TxInvalidBodyFormat",
		)
	}

	unsignedTransactionBytes, err := GetTransactionBytes(tx, false)
	if err != nil {
		return err
	}
	// verify the signature of the transaction
	if verifySignature {
		if !crypto.NewSignature().VerifySignature(unsignedTransactionBytes, tx.Signature, tx.SenderAccountAddress) {
			return blocker.NewBlocker(
				blocker.ValidationErr,
				"TxInvalidSignature",
			)
		}
	}

	return nil
}
