package util

import (
	"bytes"
	"errors"
)

func ReadTransactionBytes(buf *bytes.Buffer, nBytes int) ([]byte, error) {
	nextBytes := buf.Next(nBytes)
	if len(nextBytes) < nBytes {
		return nil, errors.New("EndOfBufferReached")
	}
	return nextBytes, nil
}
