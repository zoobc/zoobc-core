package constant

import "time"

var (
	MaxNumberOfTransactionsInBlock       = 330
	MinTransactionSizeInBlock            = 176
	MaxPayloadLengthInBlock              = MinTransactionSizeInBlock * MaxNumberOfTransactionsInBlock
	TransactionExpirationOffset    int64 = 3600 // 3600 seconds
	// OneFeePerByteTransaction use to level up accuracy fee per byte of transaction bytes
	// Will be useful when ordering tx in mempool based on fee per byte
	OneFeePerByteTransaction int64 = 10000
	// TransactionTimeOffset use to put time offset for transaction timestamp when validate transaction
	TransactionTimeOffset = 10 * time.Second
	CompleteMinutesUnit   = 60 // 60 seconds
)
