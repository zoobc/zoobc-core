package util

import (
	"bytes"
	"errors"
	"hash"
	"io"
	"os"

	"github.com/zoobc/zoobc-core/common/constant"
)

// ReadTransactionBytes get a slice containing the next nBytes from the buffer
func ReadTransactionBytes(buf *bytes.Buffer, nBytes int) ([]byte, error) {
	// TODO: renaming function, this function is not just use for reading bytes of transaction
	nextBytes := buf.Next(nBytes)
	if len(nextBytes) < nBytes {
		return nil, errors.New("EndOfBufferReached")
	}
	return nextBytes, nil
}

// FeePerByteTransaction use to calculate fee of each bytes transaction
func FeePerByteTransaction(feeTransaction int64, transactionBytes []byte) int64 {
	if len(transactionBytes) != 0 {
		return (feeTransaction * constant.OneFeePerByteTransaction) / int64(len(transactionBytes))
	}
	return feeTransaction * constant.OneFeePerByteTransaction
}

func VerifyFileHash(filePath string, hash []byte, hasher hash.Hash) (bool, error) {
	fc, err := ComputeFileHash(filePath, hasher)
	if err != nil {
		return false, err
	}
	if bytes.Equal(fc, hash) {
		return true, nil
	}
	return false, nil
}

func ComputeFileHash(filePath string, hasher hash.Hash) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}
