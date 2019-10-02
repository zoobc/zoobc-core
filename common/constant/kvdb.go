package constant

var (
	TableNodeReceiptPrefix            = "node_receipt_"
	TablePublishedReceiptTransaction  = "published_receipt_transaction_"
	TablePublishedReceiptBlock        = "published_receipt_block_"
	TableBlockReminderKey             = "block_reminder"
	ExpiryPublishedReceiptTransaction = 1440 // expiration represented in number of minutes
	ExpiryPublishedReceiptBlock       = 1440
	ExpiryBlockReminder               = 43200 // one month adjust later
)
