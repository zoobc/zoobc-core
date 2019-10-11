package util

import (
	"bytes"
	"fmt"
	"strconv"
)

// RenderByteArrayAsString renders a byte array in the form of a string that can easily be copy/pasted into (golang) code, eg. for testing
func RenderByteArrayAsString(bArray []byte) string {
	var buffer bytes.Buffer
	for i, b := range bArray {
		buffer.WriteString(strconv.Itoa(int(b)))
		if i != len(bArray)-1 {
			if i != 0 && i%18 == 0 {
				buffer.WriteString(", \n\t\t\t\t")
			} else {
				buffer.WriteString(", ")
			}
		}
	}
	return fmt.Sprintf("[]byte{%s}", buffer.String())
}
