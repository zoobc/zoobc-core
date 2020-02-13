package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// FileDownloaderServiceInterface snapshot logic shared across block types
	FileDownloaderServiceInterface interface {
		DownloadFileByName(fileName string, fileHash []byte) error
		GetFileNameFromHash(fileHash []byte) (string, error)
		GetHashFromFileName(fileName string) ([]byte, error)
	}

	FileDownloaderService struct {
		DownloadPath string
		Logger       *log.Logger
	}
)

func NewFileDownloaderService(
	downloadPath string,
	logger *log.Logger,
) *FileDownloaderService {
	return &FileDownloaderService{
		DownloadPath: downloadPath,
		Logger:       logger,
	}
}

// GetFileNameFromHash file hash to fileName conversion
// TODO: refactor GetAddressFromPublicKey name as it can be applied to other use cases, such as this one
func (*FileDownloaderService) GetFileNameFromHash(fileHash []byte) (string, error) {
	fileName, err := util.GetAddressFromPublicKey(fileHash)
	if err != nil {
		return "", blocker.NewBlocker(
			blocker.ServerError,
			"invalid file hash length",
		)
	}
	return fileName, nil
}

// GetHashFromFileName file name to hash conversion
// TODO: refactor GetPublicKeyFromAddress name as it can be applied to other use cases, such as this one
func (*FileDownloaderService) GetHashFromFileName(fileName string) ([]byte, error) {
	hash, err := util.GetPublicKeyFromAddress(fileName)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.AppErr,
			"invalid file name",
		)
	}
	return hash, nil
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
