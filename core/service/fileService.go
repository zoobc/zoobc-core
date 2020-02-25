package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
	"hash"
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	FileServiceInterface interface {
		SaveBytesToFile(fileBasePath, filename string, b []byte) error
		GetFileNameFromHash(fileHash []byte) (string, error)
		GetHashFromFileName(fileName string) ([]byte, error)
		VerifyFileHash(filePath string, hash []byte) (bool, error)
		HashPayload(b []byte) []byte
		EncodePayload(v interface{}) (b []byte, err error)
		DecodePayload(b []byte, v interface{}) error
		GetEncoderHandler() codec.Handle
	}

	FileService struct {
		Logger *log.Logger
		h      codec.Handle
		hasher hash.Hash
	}
)

func NewFileService(
	logger *log.Logger,
	encoderHandler codec.Handle,
	fileHasher hash.Hash,
) FileServiceInterface {
	return &FileService{
		Logger: logger,
		h:      encoderHandler, // this variable is only set when constructing the service and never mutated
		hasher: fileHasher,
	}
}

func (fs *FileService) VerifyFileHash(filePath string, hash []byte) (bool, error) {
	return util.VerifyFileHash(filePath, hash, fs.hasher)
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

func (fs *FileService) HashPayload(b []byte) []byte {
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
