package transaction

import (
	"bytes"
	"errors"
	"time"
	"unicode/utf8"

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

	// Address format: [len][address]
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
	/***
	Escrow part
	1. ApproverAddress
	2. Commission
	3. Timeout
	*/
	buffer.Write(util.ConvertUint32ToBytes(uint32(len([]byte(transaction.GetEscrow().GetApproverAddress())))))
	if transaction.GetEscrow() != nil {
		buffer.Write([]byte(transaction.GetEscrow().GetApproverAddress()))
		buffer.Write(util.ConvertUint64ToBytes(uint64(transaction.GetEscrow().GetCommission())))
		buffer.Write(util.ConvertUint64ToBytes(transaction.GetEscrow().GetTimeout()))
	} else {
		buffer.Write(make([]byte, constant.AccountAddress))
		buffer.Write(make([]byte, constant.EscrowCommissionLength))
		buffer.Write(make([]byte, constant.EscrowTimeoutLength))
	}

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
	var (
		chunkedBytes []byte
		transaction  model.Transaction
		buffer       = bytes.NewBuffer(transactionBytes)
		escrow       model.Escrow
		err          error
	)

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionType))
	if err != nil {
		return nil, err
	}
	transaction.TransactionType = util.ConvertBytesToUint32(chunkedBytes)

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionVersion))
	if err != nil {
		return nil, err
	}
	transaction.Version = uint32(chunkedBytes[0])

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Timestamp))
	if err != nil {
		return nil, err
	}
	transaction.Timestamp = int64(util.ConvertBytesToUint64(chunkedBytes))

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	transaction.SenderAccountAddress = string(ReadAccountAddress(util.ConvertBytesToUint32(chunkedBytes), buffer))

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	transaction.RecipientAccountAddress = string(ReadAccountAddress(util.ConvertBytesToUint32(chunkedBytes), buffer))

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Fee))
	if err != nil {
		return nil, err
	}
	transaction.Fee = int64(util.ConvertBytesToUint64(chunkedBytes))

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionBodyLength))
	if err != nil {
		return nil, err
	}
	transaction.TransactionBodyLength = util.ConvertBytesToUint32(chunkedBytes)

	transaction.TransactionBodyBytes, err = util.ReadTransactionBytes(buffer, int(transaction.TransactionBodyLength))
	if err != nil {
		return nil, err
	}
	/***
	Escrow part
	1. ApproverAddress
	2. Commission
	3. Timeout
	*/
	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.AccountAddressLength))
	if err != nil {
		return nil, err
	}
	approvedAddress := ReadAccountAddress(util.ConvertBytesToUint32(chunkedBytes), buffer)
	if ok, _ := utf8.DecodeRune(approvedAddress); ok == 0 {
		escrow.ApproverAddress = ""
	} else {
		escrow.ApproverAddress = string(approvedAddress)
	}

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowCommissionLength))
	if err != nil {
		return nil, err
	}
	escrow.Commission = int64(util.ConvertBytesToUint64(chunkedBytes))

	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowTimeoutLength))
	if err != nil {
		return nil, err
	}
	escrow.Timeout = util.ConvertBytesToUint64(chunkedBytes)

	transaction.Escrow = &escrow

	if sign {
		// TODO: implement below logic to allow multiple signature algorithm to work
		// first 4 bytes of signature are the signature type
		// signatureLengthBytes, err := ReadTransactionBytes(buffer, 2)
		// if err != nil {
		// 	return nil, err
		// }
		// signatureLength := int(ConvertBytesToUint32(signatureLengthBytes))
		transaction.Signature, err = util.ReadTransactionBytes(buffer, int(constant.SignatureType+constant.AccountSignature))
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
	transaction.ID = txID
	transaction.TransactionHash = transactionHash[:]
	return &transaction, nil
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
	// There is additional time offset for the transaction timestamp before comparing with time now
	if time.Duration(tx.Timestamp)*time.Second-constant.TransactionTimeOffset > time.Duration(time.Now().UnixNano()) {
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
	res, err := accountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)

	if err != nil || len(res) == 0 {
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
