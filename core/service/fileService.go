package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	FileServiceInterface interface {
		ReadFileByHash(filePath string, fileHash []byte) ([]byte, error)
		DeleteFilesByHash(filePath string, fileHashes [][]byte) error
		SaveBytesToFile(fileBasePath, filename string, b []byte) error
		GetFileNameFromHash(fileHash []byte) (string, error)
		GetHashFromFileName(fileName string) ([]byte, error)
		VerifyFileHash(filePath string, hash []byte) (bool, error)
		HashPayload(b []byte) ([]byte, error)
		EncodePayload(v interface{}) (b []byte, err error)
		DecodePayload(b []byte, v interface{}) error
		GetEncoderHandler() codec.Handle
	}

	FileService struct {
		Logger *log.Logger
		h      codec.Handle
	}
)

func NewFileService(
	logger *log.Logger,
	encoderHandler codec.Handle,
) FileServiceInterface {
	return &FileService{
		Logger: logger,
		h:      encoderHandler, // this variable is only set when constructing the service and never mutated
	}
}

func (fs *FileService) VerifyFileHash(filePath string, hash []byte) (bool, error) {
	return util.VerifyFileHash(filePath, hash, sha3.New256())
}

func (fs *FileService) ReadFileByHash(filePath string, fileHash []byte) ([]byte, error) {
	fileName, err := fs.GetFileNameFromHash(fileHash)
	if err != nil {
		return nil, err
	}
	filePathName := filepath.Join(filePath, fileName)
	chunkBytes, err := ioutil.ReadFile(filePathName)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AppErr,
			fmt.Sprintf("Cannot read file from storage. file : %s Error: %v", filePathName, err))
	}
	return chunkBytes, nil
}

func (fs *FileService) GetEncoderHandler() codec.Handle {
	return fs.h
}

func (fs *FileService) SetEncoder(hh codec.Handle) {
	fs.h = hh
}

// EncodePayload encodes a generic interface (eg. any model) using service's encoder handler (default should be CBOR)
func (fs *FileService) EncodePayload(v interface{}) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, fs.h)
	err = enc.Encode(v)
	return b, err
}

// DecodePayload decodes a byte slice encoded using service's encoder handler (default should be CBOR) into a model.
func (fs *FileService) DecodePayload(b []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(b, fs.h)
	err := dec.Decode(&v)
	return err
}

func (fs *FileService) SaveBytesToFile(fileBasePath, fileName string, b []byte) error {
	// try to create folder if doesn't exist
	if _, err := os.Stat(fileBasePath); os.IsNotExist(err) {
		_ = os.MkdirAll(fileBasePath, os.ModePerm)
	}

	filePath := filepath.Join(fileBasePath, fileName)
	err := ioutil.WriteFile(filePath, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (fs *FileService) HashPayload(b []byte) ([]byte, error) {
	hasher := sha3.New256()
	_, err := hasher.Write(b)
	if err != nil {
		return nil, err
	}
	return hasher.Sum([]byte{}), nil
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

// DeleteFilesByHash remove a list of files by their hash/names
func (fs *FileService) DeleteFilesByHash(filePath string, fileHashes [][]byte) error {
	for _, fileChunkHash := range fileHashes {
		fileName, err := fs.GetFileNameFromHash(fileChunkHash)
		if err != nil {
			return err
		}
		filePathName := filepath.Join(filePath, fileName)
		if err := os.Remove(filePathName); err != nil {
			return err
		}
	}
	return nil
}
