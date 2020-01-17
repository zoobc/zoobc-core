package service

import (
	"database/sql"
	"sort"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		CleanTimedoutBlockTxCached()
		DeleteBlockTxCached(txIds []int64, needAddToMempool bool)
		GetBlockTxCached(txID int64) *model.Transaction
		GetMempoolTransactions() ([]*model.MempoolTransaction, error)
		GetMempoolTransaction(id int64) (*model.MempoolTransaction, error)
		AddMempoolTransaction(mpTx *model.MempoolTransaction) error
		SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.Transaction, error)
		ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error
		ReceivedTransaction(
			senderPublicKey, receivedTxBytes []byte,
			lastBlock *model.Block,
			nodeSecretPhrase string,
		) (*model.BatchReceipt, error)
		DeleteExpiredMempoolTransactions() error
		GetMempoolTransactionsWantToBackup(height uint32) ([]*model.MempoolTransaction, error)
	}

	// MempoolService contains all transactions in mempool plus a mux to manage locks in concurrency
	MempoolService struct {
		MempoolServiceUtil  MempoolServiceUtilInterface
		MempoolGetter       MempoolGetterInterface
		TransactionUtil     transaction.UtilInterface
		Chaintype           chaintype.ChainType
		KVExecutor          kvdb.KVExecutorInterface
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		MerkleTreeQuery     query.MerkleTreeQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		BlockQuery          query.BlockQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		Signature           crypto.SignatureInterface
		Observer            *observer.Observer
		Logger              *log.Logger
		BlockTxCached       map[int64]*MempoolTxWithMetaData
		BlockTxCachedMutex  sync.Mutex
	}

	MempoolTxWithMetaData struct {
		MempoolTx   *model.MempoolTransaction
		Transaction *model.Transaction
		Timestamp   int64
	}
)

// NewMempoolService returns an instance of mempool service
func NewMempoolService(
	transactionUtil transaction.UtilInterface,
	ct chaintype.ChainType,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	mempoolQuery query.MempoolQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	actionTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	blockQuery query.BlockQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	signature crypto.SignatureInterface,
	observer *observer.Observer,
	logger *log.Logger,
) *MempoolService {
	mempoolGetter := &MempoolGetter{
		MempoolQuery:  mempoolQuery,
		QueryExecutor: queryExecutor,
	}

	return &MempoolService{
		MempoolServiceUtil: NewMempoolServiceUtil(transactionUtil,
			transactionQuery,
			queryExecutor,
			mempoolQuery,
			actionTypeSwitcher,
			accountBalanceQuery,
			blockQuery,
			mempoolGetter,
		),
		MempoolGetter:       mempoolGetter,
		TransactionUtil:     transactionUtil,
		Chaintype:           ct,
		KVExecutor:          kvExecutor,
		QueryExecutor:       queryExecutor,
		MempoolQuery:        mempoolQuery,
		MerkleTreeQuery:     merkleTreeQuery,
		ActionTypeSwitcher:  actionTypeSwitcher,
		AccountBalanceQuery: accountBalanceQuery,
		Signature:           signature,
		TransactionQuery:    transactionQuery,
		Observer:            observer,
		Logger:              logger,
		BlockQuery:          blockQuery,
		BlockTxCached:       make(map[int64]*MempoolTxWithMetaData),
	}
}

// CleanTimedoutBlockTxCached deletes timed out tx candidate that are needed by received block
func (mps *MempoolService) CleanTimedoutBlockTxCached() {
	mps.BlockTxCachedMutex.Lock()
	defer mps.BlockTxCachedMutex.Unlock()

	for txID, txWithMetaData := range mps.BlockTxCached {
		if txWithMetaData.Timestamp >= time.Now().Unix()-constant.TxCachedTimeout {
			delete(mps.BlockTxCached, txID)
		}
	}
}

// DeleteBlockTxCached deletes transactions candidate cached for blocks
func (mps *MempoolService) DeleteBlockTxCached(txIds []int64, needAddToMempool bool) {
	mps.BlockTxCachedMutex.Lock()
	defer mps.BlockTxCachedMutex.Unlock()

	for _, txID := range txIds {
		if needAddToMempool {
			err := mps.ProcessTransactionBytesToMempool(mps.BlockTxCached[txID].MempoolTx)
			mps.Logger.Errorln(err)
		}
		delete(mps.BlockTxCached, txID)
	}
}

// GetBlockTxCached gets transactions that are requested by the received blocks
func (mps *MempoolService) GetBlockTxCached(txID int64) *model.Transaction {
	mps.BlockTxCachedMutex.Lock()
	defer mps.BlockTxCachedMutex.Unlock()

	if mps.BlockTxCached[txID] == nil {
		return nil
	}
	return mps.BlockTxCached[txID].Transaction
}

