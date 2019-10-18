package constant

var (
	MaxNumberOfTransactions           = 255
	MinTransactionSize                = 176
	MaxPayloadLength                  = MinTransactionSize * MaxNumberOfTransactions
	TransactionExpirationOffset int64 = 3600 // 3600 seconds
	SignatureTypeDefault              = uint32(0)
	// OneFeePerByteTransaction use to level up accuracy fee per byte of transaction bytes
	// Will be usefull when ordering tx in mempool based on fee per byte
	OneFeePerByteTransaction int64 = 10000
)
