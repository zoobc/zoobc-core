package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/blocker"
	"golang.org/x/crypto/sha3"
)

type (
	FileServiceInterface interface {
		GetDownloadPath() string
		ParseFileChunkHashes(fileHashes []byte, hashLength int) (fileHashesAry [][]byte, err error)
		ReadFileByHash(filePath string, fileHash []byte) ([]byte, error)
		ReadFileByName(filePath, fileName string) ([]byte, error)
		DeleteFilesByHash(filePath string, fileHashes [][]byte) error
		SaveBytesToFile(fileBasePath, filename string, b []byte) error
		GetFileNameFromHash(fileHash []byte) string
		GetFileNameFromBytes(fileBytes []byte) string
		GetHashFromFileName(fileName string) ([]byte, error)
		VerifyFileChecksum(fileBytes, hash []byte) bool
		HashPayload(b []byte) ([]byte, error)
		EncodePayload(v interface{}) (b []byte, err error)
		DecodePayload(b []byte, v interface{}) error
		GetEncoderHandler() codec.Handle
	}

	FileService struct {
		Logger       *log.Logger
		h            codec.Handle
		snapshotPath string
	}
)

func NewFileService(
	logger *log.Logger,
	encoderHandler codec.Handle,
	snapshotPath string,
) FileServiceInterface {
	return &FileService{
		Logger:       logger,
		h:            encoderHandler, // this variable is only set when constructing the service and never mutated
		snapshotPath: snapshotPath,
	}
}

func (fs *FileService) GetDownloadPath() string {
	return fs.snapshotPath
}

func (fs *FileService) ParseFileChunkHashes(fileHashes []byte, hashLength int) (fileHashesAry [][]byte, err error) {
	// math.Mod returns the reminder of len(fileHashes)/hashLength
	// we use it to check if the length of fileHashes is a multiple of the single hash's length (32 bytes for sha256)
	if len(fileHashes) < hashLength || math.Mod(float64(len(fileHashes)), float64(hashLength)) > 0 {
		return nil, blocker.NewBlocker(blocker.ValidationErr, "invalid file chunks hashes length")
	}
	for i := 0; i < len(fileHashes); i += hashLength {
		fileHashesAry = append(fileHashesAry, fileHashes[i:i+hashLength])
	}
	return fileHashesAry, nil
}

func (fs *FileService) VerifyFileChecksum(fileBytes, hash []byte) bool {
	computed := sha3.Sum256(fileBytes)
	return bytes.Equal(computed[:], hash)
}

func (fs *FileService) ReadFileByHash(filePath string, fileHash []byte) ([]byte, error) {
	return fs.ReadFileByName(filePath, fs.GetFileNameFromHash(fileHash))
}

func (fs *FileService) ReadFileByName(filePath, fileName string) ([]byte, error) {
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

// GetHashFromFileName file hash to hash-name conversion: base64 urlencoded
func (*FileService) GetHashFromFileName(fileName string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(fileName)
}

// GetFileNameFromHash file hash to fileName conversion: base64 urlencoded
func (*FileService) GetFileNameFromHash(fileHash []byte) string {
	return base64.URLEncoding.EncodeToString(fileHash)
}

// GetFileNameFromBytes helper method to get a hash-name from file raw bytes
func (fs *FileService) GetFileNameFromBytes(fileBytes []byte) string {
	fileHash := sha3.Sum256(fileBytes)
	return fs.GetFileNameFromHash(fileHash[:])
}

// DeleteFilesByHash remove a list of files by their hash/names
func (fs *FileService) DeleteFilesByHash(filePath string, fileHashes [][]byte) error {
	for _, fileChunkHash := range fileHashes {
		filePathName := filepath.Join(filePath, fs.GetFileNameFromHash(fileChunkHash))
		if err := os.Remove(filePathName); err != nil {
			return err
		}
	}
	return nil
}
