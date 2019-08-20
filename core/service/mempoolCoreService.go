package service

import (
	"bytes"
	"database/sql"
	"errors"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		GetMempoolTransactions() ([]*model.MempoolTransaction, error)
		GetMempoolTransaction(id int64) (*model.MempoolTransaction, error)
		AddMempoolTransaction(mpTx *model.MempoolTransaction) error
		SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error)
		ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error
		ReceivedTransactionListener() observer.Listener
	}

	// MempoolService contains all transactions in mempool plus a mux to manage locks in concurrency
	MempoolService struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Observer            *observer.Observer
	}
)

// NewMempoolService returns an instance of mempool service
func NewMempoolService(
	ct chaintype.ChainType,
	queryExecutor query.ExecutorInterface,
	mempoolQuery query.MempoolQueryInterface,
	actionTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	obsr *observer.Observer,
) *MempoolService {
	return &MempoolService{
		Chaintype:           ct,
		QueryExecutor:       queryExecutor,
		MempoolQuery:        mempoolQuery,
		ActionTypeSwitcher:  actionTypeSwitcher,
		AccountBalanceQuery: accountBalanceQuery,
		Observer:            obsr,
	}
}

// GetMempoolTransactions fetch transactions from mempool
func (mps *MempoolService) GetMempoolTransactions() ([]*model.MempoolTransaction, error) {
	var rows *sql.Rows
	var err error
	sqlStr := mps.MempoolQuery.GetMempoolTransactions()
	rows, err = mps.QueryExecutor.ExecuteSelect(sqlStr)
	if err != nil {
		log.Printf("GetMempoolTransactions fails %s\n", err)
		return nil, err
	}
	defer rows.Close()
	mempoolTransactions := []*model.MempoolTransaction{}
	mempoolTransactions = mps.MempoolQuery.BuildModel(mempoolTransactions, rows)
	return mempoolTransactions, nil
}

// GetMempoolTransaction return a mempool transaction by its ID
func (mps *MempoolService) GetMempoolTransaction(id int64) (*model.MempoolTransaction, error) {
	rows, err := mps.QueryExecutor.ExecuteSelect(mps.MempoolQuery.GetMempoolTransaction(), id)
	if err != nil {
		return &model.MempoolTransaction{
			ID: -1,
		}, err
	}
	defer rows.Close()
	var mpTx []*model.MempoolTransaction
	mpTx = mps.MempoolQuery.BuildModel(mpTx, rows)
	if len(mpTx) > 0 {
		return mpTx[0], nil
	}
	return &model.MempoolTransaction{
		ID: -1,
	}, errors.New("MempoolTransactionNotFound")
}

// AddMempoolTransaction validates and insert a transaction into the mempool
func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	// check if already in db
	_, err := mps.GetMempoolTransaction(mpTx.ID)
	if err == nil {
		return errors.New("DuplicateRecordAttempted")
	}
	if err.Error() != "MempoolTransactionNotFound" {
		log.Println(err)
		return errors.New("DatabaseError")
	}

	err = mps.QueryExecutor.ExecuteTransaction(mps.MempoolQuery.InsertMempoolTransaction(), mps.MempoolQuery.ExtractModel(mpTx)...)
	if err != nil {
		return err
	}
	// broadcast transaction
	mps.Observer.Notify(observer.TransactionAdded, mpTx.GetTransactionBytes(), nil)
	return nil
}

func (mps *MempoolService) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	tx, err := util.ParseTransactionBytes(mpTx.TransactionBytes, true)
	if err != nil {
		return err
	}
	if err := util.ValidateTransaction(tx, mps.QueryExecutor, mps.AccountBalanceQuery, true); err != nil {
		return err
	}

	if err := mps.ActionTypeSwitcher.GetTransactionType(tx).Validate(); err != nil {

		return err
	}
	return nil
}

