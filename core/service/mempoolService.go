package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		InitMempool() error
		GetTransactions(ctNum int32) ([]*model.MempoolTransaction, error)
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
func NewMempoolService(chaintype contract.ChainType, queryExecutor query.ExecutorInterface,
	mempoolQuery query.MempoolQueryInterface) *MempoolService {
	return &MempoolService{
		Chaintype:     chaintype,
		QueryExecutor: queryExecutor,
		MempoolQuery:  mempoolQuery,
	}
}

// GetTransactions fetch transactions from mempool
func (mps *MempoolService) GetTransactions(ctNum int32) ([]*model.MempoolTransaction, error) {
	cType := chaintype.GetChainType(ctNum)
	var rows *sql.Rows
	var err error
	rows, err = mps.QueryExecutor.ExecuteSelect(query.NewMempoolQuery(cType).GetMempoolTransactions())
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

// ProcessPeerTransaction reference: processPeerTransactions()
// func ProcessPeerTransaction(chaintype contract.ChainType, mpTx *model.Mempool) error {
// 	// iterate the transactions

// 	// check if the tranasction is already in mempool transaction

// 	// validate the transaction
// 	mpTx.GetTransaction().Validate()

// 	// create mempool transaction out of the received transaction

// 	// process transaction

// 	// add the mempool transaction
// 	AddMempoolTransaction(chaintype, mpTx)

// 	// notify the listener that there is a new mempool transactions received

// 	return nil
// }

// AddMempoolTransaction add a transaction to mempool
func AddMempoolTransaction(ctNum int32, mpTx model.MempoolTransaction) error {
	tcJSON, _ := json.MarshalIndent(mpTx, " ", "  ")
	fmt.Printf("AddMempoolTransaction %s\n", tcJSON)

	//FIXME:
	// validationError := mpTx.GetTransaction().Validate()
	// if validationError != nil {
	// 	fmt.Printf("AddMempoolTransaction failure: %v", validationError)
	// 	return validationError
	// }
	// // save to mempool mpTx table
	// err := MempoolTransactionRepository(chaintype).Save(mpTx)
	// if err != nil {
	// 	return err
	// }
	// mempool.Transactions = append(mempool.Transactions, mpTx)
	// fmt.Printf("mempool length %d \n", len(mempool.Transactions))
	return nil
}

// // RemoveTransactionById
// func RemoveTransactionById(chaintype contract.ChainType, id int64) {
// 	mempool := GetMempool(chaintype)
// 	for i, utx := range mempool.GetTransactions() {
// 		mpTx := utx.GetTransaction()
// 		txID, _ := mpTx.GetID(chaintype)
// 		if txID == id {
// 			mempool.Transactions[len(mempool.Transactions)-1], mempool.Transactions[i] = mempool.Transactions[i], mempool.Transactions[len(mempool.Transactions)-1]
// 			mempool.Transactions = mempool.Transactions[:len(mempool.Transactions)-1]
// 			break
// 		}
// 	}
// 	MempoolTransactionRepository(chaintype).Delete(id, nil)
// }

// // TODO: delete this function [temporary use only]
// func PopAllTransaction(chaintype contract.ChainType) []model.Mempool {
// 	mempool := GetMempool(chaintype)
// 	return mempool.Transactions[:]
// }

// // SelectTransactionsFromMempool Select transactions from mempool to be included in the block and return an ordered list.
// // 1. get all unconfirmet mpTx from db (all mpTx already processed but still not included in a block)
// // 2. filter out the ones that have referenced mpTx not confirmed yet (implements basic logic for chained transactions)
// // 3. merge with mempool, untill it's full (payload <= MAX_PAYLOAD_LENGTH and max 255 mpTx) and do formal validation (timestamp <= MAX_TIMEDRIFT, mpTx is formally valid)
// // 4. sort new mempool by arrival time then height then ID (this last one sounds useless to me unless ids are sortable..)
// // Note: Tx Order is important to allow every node with a same set of transactions to  build the block and always obtain the same block hash.
// // This function is equivalent of selectMempoolTransactions in NXT
// func SelectTransactionsFromMempool(chaintype contract.ChainType, blockTimestamp uint32, utxRepo contracts.MempoolTransactionRepository, txRepo contracts.TransactionRepository) []model.Mempool {
// 	// unconfirmedTransactions are all mpTx in db still to be put in (note this method implements an interface signature, so can be mocked in unit tests)
// 	//STEF unconfirmedDbTransactions, err := utxRepo.FindAllMempoolTransactions(new(TransactionUtil))
// 	// if err != nil {
// 	// 	log.Fatal("Error finding mempool transactions")
// 	// }
// 	//merge mempool mpTx from db with mpTx already in mempool and remove duplicates
// 	//TODO: this shouldn't be necessary, since mempool mpTx form db and mempool should always be in sync
// 	//STEF mempoolTx := GetMempoolTransactions(chaintype)
// 	//STEF newMempoolTransactions := uniqueTransactions(append(mempoolTx, unconfirmedDbTransactions...), chaintype)
// 	newMempoolTransactions := GetMempoolTransactions(chaintype)
// 	// TODO: delete this if we don't use referenced transactions
// 	// note: instead of removing elements from the slice we create a new slice with just the elements that we want to keep
// 	// tmp := newMempoolTransactions[:0]
// 	// for _, mpTx := range newMempoolTransactions {
// 	// 	if mpTx.GetTransaction().HasAllReferencedTransactions(mpTx.GetTransaction().GetTimestamp(), 0, txRepo, chaintype) {
// 	// 		tmp = append(tmp, mpTx)
// 	// 	}
// 	// }
// 	// newMempoolTransactions = tmp
// 	var payloadLength int
// 	sortedTransactions := make([]model.Mempool, 0)
// 	for payloadLength <= constant.MAX_PAYLOAD_LENGTH && len(newMempoolTransactions) <= constant.MAX_NUMBER_OF_TRANSACTIONS {
// 		prevNumberOfNewTransactions := len(sortedTransactions)
// 		for _, newMempoolTransaction := range newMempoolTransactions {
// 			transactionLength := newMempoolTransaction.GetTransaction().GetSize()
// 			if transactionsContain(sortedTransactions, newMempoolTransaction, chaintype) || payloadLength+transactionLength > constant.MAX_PAYLOAD_LENGTH {
// 				continue
// 			}
// 			// txTimestamp := newMempoolTransaction.GetTransaction().GetTimestamp()
// 			txExpirationTime := newMempoolTransaction.GetTransaction().GetExpiration()
// 			// if blockTimestamp > 0 && txExpirationTime < blockTimestamp {
// 			// this condition leads to throw away many new transactions..
// 			// log.Printf("\ntx ts: %v\ntx ex: %v\nbl ts: %v\nbl ts+ %v", txTimestamp, txExpirationTime, blockTimestamp, blockTimestamp+constant.MAX_TIMEDRIFT)
// 			// if blockTimestamp > 0 && txTimestamp > blockTimestamp+constant.MAX_TIMEDRIFT {
// 			// 	continue
// 			// }
// 			if blockTimestamp > 0 && txExpirationTime < blockTimestamp {
// 				continue
// 			}
// 			if newMempoolTransaction.GetTransaction().Validate() != nil {
// 				continue
// 			}

// 			sortedTransactions = append(sortedTransactions, newMempoolTransaction)
// 			payloadLength += transactionLength
// 		}
// 		if len(sortedTransactions) == prevNumberOfNewTransactions {
// 			break
// 		}
// 	}
// 	sortByTimestampThenHeightThenID(sortedTransactions, chaintype)
// 	return sortedTransactions
// }

// func IsTransactionAlreadyExist(chaintype contract.ChainType, mpTx model.Mempool) bool {
// 	mempool := GetMempool(chaintype)
// 	return transactionsContain(mempool.Transactions, mpTx, chaintype)
// }

// func transactionsContain(a []model.Mempool, x model.Mempool, chaintype contract.ChainType) bool {
// 	for _, n := range a {
// 		xID, _ := x.GetTransaction().GetID(chaintype)
// 		nID, _ := x.GetTransaction().GetID(chaintype)
// 		if bytes.Equal(x.GetTransaction().GetSignature(), n.GetTransaction().GetSignature()) && xID == nID {
// 			return true
// 		}
// 	}
// 	return false
// }

// func uniqueTransactions(transactions []model.Mempool, chaintype contract.ChainType) []model.Mempool {
// 	keys := make(map[int64]model.Mempool)
// 	list := []model.Mempool{}
// 	for _, transaction := range transactions {
// 		txID, _ := transaction.GetTransaction().GetID(chaintype)
// 		if _, value := keys[txID]; !value {
// 			keys[txID] = transaction
// 			list = append(list, transaction)
// 		}
// 	}
// 	return list
// }

// // SortByTimestampThenHeightThenID sort a slice of mpTx by timestamp, height, id DESC
// func sortByTimestampThenHeightThenID(members []model.Mempool, chaintype contract.ChainType) {
// 	sort.SliceStable(members, func(i, j int) bool {
// 		mi, mj := members[i].GetTransaction(), members[j].GetTransaction()
// 		switch {
// 		case mi.GetTimestamp() != mj.GetTimestamp():
// 			return mi.GetTimestamp() < mj.GetTimestamp()
// 		case mi.GetHeight() != mj.GetHeight():
// 			return mi.GetHeight() < mj.GetHeight()
// 		default:
// 			miID, _ := mi.GetID(chaintype)
// 			mjID, _ := mj.GetID(chaintype)
// 			return miID < mjID
// 		}
// 	})
// }
