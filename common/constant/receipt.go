package constant

const (
	ReceiptDatumTypeBlock       = uint32(1)
	ReceiptDatumTypeTransaction = uint32(2)
	ReceiptBatchMaximum         = uint32(8) // 256 in production
	ReceiptNumberToPick         = 20
	ReceiptNumberOfBlockToPick  = 1000
	ReceiptHashSize             = 32 // sha256
)
