package constant

var (
	MaxNumberOfTransactions           = 255
	MinTransactionSize                = 176
	MaxPayloadLength                  = MinTransactionSize * MaxNumberOfTransactions
	TransactionExpirationOffset int64 = 3600 // 3600 seconds
	TxFeePerByte                int32
	SignatureTypeDefault        = uint32(0)
)
