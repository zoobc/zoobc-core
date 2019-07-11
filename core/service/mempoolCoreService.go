package service

import (
	"bytes"
	"database/sql"
	"errors"
	"math"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/util"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		InitMempool() error
		GetMempoolTransactions() ([]*model.MempoolTransaction, error)
		GetMempoolTransaction(id []byte) (*model.MempoolTransaction, error)
		AddMempoolTransaction(mpTx *model.MempoolTransaction) error
		SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error)
		ValidateMempoolTrasnsaction(mpTx *model.MempoolTransaction) error
	}

	// MempoolService contains all transactions in mempool plus a mux to manage locks in concurrency
	MempoolService struct {
		MempoolMutex  *sync.Mutex
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
	rows, err = mps.QueryExecutor.ExecuteSelect(query.NewMempoolQuery(mps.Chaintype).GetMempoolTransactions())
	defer rows.Close()
	if err != nil {
		log.Printf("GetMempoolTransactions fails %v\n", err)
		return nil, err
	}

	mempoolTransactions := []*model.MempoolTransaction{}
	for rows.Next() {
		var bl model.MempoolTransaction
		err = rows.Scan(
			&bl.ID,
			&bl.FeePerByte,
			&bl.ArrivalTimestamp,
			&bl.TransactionBytes,
		)
		if err != nil {
			log.Printf("GetMempoolTransactions fails scan %v\n", err)
			return nil, err
		}
		mempoolTransactions = append(mempoolTransactions, &bl)
	}

	return mempoolTransactions, nil

}

// GetMempoolTransaction return a mempool transaction by its ID
func (mps *MempoolService) GetMempoolTransaction(id []byte) (*model.MempoolTransaction, error) {
	rows, err := mps.QueryExecutor.ExecuteSelect(mps.MempoolQuery.GetMempoolTransaction(id))
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		return &model.MempoolTransaction{
			ID: make([]byte, 0),
		}, err
	}
	var mpTx model.MempoolTransaction
	if rows.Next() {
		err = rows.Scan(&mpTx.ID, &mpTx.ArrivalTimestamp, &mpTx.FeePerByte, &mpTx.TransactionBytes)
		if err != nil {
			return &model.MempoolTransaction{
				ID: make([]byte, 0),
			}, err
		}
		return &mpTx, nil
	}
	return &model.MempoolTransaction{
		ID: make([]byte, 0),
	}, errors.New("MempoolTransactionNotFound")
}

// AddMempoolTransaction validates and insert a transaction into the mempool
func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	// check if already in db
	_, err := mps.GetMempoolTransaction(mpTx.ID)
	if err == nil {
		return errors.New("DuplicateRecordAttempted")
	}
	//TODO: validate the transaction
	_, err = util.ParseTransactionBytes(mpTx.TransactionBytes, true)
	if err != nil {
		return err
	}
	// 	mpTx.GetMempoolTransaction().Validate()

	result, err := mps.QueryExecutor.ExecuteStatement(mps.MempoolQuery.InsertMempoolTransaction(), mps.MempoolQuery.ExtractModel(mpTx)...)
	if err != nil {
		return err
	}
	log.Printf("got new mempool transaction, %v", result)
	return nil
}

// RemoveMempoolTransaction removes a tx from mempool
func (mps *MempoolService) RemoveMempoolTransaction(id []byte) error {
	_, err := mps.QueryExecutor.ExecuteStatement(mps.MempoolQuery.DeleteMempoolTransaction(id))
	if err != nil {
		return err
	}
	log.Printf("mempool transaction with ID = %s deleted", id)
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
			//TODO: compute transaction expiration date
			txExpirationTime := int64(math.MaxInt64) // dummy data
			if blockTimestamp > 0 && txExpirationTime < blockTimestamp {
				continue
			}
			if err := mps.ValidateMempoolTrasnsaction(mempoolTransaction); err != nil {
				log.Println(err)
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

func (mps *MempoolService) ValidateMempoolTrasnsaction(mpTx *model.MempoolTransaction) error {
	return nil
}

func transactionsContain(a []*model.MempoolTransaction, x *model.MempoolTransaction) bool {
	for _, n := range a {
		if bytes.Equal(x.TransactionBytes, n.TransactionBytes) {
			return true
		}
	}
	return false
}

// SortByTimestampThenHeightThenID sort a slice of mpTx by timestamp, height, id DESC
func sortFeePerByteThenTimestampThenID(members []*model.MempoolTransaction) {
	sort.SliceStable(members, func(i, j int) bool {
		mi, mj := members[i], members[j]
		switch {
		case mi.FeePerByte != mj.FeePerByte:
			return mi.FeePerByte < mj.FeePerByte
		case mi.ArrivalTimestamp != mj.ArrivalTimestamp:
			return mi.ArrivalTimestamp < mj.ArrivalTimestamp
		default:
			miID := commonUtil.ConvertBytesToUint64(mi.ID)
			mjID := commonUtil.ConvertBytesToUint64(mj.ID)
			return miID < mjID
		}
	})
}
