package util

import (
	"bytes"
	"crypto/rand"
	"errors"
	"hash"
	"io/ioutil"

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
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	_, err = hasher.Write(b)
	if err != nil {
		return nil, err
	}
	return hasher.Sum([]byte{}), nil
}

// SplitByteSliceByChunkSize split a byte slice into multiple chunks of equal size,
// beside the last chunk which could be shorter than the others, if the original slice's length is not multiple of chunkSize
func SplitByteSliceByChunkSize(b []byte, chunkSize int) (splitSlice [][]byte) {
	for i := 0; i < len(b); i += chunkSize {
		end := i + chunkSize
		if end > len(b) {
			end = len(b)
		}
		splitSlice = append(splitSlice, b[i:end])
	}
	return splitSlice
}

// GetChecksumByte Calculate a checksum byte from a collection of bytes
// checksum 255 = 255, 256 = 0, 257 = 1 and so on.
func GetChecksumByte(bytes []byte) byte {
	n := len(bytes)
	var a byte
	for i := 0; i < n; i++ {
		a += bytes[i]
	}
	return a
}

// GenerateRandomBytes returns securely generated random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
