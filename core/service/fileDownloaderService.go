package service

import (
	log "github.com/sirupsen/logrus"
)

type (
	// FileDownloaderServiceInterface snapshot logic shared across block types
	FileDownloaderServiceInterface interface {
		DownloadFileByName(fileName string, fileHash []byte) error
	}

	FileDownloaderService struct {
		DownloadPath string
		FileService  FileServiceInterface
		Logger       *log.Logger
	}
)

func NewFileDownloaderService(
	downloadPath string,
	fileService FileServiceInterface,
	logger *log.Logger,
) *FileDownloaderService {
	return &FileDownloaderService{
		DownloadPath: downloadPath,
		FileService:  fileService,
		Logger:       logger,
	}
}

// DownloadSnapshotChunk TODO: implement logic to download a file from a random peer
func (fds *FileDownloaderService) DownloadFileByName(fileName string, fileHash []byte) error {
	// TODO: download file from a peer
	// FIXME uncomment once file download has been fully implemented
	// filePath := filepath.Join(fds.DownloadPath, fileName)
	// ok, err := util.VerifyFileHash(filePath, fileHash, sha3.New256())
	// if err != nil {
	// 	return err
	// }
	// if !ok {
	// 	return blocker.NewBlocker(
	// 	blocker.AppErr,
	// 	"CorruptedFile",
	// )
	// }
	return nil
}
