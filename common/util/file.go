package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
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

// GetRootPath return root of project
func GetRootPath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "./", err
	}
	if strings.Contains(p, "exe") || strings.Contains(p, "private") {
		// running go file
		return os.Getwd()
	}
	// gops executable must be in the path. See https://github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		// our parent process is (probably) the Delve debugger
		return os.Getwd()
	}
	return filepath.Dir(p), nil
}
