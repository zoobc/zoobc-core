package blockers

import "errors"

var (
	ErrBlockInvalidSignature         = errors.New("ErrBlockInvalidSignature")
	ErrBlockNoSignature              = errors.New("ErrBlockNoSignature")
	ErrBlockInvalidPreviousBlockHash = errors.New("ErrBlockInvalidPreviousBlockHash")
	ErrBlockDuplicateBlock           = errors.New("ErrBlockDuplicateBlock")
	ErrBlockInvalidTimestamp         = errors.New("ErrBlockInvalidTimestamp")
)
