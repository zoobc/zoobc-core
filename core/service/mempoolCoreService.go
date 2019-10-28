package service

import (
	"database/sql"
	"errors"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		GetMempoolTransactions() ([]*model.MempoolTransaction, error)
		GetMempoolTransaction(id int64) (*model.MempoolTransaction, error)
		AddMempoolTransaction(mpTx *model.MempoolTransaction) error
		SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error)
		ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error
		ReceivedTransaction(
			senderPublicKey, receivedTxBytes []byte,
			lastBlock *model.Block,
			nodeSecretPhrase string,
		) (*model.BatchReceipt, error)
		DeleteExpiredMempoolTransactions() error
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
		Signature           crypto.SignatureInterface
		TransactionQuery    query.TransactionQueryInterface
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
	signature crypto.SignatureInterface,
	transactionQuery query.TransactionQueryInterface,
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
	rows, err := mps.QueryExecutor.ExecuteSelect(mps.MempoolQuery.GetMempoolTransaction(), false, id)
	if err != nil {
		return &model.MempoolTransaction{
			ID: -1,
		}, err
	}
	defer rows.Close()

	var mpTx []*model.MempoolTransaction
	mpTx, err = mps.MempoolQuery.BuildModel(mpTx, rows)
	if err != nil {
		return nil, err
	}
	if len(mpTx) > 0 {
		return mpTx[0], nil
	}
	return &model.MempoolTransaction{
		ID: -1,
	}, errors.New("MempoolTransactionNotFound")
}

