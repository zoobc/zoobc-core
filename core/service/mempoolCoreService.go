package service

import (
	"database/sql"
	"sort"
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
	}
)

// NewMempoolService returns an instance of mempool service
func NewMempoolService(
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
	return &MempoolService{
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
	}
}

// GetMempoolTransactions fetch transactions from mempool
func (mps *MempoolService) GetMempoolTransactions() ([]*model.MempoolTransaction, error) {
	var (
		rows *sql.Rows
		err  error
	)
	sqlStr := mps.MempoolQuery.GetMempoolTransactions()
	rows, err = mps.QueryExecutor.ExecuteSelect(sqlStr, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mempoolTransactions []*model.MempoolTransaction
	mempoolTransactions, err = mps.MempoolQuery.BuildModel(mempoolTransactions, rows)
	if err != nil {
		return nil, err
	}
	return mempoolTransactions, nil
}

// GetMempoolTransaction return a mempool transaction by its ID
func (mps *MempoolService) GetMempoolTransaction(id int64) (*model.MempoolTransaction, error) {
	var (
		rows *sql.Rows
		mpTx []*model.MempoolTransaction
		err  error
	)

	rows, err = mps.QueryExecutor.ExecuteSelect(mps.MempoolQuery.GetMempoolTransaction(), false, id)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	defer rows.Close()

	mpTx, err = mps.MempoolQuery.BuildModel(mpTx, rows)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	if len(mpTx) > 0 {
		return mpTx[0], nil
	}

	return nil, blocker.NewBlocker(blocker.DBRowNotFound, "MempoolTransactionNotFound")
}

// AddMempoolTransaction validates and insert a transaction into the mempool and also set the BlockHeight as well
func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {

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

	// note: this select is always insid a db transaction because AddMempoolTransaction is always called within a db tx
	row, err := mps.QueryExecutor.ExecuteSelectRow(mps.BlockQuery.GetLastBlock(), true)
	if err != nil {
		return err
	}
	var lastBlock model.Block
	err = mps.BlockQuery.Scan(&lastBlock, row)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "GetLastBlockFail")
	}

	mpTx.BlockHeight = lastBlock.GetHeight()
	insertMempoolQ, insertMempoolArgs := mps.MempoolQuery.InsertMempoolTransaction(mpTx)
	err = mps.QueryExecutor.ExecuteTransaction(insertMempoolQ, insertMempoolArgs...)
	if err != nil {
		return err
	}
	return nil
}

func (mps *MempoolService) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
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

// SelectTransactionsFromMempool Select transactions from mempool to be included in the block and return an ordered list.
// 1. get all mempool transaction from db (all mpTx already processed but still not included in a block)
// 2. merge with mempool, until it's full (payload <= MAX_PAYLOAD_LENGTH and max 255 mpTx) and do formal validation
//	  (timestamp <= MAX_TIMEDRIFT, mpTx is formally valid)
// 3. sort new mempool by fee per byte, arrival timestamp then ID (this last one sounds useless to me unless ids are sortable..)
// Note: Tx Order is important to allow every node with a same set of transactions to  build the block and always obtain
//		 the same block hash.
// This function is equivalent of selectMempoolTransactions in NXT
func (mps *MempoolService) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.Transaction, error) {
	mempoolTransactions, err := mps.GetMempoolTransactions()
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

		tx, err := transaction.ParseTransactionBytes(mempoolTransaction.TransactionBytes, true)
		if err != nil {
			continue
		}
		// compute transaction expiration time
		txExpirationTime := tx.Timestamp + constant.TransactionExpirationOffset
		// compare to millisecond representation of block timestamp
		if blockTimestamp == 0 || blockTimestamp > txExpirationTime {
			continue
		}

		if err := transaction.ValidateTransaction(tx, mps.QueryExecutor, mps.AccountBalanceQuery, true); err != nil {
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
	receivedTx, err = transaction.ParseTransactionBytes(receivedTxBytes, true)
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
	// Validate received transaction
	if err = mps.ValidateMempoolTransaction(mempoolTx); err != nil {
		specificErr := err.(blocker.Blocker)
		if specificErr.Type == blocker.DuplicateMempoolErr {
			// already exist in mempool, check if already generated a receipt for this sender
			_, err := mps.KVExecutor.Get(constant.KVdbTableTransactionReminderKey + string(receiptKey))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
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
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := mps.QueryExecutor.BeginTx(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Apply Unconfirmed transaction
	txType, err := mps.ActionTypeSwitcher.GetTransactionType(receivedTx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = txType.ApplyUnconfirmed()
	if err != nil {
		mps.Logger.Infof("fail ApplyUnconfirmed tx: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Store to Mempool Transaction
	if err = mps.AddMempoolTransaction(mempoolTx); err != nil {
		mps.Logger.Infof("error AddMempoolTransaction: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = mps.QueryExecutor.CommitTx(); err != nil {
		mps.Logger.Warnf("error committing db transaction: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	// broadcast transaction
	mps.Observer.Notify(observer.TransactionAdded, mempoolTx.GetTransactionBytes(), mps.Chaintype)
	batchReceipt, err := coreUtil.GenerateBatchReceiptWithReminder(
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
		tx, err := transaction.ParseTransactionBytes(m.GetTransactionBytes(), true)
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
