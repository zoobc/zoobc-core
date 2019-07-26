package util

import (
	"regexp"
	"strings"
)

/*
ValidateIP4 validates format of ipv4
*/
func ValidateIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	pattern := `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	re, _ := regexp.Compile(pattern)
	return re.MatchString(ipAddress)
}
