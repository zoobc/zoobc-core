package service

type (
	MainBlockStatusService struct {
		isFirstDownloadFinished bool
	}
)

func NewMainBlockStatusService() *MainBlockStatusService {
	return &MainBlockStatusService{
		isFirstDownloadFinished: false,
	}
}

func (sbds *MainBlockStatusService) SetFirstDownloadFinished(finished bool) {
	sbds.isFirstDownloadFinished = finished
}

func (sbds *MainBlockStatusService) IsFirstDownloadFinished() bool {
	return sbds.isFirstDownloadFinished
}
