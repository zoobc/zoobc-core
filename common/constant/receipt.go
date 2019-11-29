package constant

import "time"

const (
	ReceiptDatumTypeBlock        = uint32(1)
	ReceiptDatumTypeTransaction  = uint32(2)
	ReceiptBatchMaximum          = uint32(8) // 256 in production
	ReceiptNodeMaximum           = uint32(8) // 256 in production
	ReceiptNumberToPick          = 20
	NodeReceiptExpiryBlockHeight = 1000
	PruningChunkedSize           = 500
	PruningNodeReceiptPeriod     = 5 * time.Second
	// this multiplier is used to expand the receipt selection windows, this avoid multiple database read
	ReceiptBatchPickMultiplier      = uint32(5)
	ReceiptHashSize                 = 32 // sha256
	ReceiptGenerateMarkleRootPeriod = 20 * time.Second
)
