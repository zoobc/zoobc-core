type(
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
)

func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
    if constant.MaxMempoolTransactions > 0 {
        ...
        sqlStr := mps.MempoolQuery.GetMempoolTransactions()
        ...
    }
    ...
}

func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
    ...
    // check if already in db
    mempool, err := mps.GetMempoolTransaction(mpTx.ID)
    ...
}

func (mps *MempoolService) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
    ...
    // check for duplication in mempool table
    mempoolQ := mps.MempoolQuery.GetMempoolTransaction()
    ...
}

func (mps *MempoolService) ReceivedTransaction(
    senderPublicKey,
    receivedTxBytes []byte,
    lastBlock *model.Block,
    nodeSecretPhrase string,
) (*model.BatchReceipt, error) {
    ...
    if err = mps.ValidateMempoolTransaction(mempoolTx); err != nil {}
    ...
}

func (mps *MempoolService) ProcessTransactionBytesToMempool(mempoolTx *model.MempoolTransaction) error {
    ...
    if err = mps.ValidateMempoolTransaction(mempoolTx); err != nil {}
    if err = mps.AddMempoolTransaction(mempoolTx); err != nil {}
    ...
}




























// AddMempoolTransaction validates and insert a transaction into the mempool and also set the BlockHeight as well
func (mps *MempoolService) AddMempoolTransaction(mpTx *model.MempoolTransaction, blockHeight uint32) error {
    // check if the max mempool constant not 0
        // validate if the number of mempool transactions is not more than constant
    
    // check if the mempool TX is aleady exist in the DB

    // put the block height to the mempool TX object

    // add the mempool to the DB
}

func (mps *MempoolService) ProcessTransactionBytesToMempool(mempoolTx *model.MempoolTransaction) error {
    // validate

    // begin dbTX

    // tx apply unconfirmed
        // x -> rollbackTx & return error

    // AddMempoolTransaction
        // x -> rollbackTx & return error
    
    // commitTx
}
