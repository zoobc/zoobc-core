package service

type (
	BlockStatusServiceInterface interface {
		SetFirstDownloadFinished(isSpineBlocksDownloadFinished bool)
		IsFirstDownloadFinished() bool
	}
)
