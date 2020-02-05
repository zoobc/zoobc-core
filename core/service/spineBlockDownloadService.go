package service

type (
	SpineBlockDownloadServiceInterface interface {
		SetSpineBlocksDownloadFinished(isSpineBlocksDownloadFinished bool)
		IsSpineBlocksDownloadFinished() bool
	}

	SpineBlockDownloadService struct {
		isSpineBlocksDownloadFinished bool
	}
)

func NewSpineBlockDownloadService() *SpineBlockDownloadService {
	return &SpineBlockDownloadService{
		isSpineBlocksDownloadFinished: false,
	}
}

func (sbds *SpineBlockDownloadService) SetSpineBlocksDownloadFinished(isSpineBlocksDownloadFinished bool) {
	sbds.isSpineBlocksDownloadFinished = isSpineBlocksDownloadFinished
}

func (sbds *SpineBlockDownloadService) IsSpineBlocksDownloadFinished() bool {
	return sbds.isSpineBlocksDownloadFinished
}
