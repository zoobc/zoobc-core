package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"os"
)

type (
	FileServiceInterface interface {
		SaveBytesToFile(filePath string, b []byte) (*os.File, error)
		GetFileNameFromHash(fileHash []byte) (string, error)
		GetHashFromFileName(fileName string) ([]byte, error)
		HashPayload(b []byte) []byte
	}

	FileService struct {
		Logger *log.Logger
	}
)

// GetAccountBalances get account balances for snapshot (wrapper function around account balance query)
func (sfs *FileService) SaveBytesToFile(filePath string, b []byte) (*os.File, error) {
	// TODO: implement real method
	return &os.File{}, nil
}

// GetAccountBalances get account balances for snapshot (wrapper function around account balance query)
func (sfs *FileService) HashPayload(b []byte) []byte {
	h := sha3.Sum256(b)
	return h[:]
}

// GetHashFromFileName file name to hash conversion
// TODO: refactor GetPublicKeyFromAddress name as it can be applied to other use cases, such as this one
func (*FileService) GetHashFromFileName(fileName string) ([]byte, error) {
	hash, err := util.GetPublicKeyFromAddress(fileName)
	if err != nil {
		return nil, blocker.NewBlocker(
			blocker.AppErr,
			"invalid file name",
		)
	}
	return hash, nil
}

// GetFileNameFromHash file hash to fileName conversion
// TODO: refactor GetAddressFromPublicKey name as it can be applied to other use cases, such as this one
func (*FileService) GetFileNameFromHash(fileHash []byte) (string, error) {
	fileName, err := util.GetAddressFromPublicKey(fileHash)
	if err != nil {
		return "", blocker.NewBlocker(
			blocker.ServerError,
			"invalid file hash length",
		)
	}
	return fileName, nil
}