// SelectTransactionsFromMempool Select transactions from mempool to be included in the block and return an ordered list.
// 1. get all mempool transaction from db (all mpTx already processed but still not included in a block)
// 2. merge with mempool, until it's full (payload <= MAX_PAYLOAD_LENGTH and max 255 mpTx) and do formal validation
//	  (timestamp <= MAX_TIMEDRIFT, mpTx is formally valid)
// 3. sort new mempool by arrival time then height then ID (this last one sounds useless to me unless ids are sortable..)
// Note: Tx Order is important to allow every node with a same set of transactions to  build the block and always obtain
//		 the same block hash.
// This function is equivalent of selectMempoolTransactions in NXT
func (mps *MempoolService) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error) {
	mempoolTransactions, err := mps.GetMempoolTransactions()
	if err != nil {
		return nil, err
	}

	var payloadLength int
	sortedTransactions := make([]*model.MempoolTransaction, 0)
	for payloadLength <= constant.MaxPayloadLength && len(mempoolTransactions) <= constant.MaxNumberOfTransactions {
		prevNumberOfNewTransactions := len(sortedTransactions)
		for _, mempoolTransaction := range mempoolTransactions {
			transactionLength := len(mempoolTransaction.TransactionBytes)
			if transactionsContain(sortedTransactions, mempoolTransaction) || payloadLength+transactionLength > constant.MaxPayloadLength {
				continue
			}

			tx, err := util.ParseTransactionBytes(mempoolTransaction.TransactionBytes, true)
			if err != nil {
				log.Println(err)
				continue
			}
			// compute transaction expiration time
			txExpirationTime := blockTimestamp + constant.TransactionExpirationOffset
			if blockTimestamp > 0 && tx.Timestamp > txExpirationTime {
				continue
			}

			if err = mps.ActionTypeSwitcher.GetTransactionType(tx).Validate(); err != nil {
				continue
			}

			sortedTransactions = append(sortedTransactions, mempoolTransaction)
			payloadLength += transactionLength
		}
		if len(sortedTransactions) == prevNumberOfNewTransactions {
			break
		}
	}
	sortFeePerByteThenTimestampThenID(sortedTransactions)
	return sortedTransactions, nil
}

func (mps *MempoolService) ReceivedTransactionListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionBytes interface{}, args interface{}) {
			var (
				err        error
				receivedTx *model.Transaction
				mempoolTx  *model.MempoolTransaction
			)

			receivedTxBytes := transactionBytes.([]byte)
			receivedTx, err = util.ParseTransactionBytes(receivedTxBytes, true)
			if err != nil {
				return
			}
			mempoolTx = &model.MempoolTransaction{
				// TODO: how to determine FeePerByte in mempool?
				FeePerByte:              0,
				ID:                      receivedTx.ID,
				TransactionBytes:        receivedTxBytes,
				ArrivalTimestamp:        time.Now().Unix(),
				SenderAccountAddress:    receivedTx.SenderAccountAddress,
				RecipientAccountAddress: receivedTx.RecipientAccountAddress,
			}

			// Validate received transaction
			if err = mps.ValidateMempoolTransaction(mempoolTx); err != nil {
				log.Warnf("Invalid received transaction submitted: %v", err)
				return
			}

			if err = mps.QueryExecutor.BeginTx(); err != nil {
				log.Warnf("error opening db transaction %v", err)
				return
			}
			// Apply Unconfirmed transaction
			err = mps.ActionTypeSwitcher.GetTransactionType(receivedTx).ApplyUnconfirmed()
			if err != nil {
				log.Warnf("fail ApplyUnconfirmed tx: %v\n", err)
				if err = mps.QueryExecutor.RollbackTx(); err != nil {
					log.Warnf("error rolling back db transaction %v", err)
					return
				}
				return
			}

			// Store to Mempool Transaction
			if err = mps.AddMempoolTransaction(mempoolTx); err != nil {
				log.Warnf("error AddMempoolTransaction: %v\n", err)
				if err = mps.QueryExecutor.RollbackTx(); err != nil {
					log.Warnf("error rolling back db transaction %v", err)
					return
				}
				return
			}

			if err = mps.QueryExecutor.CommitTx(); err != nil {
				log.Warnf("error committing db transaction: %v", err)
				return
			}
		},
	}
}

func transactionsContain(a []*model.MempoolTransaction, x *model.MempoolTransaction) bool {
	for _, n := range a {
		if bytes.Equal(x.TransactionBytes, n.TransactionBytes) {
			return true
		}
	}
	return false
}

// SortByTimestampThenHeightThenID sort a slice of mpTx by feePerByte, timestamp, id DESC
func sortFeePerByteThenTimestampThenID(members []*model.MempoolTransaction) {
	sort.SliceStable(members, func(i, j int) bool {
		mi, mj := members[i], members[j]
		switch {
		case mi.FeePerByte != mj.FeePerByte:
			return mi.FeePerByte > mj.FeePerByte
		case mi.ArrivalTimestamp != mj.ArrivalTimestamp:
			return mi.ArrivalTimestamp < mj.ArrivalTimestamp
		default:
			return mi.ID < mj.ID
		}
	})
}
