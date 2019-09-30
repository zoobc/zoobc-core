package constant

import "time"

const (
	// MempoolExpiration time in Minutes
	MempoolExpiration = 5 * time.Minute
	// CheckMempoolExpiration time in Minutes
	CheckMempoolExpiration = 60 * time.Minute
)
