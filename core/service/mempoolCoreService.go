package service

import (
	"database/sql"
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	commonUtils "github.com/zoobc/zoobc-core/common/util"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// MempoolServiceInterface represents interface for MempoolService
	MempoolServiceInterface interface {
		InitMempoolTransaction() error
		AddMempoolTransaction(tx *model.Transaction, txBytes []byte) error
		RemoveMempoolTransactions(mempoolTxs []*model.Transaction) error
		GetMempoolTransactions() (storage.MempoolMap, error)
		GetTotalMempoolTransactions() (int, error)
		SelectTransactionsFromMempool(blockTimestamp int64, blockHeight uint32) ([]*model.Transaction, error)
		ValidateMempoolTransaction(mpTx *model.Transaction) error
		ReceivedTransaction(
			senderPublicKey, receivedTxBytes []byte,
			lastBlock *model.Block,
			nodeSecretPhrase string,
		) (*model.BatchReceipt, error)
		ReceivedBlockTransactions(
			senderPublicKey []byte,
			receivedTxBytes [][]byte,
			lastBlock *model.Block,
			nodeSecretPhrase string,
		) ([]*model.BatchReceipt, error)
		DeleteExpiredMempoolTransactions() error
		GetMempoolTransactionsWantToBackup(height uint32) ([]*model.Transaction, error)
		BackupMempools(commonBlock *model.Block) error
	}

	// MempoolService contains all transactions in mempool plus a mux to manage locks in concurrency
	MempoolService struct {
		TransactionUtil        transaction.UtilInterface
		Chaintype              chaintype.ChainType
		QueryExecutor          query.ExecutorInterface
		MempoolQuery           query.MempoolQueryInterface
		MerkleTreeQuery        query.MerkleTreeQueryInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		TransactionQuery       query.TransactionQueryInterface
		Signature              crypto.SignatureInterface
		Observer               *observer.Observer
		Logger                 *log.Logger
		ReceiptUtil            coreUtil.ReceiptUtilInterface
		ReceiptService         ReceiptServiceInterface
		TransactionCoreService TransactionCoreServiceInterface
		BlockStateStorage      storage.CacheStorageInterface
		MempoolCacheStorage    storage.CacheStorageInterface
		MempoolBackupStorage   storage.CacheStorageInterface
	}
)

// NewMempoolService returns an instance of mempool service
func NewMempoolService(
	transactionUtil transaction.UtilInterface,
	ct chaintype.ChainType,
	queryExecutor query.ExecutorInterface,
	mempoolQuery query.MempoolQueryInterface,
	merkleTreeQuery query.MerkleTreeQueryInterface,
	actionTypeSwitcher transaction.TypeActionSwitcher,
	accountBalanceQuery query.AccountBalanceQueryInterface,
	transactionQuery query.TransactionQueryInterface,
	signature crypto.SignatureInterface,
	observer *observer.Observer,
	logger *log.Logger,
	receiptUtil coreUtil.ReceiptUtilInterface,
	receiptService ReceiptServiceInterface,
	transactionCoreService TransactionCoreServiceInterface,
	blockStateStorage, mempoolCacheStorage, mempoolBackupStorage storage.CacheStorageInterface,
) *MempoolService {
	return &MempoolService{
		TransactionUtil:        transactionUtil,
		Chaintype:              ct,
		QueryExecutor:          queryExecutor,
		MempoolQuery:           mempoolQuery,
		MerkleTreeQuery:        merkleTreeQuery,
		ActionTypeSwitcher:     actionTypeSwitcher,
		AccountBalanceQuery:    accountBalanceQuery,
		Signature:              signature,
		TransactionQuery:       transactionQuery,
		Observer:               observer,
		Logger:                 logger,
		ReceiptUtil:            receiptUtil,
		ReceiptService:         receiptService,
		TransactionCoreService: transactionCoreService,
		BlockStateStorage:      blockStateStorage,
		MempoolCacheStorage:    mempoolCacheStorage,
		MempoolBackupStorage:   mempoolBackupStorage,
	}
}

