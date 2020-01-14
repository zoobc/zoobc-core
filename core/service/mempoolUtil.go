package service

import (
	"database/sql"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
)

type (
	MempoolServiceUtilInterface interface {
		ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error
	}

	MempoolServiceUtil struct {
	}
)

func (mps *MempoolServiceUtil) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	var (
		tx        model.Transaction
		mempoolTx model.MempoolTransaction
		parsedTx  *model.Transaction
		err       error
	)
	// check for duplication in transaction table
	transactionQ := mps.TransactionQuery.GetTransaction(mpTx.ID)
	row, _ := mps.QueryExecutor.ExecuteSelectRow(transactionQ, false)
	err = mps.TransactionQuery.Scan(&tx, row)
	if err != nil && err != sql.ErrNoRows {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	if mpTx.GetID() == tx.GetID() {
		return blocker.NewBlocker(blocker.ValidationErr, "MempoolDuplicated")
	}

	// check for duplication in mempool table
	mempoolQ := mps.MempoolQuery.GetMempoolTransaction()
	row, _ = mps.QueryExecutor.ExecuteSelectRow(mempoolQ, false, mpTx.ID)
	err = mps.MempoolQuery.Scan(&mempoolTx, row)

	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
	}
	if mpTx.GetID() == mempoolTx.GetID() {
		return blocker.NewBlocker(blocker.ValidationErr, "MempoolDuplicated")
	}

	parsedTx, err = transaction.ParseTransactionBytes(mpTx.TransactionBytes, true)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	if err := transaction.ValidateTransaction(parsedTx, mps.QueryExecutor, mps.AccountBalanceQuery, true); err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}
	txType, err := mps.ActionTypeSwitcher.GetTransactionType(parsedTx)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	err = txType.Validate(false)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	return nil
}

func (mps *MempoolServiceUtil) ValidateMaxMempoolReached() error {
	// check maximum mempool
	if constant.MaxMempoolTransactions > 0 {
		var count int
		sqlStr := mps.MempoolQuery.GetMempoolTransactions()
		// note: this select is always insid a db transaction because AddMempoolTransaction is always called within a db tx
		row, err := mps.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(sqlStr), true)
		if err != nil {
			return err
		}
		err = row.Scan(&count)
		if err != nil {
			return err
		}
		if count >= constant.MaxMempoolTransactions {
			return blocker.NewBlocker(blocker.ValidationErr, "Mempool already full")
		}
	}
	return nil
}

func (mps *MempoolServiceUtil) ValidateDuplicateMempool(mempoolID int64) error {
	// check if already in db
	mempool, err := mps.GetMempoolTransaction(mpTx.ID)
	if err != nil {
		if blockErr, ok := err.(blocker.Blocker); ok && blockErr.Type != blocker.DBRowNotFound {
			return blocker.NewBlocker(blocker.ValidationErr, blockErr.Message)
		}
	}
	if mempool != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "DuplicatedRecordAttempted")
	}
	return nil
}