// AddMempoolTransaction validates and insert a transaction into the mempool
func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	// check maximum mempool
	if constant.MaxMempoolTransactions > 0 {
		var count int
		sqlStr := mps.MempoolQuery.GetMempoolTransactions()
		err := mps.QueryExecutor.ExecuteSelectRow(query.GetTotalRecordOfSelect(sqlStr)).Scan(&count)
		if err != nil {
			return err
		}
		if count >= constant.MaxMempoolTransactions {
			return blocker.NewBlocker(blocker.ValidationErr, "Mempool already full")
		}
	}

	// check if already in db
	_, err := mps.GetMempoolTransaction(mpTx.ID)
	if err == nil {
		return errors.New("DuplicateRecordAttempted")
	}
	if err.Error() != "MempoolTransactionNotFound" {
		return errors.New("DatabaseError")
	}

	err = mps.QueryExecutor.ExecuteTransaction(mps.MempoolQuery.InsertMempoolTransaction(), mps.MempoolQuery.ExtractModel(mpTx)...)
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
	row := mps.QueryExecutor.ExecuteSelectRow(transactionQ)
	err = mps.TransactionQuery.Scan(&tx, row)
	if tx.ID != 0 {
		return blocker.NewBlocker(
			blocker.DuplicateTransactionErr,
			"mempool validation: duplicate transaction",
		)
	}
	if err != sql.ErrNoRows {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}
	// check for duplication in mempool table
	mempoolQ := mps.MempoolQuery.GetMempoolTransaction()
	row = mps.QueryExecutor.ExecuteSelectRow(mempoolQ, mpTx.ID)
	err = mps.MempoolQuery.Scan(&mempoolTx, row)
	if mempoolTx.ID != 0 {
		return blocker.NewBlocker(
			blocker.DuplicateMempoolErr,
			"mempool validation: duplicate mempool",
		)
	}
	if err != sql.ErrNoRows {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	parsedTx, err = util.ParseTransactionBytes(mpTx.TransactionBytes, true)
	if err != nil {
		return err
	}

	if err := auth.ValidateTransaction(parsedTx, mps.QueryExecutor, mps.AccountBalanceQuery, true); err != nil {
		return err
	}
	txType, err := mps.ActionTypeSwitcher.GetTransactionType(parsedTx)
	if err != nil {
		return err
	}
	err = txType.Validate(false)
	if err != nil {
		return err
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
func (mps *MempoolService) SelectTransactionsFromMempool(blockTimestamp int64) ([]*model.MempoolTransaction, error) {
	mempoolTransactions, err := mps.GetMempoolTransactions()
	if err != nil {
		return nil, err
	}

	var payloadLength int
	sortedTransactions := make([]*model.MempoolTransaction, 0)
	for _, mempoolTransaction := range mempoolTransactions {
		if len(sortedTransactions) >= constant.MaxNumberOfTransactionsInBlock {
			break
		}
		transactionLength := len(mempoolTransaction.TransactionBytes)
		if payloadLength+transactionLength > constant.MaxPayloadLengthInBlock {
			continue
		}

		tx, err := util.ParseTransactionBytes(mempoolTransaction.TransactionBytes, true)
		if err != nil {
			continue
		}
		// compute transaction expiration time
		txExpirationTime := tx.Timestamp + constant.TransactionExpirationOffset
		// compare to millisecond representation of block timestamp
		if blockTimestamp == 0 || blockTimestamp > txExpirationTime {
			continue
		}

		parsedTx, err := util.ParseTransactionBytes(mempoolTransaction.TransactionBytes, true)
		if err != nil {
			continue
		}

		if err := auth.ValidateTransaction(parsedTx, mps.QueryExecutor, mps.AccountBalanceQuery, true); err != nil {
			continue
		}

		sortedTransactions = append(sortedTransactions, mempoolTransaction)
		payloadLength += transactionLength
	}
	sortFeePerByteThenTimestampThenID(sortedTransactions)
	return sortedTransactions, nil
}

func (mps *MempoolService) generateTransactionReceipt(
	receivedTxHash []byte,
	lastBlock *model.Block,
	senderPublicKey, receiptKey []byte,
	nodeSecretPhrase string,
) (*model.BatchReceipt, error) {
	var rmrLinked []byte
	nodePublicKey := util.GetPublicKeyFromSeed(nodeSecretPhrase)
	lastRmrQ := mps.MerkleTreeQuery.GetLastMerkleRoot()
	row := mps.QueryExecutor.ExecuteSelectRow(lastRmrQ)
	rmrLinked, err := mps.MerkleTreeQuery.ScanRoot(row)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	// generate receipt
	batchReceipt, err := util.GenerateBatchReceipt( // todo: var
		lastBlock,
		senderPublicKey,
		nodePublicKey,
		receivedTxHash,
		rmrLinked,
		constant.ReceiptDatumTypeTransaction,
	)
	if err != nil {
		return nil, err
	}
	batchReceipt.RecipientSignature = mps.Signature.SignByNode(
		util.GetUnsignedBatchReceiptBytes(batchReceipt),
		nodeSecretPhrase,
	)
	// store the generated batch receipt hash
	err = mps.KVExecutor.Insert(string(receiptKey), receivedTxHash, constant.ExpiryPublishedReceiptTransaction)
	if err != nil {
		return nil, err
	}
	return batchReceipt, nil
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
	receivedTx, err = util.ParseTransactionBytes(receivedTxBytes, true)
	if err != nil {
		return nil, err
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
		return nil, blocker.NewBlocker(
			blocker.AppErr,
			err.Error(),
		)
	}
	// Validate received transaction
	if err = mps.ValidateMempoolTransaction(mempoolTx); err != nil {
		specificErr := err.(blocker.Blocker)
		if specificErr.Type == blocker.DuplicateMempoolErr {
			// already exist in mempool, check if already generated a receipt for this sender
			batchReceipt, err := mps.generateTransactionReceipt(
				receivedTxHash[:], lastBlock, senderPublicKey, receiptKey, nodeSecretPhrase,
			)
			if err != nil {
				return nil, err
			}
			return batchReceipt, nil
		}
		return nil, err
	}

	if err := mps.QueryExecutor.BeginTx(); err != nil {
		return nil, err
	}
	// Apply Unconfirmed transaction
	txType, err := mps.ActionTypeSwitcher.GetTransactionType(receivedTx)
	if err != nil {
		return nil, err
	}
	err = txType.ApplyUnconfirmed()
	if err != nil {
		mps.Logger.Infof("fail ApplyUnconfirmed tx: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return nil, err
	}

	// Store to Mempool Transaction
	if err = mps.AddMempoolTransaction(mempoolTx); err != nil {
		mps.Logger.Infof("error AddMempoolTransaction: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return nil, err
	}

	if err = mps.QueryExecutor.CommitTx(); err != nil {
		mps.Logger.Warnf("error committing db transaction: %v", err)
		return nil, err
	}
	// broadcast transaction
	mps.Observer.Notify(observer.TransactionAdded, mempoolTx.GetTransactionBytes(), mps.Chaintype)
	batchReceipt, err := mps.generateTransactionReceipt(
		receivedTxHash[:], lastBlock, senderPublicKey, receiptKey, nodeSecretPhrase,
	)
	if err != nil {
		return nil, err
	}
	return batchReceipt, nil
}

// sortFeePerByteThenTimestampThenID sort a slice of mpTx by feePerByte, timestamp, id DESC
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
		tx, err := util.ParseTransactionBytes(m.GetTransactionBytes(), true)
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
