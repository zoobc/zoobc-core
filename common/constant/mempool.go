package constant

import "time"

const (
	// MempoolExpiration time in Minutes
	MempoolExpiration = 60 * time.Minute
	// CheckMempoolExpiration time in Minutes
	CheckMempoolExpiration = 5 * time.Minute
	// MaxMempoolTransactions is maximum transaction in mempool
	// For consideration, max mempool tx should equal or greater than MaxNumberOfTransactionsInBlock
	// Or just leave it 0 for unlimited mempool transaction
	MaxMempoolTransactions = 100_000
	// Timeout of the transaction candidate in second
	TxCachedTimeout = 300
	// Gap of the CleanTimedoutTxCandidateThread
	CleanTimedoutBlockTxCachedThreadGap = 10
	// MaxMoveMempoolTrasaction is maximum mempool to move from full cache to normal cache
	MempoolMaxMoveTrasactions = 2 * MaxNumberOfTransactionsInBlock
	// MempoolMovePeriod the period to move full cahce mempool into normal cache
	MempoolMoveFullCachePeriod = 4 * time.Second
	// MempoolMaxTimeGapBecameFullCacheTransaction the maximimum time gap between incoming transaction
	// to determine whether the incoming transaction became full cache mempool or not
	MempoolMaxTimeGapBecameFullCacheTransaction = 2
)