func (mps *MempoolService) InitMempoolTransaction() error {
	var (
		err      error
		mempools []*model.MempoolTransaction
	)
	// clearing cache before initialize
	err = mps.MempoolCacheStorage.ClearCache()
	if err != nil {
		return err
	}
	mpQuery := mps.MempoolQuery.GetMempoolTransactions()
	rows, err := mps.QueryExecutor.ExecuteSelect(mpQuery, false)
	if err != nil {
		return err
	}
	defer rows.Close()
	mempools, err = mps.MempoolQuery.BuildModel(mempools, rows)
	if err != nil {
		return err
	}
	for _, mempool := range mempools {
		tx, err := mps.TransactionUtil.ParseTransactionBytes(mempool.TransactionBytes, true)
		if err != nil {
			return err
		}
		err = mps.MempoolCacheStorage.SetItem(mempool.ID, storage.MempoolCacheObject{
			Tx:                  *tx,
			ArrivalTimestamp:    mempool.ArrivalTimestamp,
			FeePerByte:          mempool.FeePerByte,
			TransactionByteSize: uint32(len(mempool.TransactionBytes)),
			BlockHeight:         mempool.BlockHeight,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveMempoolTransactions removes a list of transactions tx from mempool given their Ids
func (mps *MempoolService) RemoveMempoolTransactions(transactions []*model.Transaction) error {
	var (
		idsStr []string
		ids    []int64
	)
	for _, tx := range transactions {
		idsStr = append(idsStr, "'"+strconv.FormatInt(tx.GetID(), 10)+"'")
		ids = append(ids, tx.GetID())
	}
	err := mps.QueryExecutor.ExecuteTransaction(mps.MempoolQuery.DeleteMempoolTransactions(idsStr))
	if err != nil {
		return err
	}
	err = mps.MempoolCacheStorage.RemoveItem(ids)
	if err != nil {
		return err
	}
	mps.Logger.Infof("mempool transaction with IDs = %s deleted", idsStr)
	return nil
}

func (mps *MempoolService) GetTotalMempoolTransactions() (int, error) {
	var (
		err      error
		mempools = make(storage.MempoolMap)
	)
	err = mps.MempoolCacheStorage.GetAllItems(mempools)
	if err != nil {
		return 0, err
	}
	return len(mempools), nil
}

// GetMempoolTransactions fetch transactions from mempool
func (mps *MempoolService) GetMempoolTransactions() (storage.MempoolMap, error) {
	var (
		mempoolCache = make(storage.MempoolMap)
		err          error
	)
	err = mps.MempoolCacheStorage.GetAllItems(mempoolCache)
	if err != nil {
		return nil, err
	}
	return mempoolCache, nil
}

// AddMempoolTransaction validates and insert a transaction into the mempool and also set the BlockHeight as well
func (mps *MempoolService) AddMempoolTransaction(tx *model.Transaction, txBytes []byte) error {
	// check maximum mempool
	if constant.MaxMempoolTransactions > 0 {
		var count, err = mps.GetTotalMempoolTransactions()
		if err != nil {
			return err
		}
		if count >= constant.MaxMempoolTransactions {
			return blocker.NewBlocker(blocker.ValidationErr, "Mempool already full")
		}
	}

	mpTx := &model.MempoolTransaction{
		FeePerByte:              commonUtils.FeePerByteTransaction(tx.GetFee(), txBytes),
		ID:                      tx.GetID(),
		TransactionBytes:        txBytes,
		ArrivalTimestamp:        time.Now().Unix(),
		SenderAccountAddress:    tx.GetSenderAccountAddress(),
		RecipientAccountAddress: tx.GetRecipientAccountAddress(),
	}

	// NOTE: this select is always inside a db transaction because AddMempoolTransaction is always called within a db tx
	var lastBlock model.Block
	err := mps.BlockStateStorage.GetItem(nil, &lastBlock)
	if err != nil {
		return err
	}
	mpTx.BlockHeight = lastBlock.GetHeight()
	insertMempoolQ, insertMempoolArgs := mps.MempoolQuery.InsertMempoolTransaction(mpTx)
	err = mps.QueryExecutor.ExecuteTransaction(insertMempoolQ, insertMempoolArgs...)
	if err != nil {
		return err
	}
	err = mps.MempoolCacheStorage.SetItem(tx.GetID(), storage.MempoolCacheObject{
		Tx:                  *tx,
		ArrivalTimestamp:    time.Now().UTC().Unix(),
		FeePerByte:          mpTx.FeePerByte,
		TransactionByteSize: uint32(len(txBytes)),
		BlockHeight:         mpTx.BlockHeight,
	})
	if err != nil {
		return err
	}
	return nil
}

func (mps *MempoolService) ValidateMempoolTransaction(mpTx *model.Transaction) error {
	var (
		mempoolObj storage.MempoolCacheObject
		tx         model.Transaction
		err        error
		row        *sql.Row
		txType     transaction.TypeAction
	)
	// check duplication in mempool cache
	err = mps.MempoolCacheStorage.GetItem(mpTx.GetID(), &mempoolObj)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, "FailReadingMempoolCache")
	}
	if mpTx.GetID() == mempoolObj.Tx.GetID() {
		return blocker.NewBlocker(blocker.DuplicateMempoolErr, "MempoolDuplicated")
	}

	// check for duplication in transaction table
	transactionQ := mps.TransactionQuery.GetTransaction(mpTx.GetID())
	row, err = mps.QueryExecutor.ExecuteSelectRow(transactionQ, false)
	if err != nil {
		return err
	}

	err = mps.TransactionQuery.Scan(&tx, row)
	if err != nil && err != sql.ErrNoRows {
		return blocker.NewBlocker(blocker.DBErr, err.Error())
	}

	if mpTx.GetID() == tx.GetID() {
		return blocker.NewBlocker(blocker.ValidationErr, "TransactionAlreadyConfirmed")
	}

	txType, err = mps.ActionTypeSwitcher.GetTransactionType(mpTx)
	if err != nil {
		return blocker.NewBlocker(blocker.ValidationErr, err.Error())
	}

	if errVal := mps.TransactionUtil.ValidateTransaction(mpTx, txType, true); errVal != nil {
		return blocker.NewBlocker(blocker.ValidationErr, errVal.Error())
	}

	err = mps.TransactionCoreService.ValidateTransaction(txType, false)
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
func (mps *MempoolService) SelectTransactionsFromMempool(blockTimestamp int64, blockHeight uint32) ([]*model.Transaction, error) {
	mempoolTransactions, err := mps.GetMempoolTransactions()
	if err != nil {
		return nil, err
	}
	var payloadLength int
	selectedTransactions := make([]*model.Transaction, 0)
	selectedMempoolTxs := make([]storage.MempoolCacheObject, 0)
	for _, memObj := range mempoolTransactions {
		if len(selectedTransactions) >= constant.MaxNumberOfTransactionsInBlock {
			break
		}
		transactionLength := int(memObj.TransactionByteSize)
		if payloadLength+transactionLength > constant.MaxPayloadLengthInBlock {
			continue
		}

		// compute transaction expiration time
		txExpirationTime := memObj.Tx.Timestamp + constant.TransactionExpirationOffset
		// compare to millisecond representation of block timestamp
		if blockTimestamp == 0 || blockTimestamp > txExpirationTime {
			continue
		}

		memObj.Tx.Height = blockHeight

		txType, err := mps.ActionTypeSwitcher.GetTransactionType(&memObj.Tx)
		if err != nil {
			return nil, err
		}

		if err := mps.TransactionUtil.ValidateTransaction(&memObj.Tx, txType, true); err != nil {
			continue
		}

		toRemove, err := txType.SkipMempoolTransaction(
			selectedTransactions,
			blockTimestamp,
			blockHeight,
		)
		if err != nil {
			mps.Logger.Errorf("skip mempool err : %v", err)
			continue
		}
		if toRemove {
			continue
		}
		memObjCopy := memObj
		selectedMempoolTxs = append(selectedMempoolTxs, memObjCopy)
		payloadLength += transactionLength
	}
	sortFeePerByteThenTimestampThenID(selectedMempoolTxs)
	for _, mpTx := range selectedMempoolTxs {
		txCopy := mpTx.Tx
		selectedTransactions = append(selectedTransactions, &txCopy)
	}
	return selectedTransactions, nil
}

func (mps *MempoolService) ReceivedTransaction(
	senderPublicKey,
	receivedTxBytes []byte,
	lastBlock *model.Block,
	nodeSecretPhrase string,
) (*model.BatchReceipt, error) {
	var (
		err          error
		receivedTx   *model.Transaction
		batchReceipt *model.BatchReceipt
	)
	batchReceipt, receivedTx, err = mps.ProcessReceivedTransaction(
		senderPublicKey,
		receivedTxBytes,
		lastBlock,
		nodeSecretPhrase,
	)
	if err != nil {
		return nil, err
	}
	// receivedTx == nil when it's a duplicate but we haven't send receipt to this peer for this particular data yet
	if receivedTx == nil {
		return batchReceipt, nil
	}
	err = mps.QueryExecutor.BeginTx()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	txType, err := mps.ActionTypeSwitcher.GetTransactionType(receivedTx)
	if err != nil {
		rollbackErr := mps.QueryExecutor.RollbackTx()
		if rollbackErr != nil {
			mps.Logger.Warnf("rollbackErr:ReceivedTransaction - %v", rollbackErr)
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = mps.TransactionCoreService.ApplyUnconfirmedTransaction(txType)
	if err != nil {
		mps.Logger.Infof("fail ApplyUnconfirmed tx: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Store to Mempool Transaction
	if err = mps.AddMempoolTransaction(receivedTx, receivedTxBytes); err != nil {
		mps.Logger.Infof("error AddMempoolTransaction: %v\n", err)
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = mps.QueryExecutor.CommitTx()
	if err != nil {
		mps.Logger.Warnf("error committing db transaction: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	mps.Observer.Notify(observer.TransactionAdded, receivedTxBytes, mps.Chaintype)
	return batchReceipt, nil
}

// ProcessReceivedTransaction process the transaction received from peer
// will return batchReceipt, `nil`, `nil` if duplicate transaction found
func (mps *MempoolService) ProcessReceivedTransaction(
	senderPublicKey,
	receivedTxBytes []byte,
	lastBlock *model.Block,
	nodeSecretPhrase string,
) (*model.BatchReceipt, *model.Transaction, error) {
	var (
		err         error
		receivedTx  *model.Transaction
		duplicateTx bool
	)
	receivedTx, err = mps.TransactionUtil.ParseTransactionBytes(receivedTxBytes, true)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}
	receivedTxHash := sha3.Sum256(receivedTxBytes)
	receiptKey, err := mps.ReceiptUtil.GetReceiptKey(
		receivedTxHash[:], senderPublicKey,
	)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	// Validate received transaction
	if err = mps.ValidateMempoolTransaction(receivedTx); err != nil {
		specificErr := err.(blocker.Blocker)
		if specificErr.Type != blocker.DuplicateMempoolErr {
			return nil, nil, status.Error(codes.InvalidArgument, err.Error())
		}

		// already exist in mempool, check if already generated a receipt for this sender
		duplicated, duplicatedErr := mps.ReceiptService.IsDuplicated(senderPublicKey, receivedTxHash[:])
		if duplicatedErr != nil {
			return nil, nil, status.Errorf(codes.Internal, err.Error())
		}
		if duplicated {
			return nil, nil, status.Errorf(codes.Aborted, "ReceiptAlreadyExists")
		}
		duplicateTx = true
	}

	batchReceipt, err := mps.ReceiptService.GenerateBatchReceiptWithReminder(
		mps.Chaintype,
		receivedTxHash[:],
		lastBlock,
		senderPublicKey,
		nodeSecretPhrase,
		string(receiptKey),
		constant.ReceiptDatumTypeTransaction,
	)

	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}
	if duplicateTx {
		return batchReceipt, nil, nil
	}
	return batchReceipt, receivedTx, nil
}

// ReceivedBlockTransactions
func (mps *MempoolService) ReceivedBlockTransactions(
	senderPublicKey []byte,
	receivedTxBytes [][]byte,
	lastBlock *model.Block,
	nodeSecretPhrase string,
) ([]*model.BatchReceipt, error) {
	var (
		batchReceiptArray    []*model.BatchReceipt
		receivedTransactions []*model.Transaction
	)
	for _, txBytes := range receivedTxBytes {
		batchReceipt, receivedTx, err := mps.ProcessReceivedTransaction(senderPublicKey, txBytes, lastBlock, nodeSecretPhrase)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		batchReceiptArray = append(batchReceiptArray, batchReceipt)
		if receivedTx == nil {
			continue
		}
		receivedTransactions = append(receivedTransactions, receivedTx)
	}

	go mps.Observer.Notify(observer.ReceivedBlockTransactionsValidated, receivedTransactions, mps.Chaintype)

	return batchReceiptArray, nil
}

// sortFeePerByteThenTimestampThenID sort a slice of mpTx by feePerByte, timestamp, id DESC
// this sort the transaction by the mempool fields, mean both slice should have the same number of elements, and same
// order for this to work
func sortFeePerByteThenTimestampThenID(memTxs []storage.MempoolCacheObject) {
	sort.SliceStable(memTxs, func(i, j int) bool {
		mi, mj := memTxs[i], memTxs[j]
		switch {
		case mi.FeePerByte != mj.FeePerByte:
			return mi.FeePerByte > mj.FeePerByte
		case mi.ArrivalTimestamp != mj.ArrivalTimestamp:
			return mi.ArrivalTimestamp < mj.ArrivalTimestamp
		default:
			return mi.Tx.ID < mj.Tx.ID
		}
	})
}

// DeleteExpiredMempoolTransactions handle fresh clean the mempool
// which is the mempool transaction has been hit expiration time
func (mps *MempoolService) DeleteExpiredMempoolTransactions() error {
	var (
		qStr              string
		expirationTime    = time.Now().Add(-constant.MempoolExpiration).Unix()
		err               error
		cachedTxs         = make(storage.MempoolMap)
		expiredMempoolIDs []int64
	)
	err = mps.MempoolCacheStorage.GetAllItems(cachedTxs)
	if err != nil {
		return err
	}
	if len(cachedTxs) == 0 {
		return nil
	}
	err = mps.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}
	for _, memObj := range cachedTxs {
		if memObj.ArrivalTimestamp > expirationTime {
			continue
		}
		tx := memObj.Tx
		action, err := mps.ActionTypeSwitcher.GetTransactionType(&tx)
		if err != nil {
			if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
				mps.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		err = mps.TransactionCoreService.UndoApplyUnconfirmedTransaction(action)
		if err != nil {
			if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
				mps.Logger.Error(rollbackErr.Error())
			}
			return err
		}
		expiredMempoolIDs = append(expiredMempoolIDs, memObj.Tx.ID)
	}

	qStr = mps.MempoolQuery.DeleteExpiredMempoolTransactions(expirationTime)
	err = mps.QueryExecutor.ExecuteTransaction(qStr)
	if err != nil {
		if rollbackErr := mps.QueryExecutor.RollbackTx(); rollbackErr != nil {
			mps.Logger.Error(rollbackErr.Error())
		}
		return err
	}
	err = mps.MempoolCacheStorage.RemoveItem(expiredMempoolIDs)
	if err != nil {
		initMempoolErr := mps.InitMempoolTransaction()
		if initMempoolErr != nil {
			mps.Logger.Warnf("BackupMempoolsErr - InitMempoolErr - %v", initMempoolErr)
		}
	}
	err = mps.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}
	return nil
}

func (mps *MempoolService) GetMempoolTransactionsWantToBackup(height uint32) ([]*model.Transaction, error) {
	var (
		txs = make([]*model.Transaction, 0)
		err error
	)

	mempoolMap, err := mps.GetMempoolTransactions()
	if err != nil {
		return txs, err
	}
	for _, memObj := range mempoolMap {
		if memObj.BlockHeight > height {
			txs = append(txs, &memObj.Tx)
		}
	}

	return txs, nil
}

func (mps *MempoolService) BackupMempools(commonBlock *model.Block) error {

	var (
		mempoolsBackup    []*model.Transaction
		mempoolsBackupIDs []int64
		err               error
		backupMempools    = make(map[int64][]byte)
	)

	mempoolsBackup, err = mps.GetMempoolTransactionsWantToBackup(commonBlock.Height)
	if err != nil {
		return err
	}
	mps.Logger.Warnf("mempool tx backup %d in total with block_height %d", len(mempoolsBackup), commonBlock.GetHeight())
	derivedQueries := query.GetDerivedQuery(mps.Chaintype)
	err = mps.QueryExecutor.BeginTx()
	if err != nil {
		return err
	}

	for _, mempoolTx := range mempoolsBackup {
		var (
			txType transaction.TypeAction
		)
		txType, err = mps.ActionTypeSwitcher.GetTransactionType(mempoolTx)
		if err != nil {
			rollbackErr := mps.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				mps.Logger.Warnf("rollbackErr:BackupMempools - %v", rollbackErr)
			}
			return err
		}

		err = mps.TransactionCoreService.UndoApplyUnconfirmedTransaction(txType)
		if err != nil {
			rollbackErr := mps.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				mps.Logger.Warnf("rollbackErr:BackupMempools - %v", rollbackErr)
			}
			return err
		}

		mempoolByte, err := mps.TransactionUtil.GetTransactionBytes(mempoolTx, true)
		if err != nil {
			rollbackErr := mps.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				mps.Logger.Warnf("rollbackErr:BackupMempools - %v", rollbackErr)
			}
			return err
		}

		mempoolsBackupIDs = append(mempoolsBackupIDs, mempoolTx.GetID())
		backupMempools[mempoolTx.GetID()] = mempoolByte
	}

	for _, dQuery := range derivedQueries {
		queries := dQuery.Rollback(commonBlock.Height)
		err = mps.QueryExecutor.ExecuteTransactions(queries)
		if err != nil {
			rollbackErr := mps.QueryExecutor.RollbackTx()
			if rollbackErr != nil {
				mps.Logger.Warnf("rollbackErr:BackupMempools - %v", rollbackErr)
			}
			return err
		}
	}
	err = mps.MempoolCacheStorage.RemoveItem(mempoolsBackupIDs)
	if err != nil {
		initMempoolErr := mps.InitMempoolTransaction()
		if initMempoolErr != nil {
			mps.Logger.Warnf("BackupMempoolsErr - InitMempoolErr - %v", initMempoolErr)
		}
	}
	err = mps.QueryExecutor.CommitTx()
	if err != nil {
		return err
	}

	err = mps.MempoolBackupStorage.SetItems(backupMempools)
	if err != nil {
		return err
	}

	return nil
}