// GetMempoolTransactions fetch transactions from mempool
func (mps *MempoolService) GetMempoolTransactions() ([]*model.MempoolTransaction, error) {
	return mps.MempoolGetter.GetMempoolTransactions()
}

// GetMempoolTransaction return a mempool transaction by its ID
func (mps *MempoolService) GetMempoolTransaction(id int64) (*model.MempoolTransaction, error) {
	return mps.MempoolGetter.GetMempoolTransaction(id)
}

// AddMempoolTransaction validates and insert a transaction into the mempool and also set the BlockHeight as well
func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return mps.MempoolServiceUtil.AddMempoolTransaction(mpTx)
}

func (mps *MempoolService) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return mps.MempoolServiceUtil.ValidateMempoolTransaction(mpTx)
}

// SelectTransactionsFromMempool Select transactions from mempool to be included in the block and return an ordered list.
// 1. get all mempool transaction from db (all mpTx already processed but still not included in a block)
// 2. merge with mempool, until it's full (payload <= MAX_PAYLOAD_LENGTH and max 255 mpTx) and do formal validation
//	  (timestamp <= MAX_TIMEDRIFT, mpTx is formally valid)
// 3. sort new mempool by fee per byte, arrival timestamp then ID (this last one sounds useless to me unless ids are sortable..)
// Note: Tx Order is important to allow every node with a same set of transactions to  build the block and always obtain
//		 the same block hash.
// This function is equivalent of selectMempoolTransactions in NXT
func (mps *MempoolService) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.Transaction, error) {
	mempoolTransactions, err := mps.MempoolGetter.GetMempoolTransactions()
	if err != nil {
		return nil, err
	}

	var payloadLength int
	selectedTransactions := make([]*model.Transaction, 0)
	selectedMempool := make([]*model.MempoolTransaction, 0)
	for _, mempoolTransaction := range mempoolTransactions {
		if len(selectedTransactions) >= constant.MaxNumberOfTransactionsInBlock {
			break
		}
		transactionLength := len(mempoolTransaction.TransactionBytes)
		if payloadLength+transactionLength > constant.MaxPayloadLengthInBlock {
			continue
		}

		tx, err := mps.TransactionUtil.ParseTransactionBytes(mempoolTransaction.TransactionBytes, true)
		if err != nil {
			continue
		}
		// compute transaction expiration time
		txExpirationTime := tx.Timestamp + constant.TransactionExpirationOffset
		// compare to millisecond representation of block timestamp
		if blockTimestamp == 0 || blockTimestamp > txExpirationTime {
			continue
		}

		if err := mps.TransactionUtil.ValidateTransaction(tx, mps.QueryExecutor, mps.AccountBalanceQuery, true); err != nil {
			continue
		}

		txType, err := mps.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			return nil, err
		}
		toRemove, err := txType.SkipMempoolTransaction(selectedTransactions)
		if err != nil {
			return nil, err
		}
		if toRemove {
			continue
		}

		selectedTransactions = append(selectedTransactions, tx)
		selectedMempool = append(selectedMempool, mempoolTransaction)
		payloadLength += transactionLength
	}
	sortFeePerByteThenTimestampThenID(selectedTransactions, selectedMempool)
	return selectedTransactions, nil
}

