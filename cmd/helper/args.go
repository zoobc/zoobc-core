package helper

import (
	"os"
	"path"
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
	if strings.Contains(wd, "zoobc-core/") {
		return path.Join(wd, "../")
	}
	return wd
}
