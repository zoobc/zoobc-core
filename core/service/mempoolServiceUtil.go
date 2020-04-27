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
		AddMempoolTransaction(mpTx *model.MempoolTransaction) error
	}

	MempoolServiceUtil struct {
		MempoolGetter          MempoolGetterInterface
		TransactionUtil        transaction.UtilInterface
		TransactionQuery       query.TransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
		MempoolQuery           query.MempoolQueryInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		BlockQuery             query.BlockQueryInterface
		TransactionCoreService TransactionCoreServiceInterface
	}
)

func NewMempoolServiceUtil(
	transactionUtil transaction.UtilInterface,
	transactionQuery query.TransactionQueryInterface,
	queryExecutor query.ExecutorInterface,
	mempoolQuery query.MempoolQueryInterface,
	actionTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	blockQuery query.BlockQueryInterface,
	mempoolGetter MempoolGetterInterface,
	transactionCoreService TransactionCoreServiceInterface,
) MempoolServiceUtilInterface {
	return &MempoolServiceUtil{
		TransactionUtil:        transactionUtil,
		TransactionQuery:       transactionQuery,
		QueryExecutor:          queryExecutor,
		MempoolQuery:           mempoolQuery,
		ActionTypeSwitcher:     actionTypeSwitcher,
		AccountBalanceQuery:    accountBalanceQuery,
		BlockQuery:             blockQuery,
		MempoolGetter:          mempoolGetter,
		TransactionCoreService: transactionCoreService,
	}
}

func (mpsu *MempoolServiceUtil) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	var (
		mempoolTx model.MempoolTransaction
		parsedTx  *model.Transaction
		tx        model.Transaction
		err       error
		row       *sql.Row
		txType    transaction.TypeAction
	)
	// check for duplication in transaction table
	transactionQ := mpsu.TransactionQuery.GetTransaction(mpTx.ID)
	row, err = mpsu.QueryExecutor.ExecuteSelectRow(transactionQ, false)
	if err != nil {
		return err
	}

	err = mpsu.TransactionQuery.Scan(&tx, row)
	if err != nil && err != sql.ErrNoRows {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	if mpTx.GetID() == tx.GetID() {
		return blocker.NewBlocker(blocker.ValidationErr, "MempoolDuplicated")
	}

	// check for duplication in mempool table
	mempoolQ := mpsu.MempoolQuery.GetMempoolTransaction()
	row, err = mpsu.QueryExecutor.ExecuteSelectRow(mempoolQ, false, mpTx.ID)
	if err != nil {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	err = mpsu.MempoolQuery.Scan(&mempoolTx, row)
	if err != nil {
		if err != sql.ErrNoRows {
			return blocker.NewBlocker(blocker.DBErr, err.Error())
		}
	}
	if mpTx.GetID() == mempoolTx.GetID() {
		return blocker.NewBlocker(blocker.ValidationErr, "MempoolDuplicated")
	}

	parsedTx, err = mpsu.TransactionUtil.ParseTransactionBytes(mpTx.TransactionBytes, true)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	if errVal := mpsu.TransactionUtil.ValidateTransaction(parsedTx, mpsu.QueryExecutor, mpsu.AccountBalanceQuery, true); errVal != nil {
		return blocker.NewBlocker(blocker.ValidationErr, errVal.Error())
	}
	txType, err = mpsu.ActionTypeSwitcher.GetTransactionType(parsedTx)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	err = mpsu.TransactionCoreService.ValidateTransaction(txType, false)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}
	return nil
}

func (mpsu *MempoolServiceUtil) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	var (
		err error
		row *sql.Row
	)
	// check maximum mempool
	if constant.MaxMempoolTransactions > 0 {
		count, err := mpsu.MempoolGetter.GetTotalMempoolTransactions()
		if err != nil {
			return err
		}
		if count >= constant.MaxMempoolTransactions {
			return blocker.NewBlocker(blocker.ValidationErr, "Mempool already full")
		}
	}

	// NOTE: this select is always inside a db transaction because AddMempoolTransaction is always called within a db tx
	row, err = mpsu.QueryExecutor.ExecuteSelectRow(mpsu.BlockQuery.GetLastBlock(), false)
	if err != nil {
		return err
	}
	var lastBlock model.Block
	err = mpsu.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "GetLastBlockFail")
	}

	mpTx.BlockHeight = lastBlock.GetHeight()
	insertMempoolQ, insertMempoolArgs := mpsu.MempoolQuery.InsertMempoolTransaction(mpTx)
	err = mpsu.QueryExecutor.ExecuteTransaction(insertMempoolQ, insertMempoolArgs...)
	if err != nil {
		return err
	}
	return nil
}

type (
	MempoolGetterInterface interface {
		GetMempoolTransactions() ([]*model.MempoolTransaction, error)
		GetMempoolTransaction(id int64) (*model.MempoolTransaction, error)
		GetTotalMempoolTransactions() (int, error)
	}

	MempoolGetter struct {
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
	}
)

func NewMempoolGetter(queryExecutor query.ExecutorInterface, mempoolQuery query.MempoolQueryInterface) MempoolGetterInterface {
	return &MempoolGetter{
		QueryExecutor: queryExecutor,
		MempoolQuery:  mempoolQuery,
	}
}

func (mg *MempoolGetter) GetTotalMempoolTransactions() (int, error) {
	var (
		count  int
		err    error
		row    *sql.Row
		sqlStr = mg.MempoolQuery.GetMempoolTransactions()
	)
	// note: this select is always inside a db transaction because AddMempoolTransaction is always called within a db tx
	row, err = mg.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(sqlStr), true)
	if err != nil {
		return count, err
	}
	err = row.Scan(&count)
	if err != nil {
		return count, err
	}
	return count, nil
}

// GetMempoolTransactions fetch transactions from mempool
func (mg *MempoolGetter) GetMempoolTransactions() ([]*model.MempoolTransaction, error) {
	var (
		mempoolTransactions []*model.MempoolTransaction
		sqlStr              = mg.MempoolQuery.GetMempoolTransactions()
		rows                *sql.Rows
		err                 error
	)

	rows, err = mg.QueryExecutor.ExecuteSelect(sqlStr, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mempoolTransactions, err = mg.MempoolQuery.BuildModel(mempoolTransactions, rows)
	if err != nil {
		return nil, err
	}

	return mempoolTransactions, nil
}

// GetMempoolTransaction return a mempool transaction by its ID
func (mg *MempoolGetter) GetMempoolTransaction(id int64) (*model.MempoolTransaction, error) {
	var (
		rows *sql.Rows
		mpTx []*model.MempoolTransaction
		err  error
	)

	rows, err = mg.QueryExecutor.ExecuteSelect(mg.MempoolQuery.GetMempoolTransaction(), false, id)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	mpTx, err = mg.MempoolQuery.BuildModel(mpTx, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if len(mpTx) > 0 {
		return mpTx[0], nil
	}

	return nil, blocker.NewBlocker(blocker.DBRowNotFound, "MempoolTransactionNotFound")
}
