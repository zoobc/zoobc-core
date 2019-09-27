package constant

import "time"

const (
	// MempoolExpiration time in seconds
	MempoolExpiration = 5 * time.Minute
	// CheckMempoolExpiration time in seconds
	CheckMempoolExpiration = 60 * time.Minute
)
