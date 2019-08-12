package constant

import "time"

const (
	// GetMoreBlocksDelay returns delay between GetMoreBlocksThread in seconds
	GetMoreBlocksDelay               time.Duration = 10
	BlockDownloadSegSize             uint32        = 36
	MaxResponseTime                  time.Duration = 1 * time.Minute
	DefaultNumberOfForkConfirmations int32         = 1
	PeerGetBlocksLimit               uint32        = 1440
	CommonMilestoneBlockIdsLimit     int32         = 10
	SafeBlockGap                     uint32        = 1440
)