func (mps *MempoolService) ReceivedTransaction(
	senderPublicKey,
	receivedTxBytes []byte,
	lastBlock *model.Block,
	nodeSecretPhrase string,
) (*model.BatchReceipt, error) {
	var (
		err        error
		receivedTx *model.Transaction
		mempoolTx  *model.MempoolTransaction
	)
	receivedTx, err = mps.TransactionUtil.ParseTransactionBytes(receivedTxBytes, true)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	mempoolTx = &model.MempoolTransaction{
		FeePerByte:              util.FeePerByteTransaction(receivedTx.GetFee(), receivedTxBytes),
		ID:                      receivedTx.ID,
		TransactionBytes:        receivedTxBytes,
		ArrivalTimestamp:        time.Now().Unix(),
		SenderAccountAddress:    receivedTx.SenderAccountAddress,
		RecipientAccountAddress: receivedTx.RecipientAccountAddress,
	}
	receivedTxHash := sha3.Sum256(receivedTxBytes)
	receiptKey, err := util.GetReceiptKey(
		receivedTxHash[:], senderPublicKey,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if transactionCached := mps.GetBlockTxCached(receivedTx.ID); transactionCached == nil {
		// Validate received transaction
		if err = mps.MempoolServiceUtil.ValidateMempoolTransaction(mempoolTx); err != nil {
			specificErr := err.(blocker.Blocker)
			if specificErr.Type != blocker.DuplicateMempoolErr {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}

			// already exist in mempool, check if already generated a receipt for this sender
			_, err := mps.KVExecutor.Get(constant.KVdbTableTransactionReminderKey + string(receiptKey))
			if err != nil && err != badger.ErrKeyNotFound {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			mps.BlockTxCachedMutex.Lock()
			mps.BlockTxCached[receivedTx.ID] = &MempoolTxWithMetaData{
				MempoolTx:   mempoolTx,
				Transaction: receivedTx,
				Timestamp:   time.Now().Unix(),
			}
			mps.BlockTxCachedMutex.Unlock()
			mps.Observer.Notify(observer.ReceivedTransactionValidated, receivedTx, mps.Chaintype)

			// broadcast transaction
			mps.Observer.Notify(observer.TransactionAdded, mempoolTx.GetTransactionBytes(), mps.Chaintype)
		}
	}

	batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
		mps.Chaintype,
		receivedTxHash[:],
		lastBlock,
		senderPublicKey,
		nodeSecretPhrase,
		constant.KVdbTableTransactionReminderKey+string(receiptKey),
		constant.ReceiptDatumTypeTransaction,
		mps.Signature,
		mps.QueryExecutor,
		mps.KVExecutor,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return batchReceipt, nil
}

// ProcessTransactionBytesToMempool processing mempoolTx to be added to mempool
func (mps *MempoolService) ProcessTransactionBytesToMempool(mempoolTx *model.MempoolTransaction) error {
	var (
		err        error
		receivedTx *model.Transaction
	)
	if err = mps.MempoolServiceUtil.ValidateMempoolTransaction(mempoolTx); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if err := mps.QueryExecutor.BeginTx(); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	// Apply Unconfirmed transaction
	receivedTx, err = mps.TransactionUtil.ParseTransactionBytes(mempoolTx.TransactionBytes, true)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	txType, err := mps.ActionTypeSwitcher.GetTransactionType(receivedTx)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = txType.ApplyUnconfirmed()
	if err != nil {
		mps.Logger.Infof("fail ApplyUnconfirmed tx: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return status.Error(codes.InvalidArgument, err.Error())
	}
	// Store to Mempool Transaction
	if err = mps.MempoolServiceUtil.AddMempoolTransaction(mempoolTx); err != nil {
		mps.Logger.Infof("error AddMempoolTransaction: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err = mps.QueryExecutor.CommitTx(); err != nil {
		mps.Logger.Warnf("error committing db transaction: %v", err)
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

// sortFeePerByteThenTimestampThenID sort a slice of mpTx by feePerByte, timestamp, id DESC
// this sort the transaction by the mempool fields, mean both slice should have the same number of elements, and same
// order for this to work
func sortFeePerByteThenTimestampThenID(members []*model.Transaction, mempools []*model.MempoolTransaction) {
	sort.SliceStable(members, func(i, j int) bool {
		mi, mj := mempools[i], mempools[j]
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

// PruneMempoolTransactions handle fresh clean the mempool
// which is the mempool transaction has been hit expiration time
func (mps *MempoolService) DeleteExpiredMempoolTransactions() error {
	var (
		expirationTime = time.Now().Add(-constant.MempoolExpiration).Unix()
		selectQ, qStr  string
		err            error
		mempools       []*model.MempoolTransaction
	)

	selectQ = mps.MempoolQuery.GetExpiredMempoolTransactions(expirationTime)
	rows, err := mps.QueryExecutor.ExecuteSelect(selectQ, false)
	if err != nil {
		return err
	}
	defer rows.Close()

	mempools, err = mps.MempoolQuery.BuildModel(mempools, rows)
	if err != nil {
		return err
	}

	err = mps.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	for _, m := range mempools {
		tx, err := mps.TransactionUtil.ParseTransactionBytes(m.GetTransactionBytes(), true)
		if err != nil {
			if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
				mps.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		action, err := mps.ActionTypeSwitcher.GetTransactionType(tx)
		if err != nil {
			if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
				mps.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		err = action.UndoApplyUnconfirmed()
		if err != nil {
			if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
				mps.Logger.Error(rollbackErr.Error())
			}
			return err
		}
	}

	qStr = mps.MempoolQuery.DeleteExpiredMempoolTransactions(expirationTime)
	err = mps.QueryExecutor.ExecuteTransaction(qStr)
	if err != nil {
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	err = mps.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}

func (mps *MempoolService) GetMempoolTransactionsWantToBackup(height uint32) ([]*model.MempoolTransaction, error) {
	var (
		mempools []*model.MempoolTransaction
		rows     *sql.Rows
		err      error
	)

	rows, err = mps.QueryExecutor.ExecuteSelect(mps.MempoolQuery.GetMempoolTransactionsWantToByHeight(height), false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mempools, err = mps.MempoolQuery.BuildModel(mempools, rows)
	if err != nil {
		return nil, err
	}

	return mempools, nil
}
