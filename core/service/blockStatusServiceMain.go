package service

type (
	MainBlockStatusService struct {
		isFirstDownloadFinished bool
		isDownloading           bool
	}
)

func NewMainBlockStatusService() *MainBlockStatusService {
	return &MainBlockStatusService{
		isFirstDownloadFinished: false,
		isDownloading:           false,
	}
}

func (sbds *MainBlockStatusService) SetFirstDownloadFinished(finished bool) {
	sbds.isFirstDownloadFinished = finished
}

func (sbds *MainBlockStatusService) IsFirstDownloadFinished() bool {
	return sbds.isFirstDownloadFinished
}

func (sbds *MainBlockStatusService) SetIsDownloading(newValue bool) {
	sbds.isDownloading = newValue
}

func (sbds *MainBlockStatusService) IsDownloading() bool {
	return sbds.isDownloading
}
