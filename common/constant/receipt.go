package constant

import "time"

const (
	ReceiptDatumTypeBlock       = uint32(1)
	ReceiptDatumTypeTransaction = uint32(2)
	ReceiptBatchMaximum         = uint32(256)
	ReceiptNodeMaximum          = uint32(256)
	PruningChunkedSize          = 500
	// this multiplier is used to expand the receipt selection windows, this avoid multiple database read
	ReceiptBatchPickMultiplier      = uint32(5)
	ReceiptHashSize                 = 32 // sha256
	ReceiptGenerateMarkleRootPeriod = 20 * time.Second
	// cache
	// MaxReceiptBatchCacheRound
	MaxReceiptBatchCacheRound = 40
)
