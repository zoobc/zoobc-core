package constant

import "time"

var (
	KVdbExpiryReceiptReminder       = 43200 // one month adjust later
	KVdbTableBlockReminderKey       = "block_reminder_"
	KVdbTableTransactionReminderKey = "transaction_reminder_"
	KVDBMempoolsBackup              = "mempools_backup"
	KVDBMempoolsBackupExpiry        = 60 * time.Minute
)
