package constant

import "time"

const (
	// MempoolExpiration time in Minutes
	MempoolExpiration = 60 * time.Minute
	// CheckMempoolExpiration time in Minutes
	CheckMempoolExpiration = 5 * time.Minute
	// Consideration, max mempool tx should equal or greater than MaxNumberOfTransactions
	// Or just leave it 0 for unlimited mempool transaction
	MaxMempoolTransactions = 0
)
