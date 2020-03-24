package util

import (
	"golang.org/x/crypto/sha3"
	"os"
	"path/filepath"
)

func GetExecutableHash() ([]byte, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	filePath, err := filepath.Abs(ex)
	if err != nil {
		return nil, err
	}
	return ComputeFileHash(filePath, sha3.New256())
}
