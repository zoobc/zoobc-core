package helper

import (
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
