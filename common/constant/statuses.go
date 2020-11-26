package constant

var (
	BlockchainStatusIdle            = 1
	BlockchainStatusGeneratingBlock = 2
	BlockchainStatusReceivingBlock  = 3
	BlockchainStatusSyncingBlock    = 4
	// BlockchainSendingBlockTransactions needs blockchain lock because transactions are tightly coupled of what blocks the node has
	BlockchainSendingBlockTransactions                  = 5
	BlockchainSendingBlocks                             = 6
	BlockchainStatusReceivingBlockScanBlockPool         = 7
	BlockchainStatusReceivingBlockProcessCompletedBlock = 8
	BlockchainStatusGettingBlocks                       = 9
)
