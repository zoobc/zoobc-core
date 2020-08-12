package helper

import (
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// ParseBytesArgument to parse argument bytes in string into real byte format
func ParseBytesArgument(argsBytesString, separated string) ([]byte, error) {
	var (
		parsedByte    []byte
		byteCharSlice = strings.Split(argsBytesString, separated)
	)
	for _, v := range byteCharSlice {
		byteValue, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		parsedByte = append(parsedByte, byte(byteValue))
	}
	return parsedByte, nil
}

func GetAbsDBPath() string {
	wd, _ := os.Getwd()
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if strings.Contains(dir, "exe") {
		// running via build
		if strings.Contains(wd, "zoobc-core/") {
			return path.Join(wd, "../")
		}
		return wd
	}
	// running as binary
	return wd
}
