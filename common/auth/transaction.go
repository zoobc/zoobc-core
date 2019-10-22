package auth

import (
	"time"

	"github.com/zoobc/zoobc-core/common/blocker"

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

	unsignedTransactionBytes, err := util.GetTransactionBytes(tx, false)
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
