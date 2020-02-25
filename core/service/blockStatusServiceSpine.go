package service

type (
	SpineBlockStatusService struct {
		isFirstDownloadFinished bool
	}
)

func NewSpineBlockStatusService() *SpineBlockStatusService {
	return &SpineBlockStatusService{
		isFirstDownloadFinished: false,
	}
}

func (sbds *SpineBlockStatusService) SetFirstDownloadFinished(finished bool) {
	sbds.isFirstDownloadFinished = finished
}

func (sbds *SpineBlockStatusService) IsFirstDownloadFinished() bool {
	return sbds.isFirstDownloadFinished
}
