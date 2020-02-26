package service

type (
	SpineBlockStatusService struct {
		isFirstDownloadFinished bool
		isDownloading           bool
	}
)

func NewSpineBlockStatusService() *SpineBlockStatusService {
	return &SpineBlockStatusService{
		isFirstDownloadFinished: false,
		isDownloading:           false,
	}
}

func (sbds *SpineBlockStatusService) SetFirstDownloadFinished(finished bool) {
	sbds.isFirstDownloadFinished = finished
}

func (sbds *SpineBlockStatusService) IsFirstDownloadFinished() bool {
	return sbds.isFirstDownloadFinished
}

func (sbds *SpineBlockStatusService) SetIsDownloading(newValue bool) {
	sbds.isDownloading = newValue
}

func (sbds *SpineBlockStatusService) IsDownloading() bool {
	return sbds.isDownloading
}
