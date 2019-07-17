package service

import (
	"bytes"
	"database/sql"
	"errors"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		GetMempoolTransactions() ([]*model.MempoolTransaction, error)
		GetMempoolTransaction(id int64) (*model.MempoolTransaction, error)
		AddMempoolTransaction(mpTx *model.MempoolTransaction) error
		SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error)
	}

	// MempoolService contains all transactions in mempool plus a mux to manage locks in concurrency
	MempoolService struct {
		Chaintype     contract.ChainType
		QueryExecutor query.ExecutorInterface
		MempoolQuery  query.MempoolQueryInterface
	}
)

// NewMempoolService returns an instance of mempool service
func NewMempoolService(ct contract.ChainType, queryExecutor query.ExecutorInterface,
	mempoolQuery query.MempoolQueryInterface) *MempoolService {
	return &MempoolService{
		Chaintype:     ct,
		QueryExecutor: queryExecutor,
		MempoolQuery:  mempoolQuery,
	}
}

// GetMempoolTransactions fetch transactions from mempool
func (mps *MempoolService) GetMempoolTransactions() ([]*model.MempoolTransaction, error) {
	var rows *sql.Rows
	var err error
	sqlStr := query.NewMempoolQuery(mps.Chaintype).GetMempoolTransactions()
	rows, err = mps.QueryExecutor.ExecuteSelect(sqlStr)
	if err != nil {
		log.Printf("GetMempoolTransactions fails %s\n", err)
		return nil, err
	}
	defer rows.Close()
	mempoolTransactions := []*model.MempoolTransaction{}
	for rows.Next() {
		var mpTx model.MempoolTransaction
		err = rows.Scan(
			&mpTx.ID,
			&mpTx.FeePerByte,
			&mpTx.ArrivalTimestamp,
			&mpTx.TransactionBytes,
		)
		if err != nil {
			log.Printf("GetMempoolTransactions fails scan %v\n", err)
			return nil, err
		}
		mempoolTransactions = append(mempoolTransactions, &mpTx)
	}
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
	var mpTx model.MempoolTransaction
	if rows.Next() {
		err = rows.Scan(&mpTx.ID, &mpTx.ArrivalTimestamp, &mpTx.FeePerByte, &mpTx.TransactionBytes)
		if err != nil {
			return &model.MempoolTransaction{
				ID: -1,
			}, err
		}
		return &mpTx, nil
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

	if err := mps.ValidateMempoolTransaction(mpTx); err != nil {
		return err
	}

	result, err := mps.QueryExecutor.ExecuteStatement(mps.MempoolQuery.InsertMempoolTransaction(), mps.MempoolQuery.ExtractModel(mpTx)...)
	if err != nil {
		return err
	}
	log.Printf("got new mempool transaction, %v", result)
	return nil
}

func (mps *MempoolService) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	tx, err := util.ParseTransactionBytes(mpTx.TransactionBytes, true)
	if err != nil {
		return err
	}

	// formally validate tx fields
	if len(tx.TransactionHash) == 0 {
		return errors.New("InvalidTransactionHash")
	}

	txID, err := util.GetTransactionID(tx.TransactionHash)
	if err != nil {
		return err
	}

	// verify that transaction ID sent by client = transaction ID calculated from transaction bytes (TransactionHash)
	if tx.ID != txID {
		return errors.New("InvalidTransactionID")
	}

	if err := transaction.GetTransactionType(tx).Validate(); err != nil {
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
			txExpirationTime := tx.Timestamp + constant.TransactionExpirationOffset
			if blockTimestamp == 0 || txExpirationTime > blockTimestamp {
				continue
			}

			//TODO: we could remove this, since tx has already been validated when added to the mempool,
			//		but what if someone manually adds it to db?..
			if err := transaction.GetTransactionType(tx).Validate(); err != nil {
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
