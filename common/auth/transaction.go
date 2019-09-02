package auth

import (
	"errors"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

// ValidateTransaction take in transaction object and execute basic validation
func ValidateTransaction(
	tx *model.Transaction,
	queryExecutor query.ExecutorInterface,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	verifySignature bool,
) error {
	if tx.Fee <= 0 {
		return errors.New("TxFeeZero")
	}
	if tx.SenderAccountAddress == "" {
		return errors.New("TxSenderEmpty")
	}

	// validate sender account
	sqlQ, arg := accountBalanceQuery.GetAccountBalanceByAccountAddress(tx.SenderAccountAddress)
	rows, err := queryExecutor.ExecuteSelect(sqlQ, arg)
	if err != nil {
		return err
	}
	defer rows.Close()
	res := accountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
	if len(res) == 0 {
		return errors.New("TxSenderNotFound")
	}
	senderAccountBalance := res[0]
	if senderAccountBalance.SpendableBalance < tx.Fee {
		return errors.New("TxAccountBalanceNotEnough")
	}

	// formally validate transaction body
	if len(tx.TransactionBodyBytes) != int(tx.TransactionBodyLength) {
		return errors.New("TxInvalidBodyFormat")
	}

	unsignedTransactionBytes, err := util.GetTransactionBytes(tx, false)
	if err != nil {
		return err
	}
	// verify the signature of the transaction
	if verifySignature {
		if !crypto.NewSignature().VerifySignature(unsignedTransactionBytes, tx.Signature, tx.SenderAccountAddress) {
			return errors.New("TxInvalidSignature")
		}
	}

	return nil
}
